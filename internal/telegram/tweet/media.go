package tweet

import (
	"strings"

	"twitterx-bot/internal/twitterxapi"
)

const MaxMediaGroupSize = 10

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
