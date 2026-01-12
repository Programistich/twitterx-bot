package tweet

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"twitterx-bot/internal/twitterxapi"
)

const MaxMediaGroupSize = 10

// MaxVideoFileSize is the maximum file size Telegram bots can send via URL (50 MB).
const MaxVideoFileSize = 50 * 1024 * 1024

// VideoSizeChecker checks video file sizes via HTTP HEAD requests.
type VideoSizeChecker struct {
	Client *http.Client
}

// DefaultVideoSizeChecker returns a checker with default HTTP client.
func DefaultVideoSizeChecker() *VideoSizeChecker {
	return &VideoSizeChecker{
		Client: &http.Client{Timeout: 5 * time.Second},
	}
}

// Check checks if a video URL exceeds the Telegram file size limit.
// Returns true if video is within limits, false if too large or on error.
func (c *VideoSizeChecker) Check(ctx context.Context, url string) bool {
	if url == "" {
		return false
	}

	client := c.Client
	if client == nil {
		client = &http.Client{Timeout: 5 * time.Second}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodHead, url, nil)
	if err != nil {
		return false
	}

	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	contentLength := resp.Header.Get("Content-Length")
	if contentLength == "" {
		// If we can't determine size, assume it's too large
		return false
	}

	size, err := strconv.ParseInt(contentLength, 10, 64)
	if err != nil {
		return false
	}

	return size <= MaxVideoFileSize
}

// CheckVideoSize checks if a video URL exceeds the Telegram file size limit.
// Returns true if video is within limits, false if too large or on error.
func CheckVideoSize(ctx context.Context, url string) bool {
	return DefaultVideoSizeChecker().Check(ctx, url)
}

func SelectPhoto(media *twitterxapi.Media) (url, thumb string, width, height int) {
	if media == nil || len(media.Photos) == 0 {
		return "", "", 0, 0
	}

	if len(media.Photos) > 1 {
		if mosaicURL := PickMosaicURL(media.Mosaic); mosaicURL != "" {
			w, h := MosaicDimensions(media.Mosaic)
			return mosaicURL, mosaicURL, w, h
		}
	}

	photo := media.Photos[0]
	return photo.URL, photo.URL, photo.Width, photo.Height
}

func MediaPreview(media *twitterxapi.Media) (url, kind string) {
	if media == nil {
		return "", ""
	}

	if len(media.Videos) > 0 {
		if u := strings.TrimSpace(media.Videos[0].ThumbnailURL); u != "" {
			return u, "video"
		}
	}

	if len(media.Photos) > 0 {
		if len(media.Photos) > 1 {
			if mosaicURL := PickMosaicURL(media.Mosaic); mosaicURL != "" {
				return mosaicURL, "mosaic"
			}
		}
		if u := strings.TrimSpace(media.Photos[0].URL); u != "" {
			return u, "photo"
		}
	}

	return "", ""
}

func MediaHint(kind string) string {
	switch kind {
	case "video":
		return "Video"
	case "mosaic":
		return "Mosaic"
	case "photo":
		return "Photo"
	default:
		return ""
	}
}

func PickMosaicURL(mosaic *twitterxapi.Mosaic) string {
	if mosaic == nil || len(mosaic.Formats) == 0 {
		return ""
	}
	if url := strings.TrimSpace(mosaic.Formats["jpeg"]); url != "" {
		return url
	}
	if url := strings.TrimSpace(mosaic.Formats["jpg"]); url != "" {
		return url
	}
	return ""
}

func MosaicDimensions(mosaic *twitterxapi.Mosaic) (width, height int) {
	if mosaic == nil {
		return 0, 0
	}
	if mosaic.Width != nil {
		width = *mosaic.Width
	}
	if mosaic.Height != nil {
		height = *mosaic.Height
	}
	return width, height
}

func MimeTypeForVideo(format string) string {
	format = strings.TrimSpace(format)
	if format == "" {
		return "video/mp4"
	}
	if strings.Contains(format, "/") {
		return format
	}
	return "video/" + format
}
