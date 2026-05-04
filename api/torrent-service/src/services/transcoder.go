package services

import (
	"encoding/json"
	"io"
	"log"
	"os/exec"
	"path/filepath"
	"strings"
)

var nativeBrowserExtensions = map[string]bool{
	".mp4":  true,
	".webm": true,
	".ogg":  true,
	".m4v":  true,
}

// CodecInfo holds the detected video and audio codec names from ffprobe.
type CodecInfo struct {
	VideoCodec string
	AudioCodec string
}

// TranscodeJob holds the running ffmpeg process and its output reader.
type TranscodeJob struct {
	Reader      io.ReadCloser
	Cmd         *exec.Cmd
	ContentType string
}

// NeedsTranscoding returns true when the file extension is not natively playable in browsers.
func NeedsTranscoding(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return !nativeBrowserExtensions[ext]
}

// ProbeCodecs runs ffprobe on filePath and returns the first video and audio codec names.
func ProbeCodecs(filePath string) (*CodecInfo, error) {
	out, err := exec.Command(
		"ffprobe",
		"-v", "quiet",
		"-print_format", "json",
		"-show_streams",
		filePath,
	).Output()
	if err != nil {
		return nil, err
	}

	var probe struct {
		Streams []struct {
			CodecType string `json:"codec_type"`
			CodecName string `json:"codec_name"`
		} `json:"streams"`
	}
	if err := json.Unmarshal(out, &probe); err != nil {
		return nil, err
	}

	info := &CodecInfo{}
	for _, s := range probe.Streams {
		switch s.CodecType {
		case "video":
			if info.VideoCodec == "" {
				info.VideoCodec = s.CodecName
			}
		case "audio":
			if info.AudioCodec == "" {
				info.AudioCodec = s.CodecName
			}
		}
	}
	return info, nil
}

// canCopyStream returns true when the source is already H.264 video with AAC/MP3 audio,
// meaning we can remux to fMP4 without re-encoding (much faster).
func canCopyStream(info *CodecInfo) bool {
	if info == nil {
		return false
	}
	videoOK := info.VideoCodec == "h264"
	audioOK := info.AudioCodec == "aac" || info.AudioCodec == "mp3"
	return videoOK && audioOK
}

// StartTranscode spawns ffmpeg reading from reader and writing fragmented MP4 to its stdout.
// When codecInfo indicates H.264+AAC, it remuxes with -c copy; otherwise it re-encodes.
func StartTranscode(reader io.Reader, codecInfo *CodecInfo) (*TranscodeJob, error) {
	args := buildFFmpegArgs(codecInfo)
	cmd := exec.Command("ffmpeg", args...)
	cmd.Stdin = reader

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	// Log stderr so transcoding errors appear in service logs.
	cmd.Stderr = newPrefixWriter("ffmpeg")

	if err := cmd.Start(); err != nil {
		return nil, err
	}
	return &TranscodeJob{Reader: stdout, Cmd: cmd, ContentType: "video/mp4"}, nil
}

func buildFFmpegArgs(info *CodecInfo) []string {
	base := []string{
		"-i", "pipe:0",
		"-f", "mp4",
		"-movflags", "frag_keyframe+empty_moov+default_base_moof",
		"pipe:1",
	}

	if canCopyStream(info) {
		// Remux only — no re-encoding, very fast.
		return append([]string{"-c", "copy"}, base...)
	}

	// Full transcode to H.264/AAC for maximum browser compatibility.
	return append([]string{
		"-c:v", "libx264",
		"-preset", "ultrafast",
		"-tune", "zerolatency",
		"-c:a", "aac",
		"-b:a", "128k",
	}, base...)
}

// prefixWriter writes log lines prefixed with a label.
type prefixWriter struct{ label string }

func newPrefixWriter(label string) *prefixWriter { return &prefixWriter{label: label} }

func (w *prefixWriter) Write(p []byte) (int, error) {
	log.Printf("[%s] %s", w.label, strings.TrimRight(string(p), "\n"))
	return len(p), nil
}
