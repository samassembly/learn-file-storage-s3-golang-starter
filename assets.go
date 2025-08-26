package main

import (
	"os"
	"os/exec"
	"fmt"
	"bytes"
	"strings"
	"encoding/json"
	"path/filepath"
)

type Stream struct {
	Width int `json:"width"`
	Height int `json:"height"`
}

type FFProbeOutput struct {
    Streams []Stream `json:"streams"`
}

func (cfg apiConfig) ensureAssetsDir() error {
	if _, err := os.Stat(cfg.assetsRoot); os.IsNotExist(err) {
		return os.Mkdir(cfg.assetsRoot, 0755)
	}
	return nil
}

func getAssetPath(randString string, mediaType string) string {
	ext := mediaTypeToExt(mediaType)
	return fmt.Sprintf("%s%s", randString, ext)
}

func (cfg apiConfig) getObjectUrl(key string) string {
	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", cfg.s3Bucket, cfg.s3Region, key)
}

func (cfg apiConfig) getAssetDiskPath(assetPath string) string {
	return filepath.Join(cfg.assetsRoot, assetPath)
}

func (cfg apiConfig) getAssetURL(assetPath string) string {
	return fmt.Sprintf("http://localhost:%s/assets/%s", cfg.port, assetPath)
}

func (cfg apiConfig) getVideoAspectRatio(filePath string) (string, error) {
	cmd := exec.Command("ffprobe", "-v", "error", "-print_format", "json", "-show_streams", filePath)
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("Failed to run ffprobe: %w", err)
	}

	var probeOutput FFProbeOutput
	if err := json.Unmarshal(out.Bytes(), &probeOutput); err != nil {
		return "", fmt.Errorf("failed to parse ffprobe output: %w", err)
	}
	
	
	for _, stream := range probeOutput.Streams {
		if stream.Width > 0 && stream.Height > 0 {
			ratio := float64(stream.Width) / float64(stream.Height)

			switch {
			case approxEqual(ratio, 16.0/9.0):
				return "16:9", nil
			case approxEqual(ratio, 9.0/16.0):
				return "9:16", nil
			default:
				return "other", nil
			}
		}
	}

return "", fmt.Errorf("no valid video stream found")
}

func (cfg apiConfig) classifyOrientation(aspectRatio string) string {
    switch aspectRatio {
    case "16:9":
        return "landscape"
    case "9:16":
        return "portrait"
    case "other":
        return "other"
    default:
        return "unknown"
    }
}

func mediaTypeToExt(mediaType string) string {
	parts := strings.Split(mediaType, "/")
	if len(parts) != 2 {
		return ".bin"
	}
	return "." + parts[1]
}

func approxEqual(a, b float64) bool {
    const tolerance = 0.01
    return (a-b) < tolerance && (b-a) < tolerance
}
