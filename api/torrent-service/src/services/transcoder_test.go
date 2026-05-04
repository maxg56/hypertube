package services

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// makeTestMKV generates a 1-second 64x64 MKV with H.264+AAC via ffmpeg.
// Skips the test if ffmpeg is not in PATH.
func makeTestMKV(t *testing.T) string {
	t.Helper()
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("ffmpeg not in PATH")
	}
	out := filepath.Join(t.TempDir(), "test.mkv")
	cmd := exec.Command("ffmpeg",
		"-f", "lavfi", "-i", "testsrc=duration=1:size=64x64:rate=1",
		"-f", "lavfi", "-i", "sine=frequency=440:duration=1",
		"-c:v", "libx264", "-c:a", "aac",
		"-t", "1", "-y", out,
	)
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "ffmpeg test-video generation failed: %s", output)
	return out
}

// ---- NeedsTranscoding ----

func TestNeedsTranscoding(t *testing.T) {
	tests := []struct {
		filename string
		want     bool
	}{
		{"movie.mkv", true},
		{"movie.avi", true},
		{"movie.mov", true},
		{"movie.MKV", true}, // case-insensitive
		{"movie.AVI", true},
		{"movie.mp4", false},
		{"movie.MP4", false},
		{"movie.webm", false},
		{"movie.ogg", false},
		{"movie.m4v", false},
		{"noextension", true},
		{"", true},
	}
	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			assert.Equal(t, tt.want, NeedsTranscoding(tt.filename))
		})
	}
}

// ---- canCopyStream ----

func TestCanCopyStream(t *testing.T) {
	tests := []struct {
		name string
		info *CodecInfo
		want bool
	}{
		{"nil info", nil, false},
		{"h264+aac", &CodecInfo{VideoCodec: "h264", AudioCodec: "aac"}, true},
		{"h264+mp3", &CodecInfo{VideoCodec: "h264", AudioCodec: "mp3"}, true},
		{"hevc+aac needs transcode", &CodecInfo{VideoCodec: "hevc", AudioCodec: "aac"}, false},
		{"h264+opus needs transcode", &CodecInfo{VideoCodec: "h264", AudioCodec: "opus"}, false},
		{"h264+ac3 needs transcode", &CodecInfo{VideoCodec: "h264", AudioCodec: "ac3"}, false},
		{"vp9+opus needs transcode", &CodecInfo{VideoCodec: "vp9", AudioCodec: "opus"}, false},
		{"empty codecs", &CodecInfo{}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, canCopyStream(tt.info))
		})
	}
}

// ---- buildFFmpegArgs ----

func TestBuildFFmpegArgs_Remux(t *testing.T) {
	args := buildFFmpegArgs(&CodecInfo{VideoCodec: "h264", AudioCodec: "aac"})
	assert.Contains(t, args, "copy")
	assert.Contains(t, args, "frag_keyframe+empty_moov+default_base_moof")
	assert.NotContains(t, args, "libx264")
	assert.NotContains(t, args, "ultrafast")
}

func TestBuildFFmpegArgs_Transcode(t *testing.T) {
	tests := []struct {
		name string
		info *CodecInfo
	}{
		{"nil info", nil},
		{"hevc+aac", &CodecInfo{VideoCodec: "hevc", AudioCodec: "aac"}},
		{"h264+opus", &CodecInfo{VideoCodec: "h264", AudioCodec: "opus"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := buildFFmpegArgs(tt.info)
			assert.Contains(t, args, "libx264")
			assert.Contains(t, args, "ultrafast")
			assert.Contains(t, args, "aac")
			assert.Contains(t, args, "frag_keyframe+empty_moov+default_base_moof")
			assert.NotContains(t, args, "copy")
		})
	}
}

// ---- ProbeCodecs ----

func TestProbeCodecs(t *testing.T) {
	if _, err := exec.LookPath("ffprobe"); err != nil {
		t.Skip("ffprobe not in PATH")
	}
	path := makeTestMKV(t)

	info, err := ProbeCodecs(path)
	require.NoError(t, err)
	assert.Equal(t, "h264", info.VideoCodec)
	assert.Equal(t, "aac", info.AudioCodec)
}

func TestProbeCodecs_NonVideoFile(t *testing.T) {
	if _, err := exec.LookPath("ffprobe"); err != nil {
		t.Skip("ffprobe not in PATH")
	}
	f, err := os.CreateTemp(t.TempDir(), "notavideo*.txt")
	require.NoError(t, err)
	_, _ = f.WriteString("not a video")
	f.Close()

	// ffprobe on a text file should error (non-zero exit)
	_, err = ProbeCodecs(f.Name())
	assert.Error(t, err)
}

func TestProbeCodecs_NonExistentFile(t *testing.T) {
	if _, err := exec.LookPath("ffprobe"); err != nil {
		t.Skip("ffprobe not in PATH")
	}
	_, err := ProbeCodecs("/tmp/this_file_does_not_exist_xyz.mkv")
	assert.Error(t, err)
}

// ---- StartTranscode ----

func TestStartTranscode_FullTranscode(t *testing.T) {
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("ffmpeg not in PATH")
	}
	path := makeTestMKV(t)

	f, err := os.Open(path)
	require.NoError(t, err)
	defer f.Close()

	job, err := StartTranscode(f, nil) // nil → full transcode
	require.NoError(t, err)
	defer job.Cmd.Wait()

	out, err := io.ReadAll(job.Reader)
	job.Reader.Close()

	require.NoError(t, err)
	assert.Equal(t, "video/mp4", job.ContentType)
	// A valid fragmented MP4 has at least a few hundred bytes.
	assert.Greater(t, len(out), 64, "output should contain fragmented MP4 data")
}

func TestStartTranscode_Remux(t *testing.T) {
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("ffmpeg not in PATH")
	}
	path := makeTestMKV(t)

	info, err := ProbeCodecs(path)
	if err != nil || !canCopyStream(info) {
		t.Skip("test MKV does not have h264+aac, skipping remux test")
	}

	f, err := os.Open(path)
	require.NoError(t, err)
	defer f.Close()

	job, err := StartTranscode(f, info) // compatible codecs → remux
	require.NoError(t, err)
	defer job.Cmd.Wait()

	out, err := io.ReadAll(job.Reader)
	job.Reader.Close()

	require.NoError(t, err)
	assert.Equal(t, "video/mp4", job.ContentType)
	assert.Greater(t, len(out), 64)
}

func TestStartTranscode_FFmpegNotFound(t *testing.T) {
	// Temporarily shadow PATH to simulate ffmpeg absence.
	orig := os.Getenv("PATH")
	require.NoError(t, os.Setenv("PATH", t.TempDir()))
	t.Cleanup(func() { os.Setenv("PATH", orig) })

	_, err := StartTranscode(nil, nil)
	assert.Error(t, err)
}
