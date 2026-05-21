package services

import (
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"torrent-service/src/conf"
	"torrent-service/src/models"
)

// ReencodeToHEVC re-encodes a completed film to H.265/HEVC in the background.
// The original file is removed only after a successful re-encode.
// Safe on Linux: existing open file descriptors survive an unlink.
func ReencodeToHEVC(record *models.TorrentRecord) {
	if record.Reencoded || record.FilePath == "" {
		return
	}

	codecInfo, err := ProbeCodecs(record.FilePath)
	if err == nil && codecInfo.VideoCodec == "hevc" {
		conf.DB.Model(record).Update("reencoded", true)
		return
	}

	ext := filepath.Ext(record.FilePath)
	outputPath := strings.TrimSuffix(record.FilePath, ext) + ".hevc.mp4"

	cmd := exec.Command("ffmpeg",
		"-i", record.FilePath,
		"-c:v", "libx265",
		"-preset", hevcPreset(),
		"-crf", "28",
		"-c:a", "aac",
		"-b:a", "128k",
		"-movflags", "+faststart",
		"-y",
		outputPath,
	)
	cmd.Stderr = io.Discard

	if err := cmd.Run(); err != nil {
		log.Printf("[reencoder] failed for %s: %v", record.InfoHash, err)
		_ = os.Remove(outputPath)
		return
	}

	originalPath := record.FilePath

	fi, err := os.Stat(outputPath)
	newSize := record.FileSize
	if err == nil {
		newSize = fi.Size()
	}

	// Update DB first so new streams use the re-encoded file.
	conf.DB.Model(record).Updates(map[string]any{
		"file_path": outputPath,
		"file_size": newSize,
		"reencoded": true,
	})

	// Unlink the original — existing open fds remain valid on Linux.
	if err := os.Remove(originalPath); err != nil {
		log.Printf("[reencoder] could not remove original %s: %v", originalPath, err)
	}

	saved := float64(record.FileSize-newSize) / 1024 / 1024
	log.Printf("[reencoder] done for %s: %.1f MB saved", record.InfoHash, saved)

	// Now that this version is the smallest, clean up any other quality variants.
	purgeRedundantQualities()
}

func hevcPreset() string {
	if p := os.Getenv("HEVC_PRESET"); p != "" {
		return p
	}
	return "medium"
}
