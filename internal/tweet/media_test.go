package tweet

import (
	"testing"

	"twitterx-bot/internal/twitterxapi"
)

func TestSelectPhoto(t *testing.T) {
	tests := []struct {
		name       string
		media      *twitterxapi.Media
		wantURL    string
		wantThumb  string
		wantWidth  int
		wantHeight int
	}{
		{
			name:       "nil media",
			media:      nil,
			wantURL:    "",
			wantThumb:  "",
			wantWidth:  0,
			wantHeight: 0,
		},
		{
			name:       "no photos",
			media:      &twitterxapi.Media{},
			wantURL:    "",
			wantThumb:  "",
			wantWidth:  0,
			wantHeight: 0,
		},
		{
			name: "single photo",
			media: &twitterxapi.Media{
				Photos: []twitterxapi.Photo{
					{URL: "https://img/1.jpg", Width: 640, Height: 480},
				},
			},
			wantURL:    "https://img/1.jpg",
			wantThumb:  "https://img/1.jpg",
			wantWidth:  640,
			wantHeight: 480,
		},
		{
			name: "multiple photos with mosaic",
			media: &twitterxapi.Media{
				Photos: []twitterxapi.Photo{
					{URL: "https://img/1.jpg", Width: 640, Height: 480},
					{URL: "https://img/2.jpg", Width: 800, Height: 600},
				},
				Mosaic: &twitterxapi.Mosaic{
					Formats: map[string]string{"jpeg": "https://img/mosaic.jpg"},
					Width:   intPtr(1200),
					Height:  intPtr(800),
				},
			},
			wantURL:    "https://img/mosaic.jpg",
			wantThumb:  "https://img/mosaic.jpg",
			wantWidth:  1200,
			wantHeight: 800,
		},
		{
			name: "multiple photos with unusable mosaic",
			media: &twitterxapi.Media{
				Photos: []twitterxapi.Photo{
					{URL: "https://img/1.jpg", Width: 640, Height: 480},
					{URL: "https://img/2.jpg", Width: 800, Height: 600},
				},
				Mosaic: &twitterxapi.Mosaic{
					Formats: map[string]string{"webp": "https://img/mosaic.webp"},
				},
			},
			wantURL:    "https://img/1.jpg",
			wantThumb:  "https://img/1.jpg",
			wantWidth:  640,
			wantHeight: 480,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotURL, gotThumb, gotWidth, gotHeight := SelectPhoto(tt.media)
			if gotURL != tt.wantURL || gotThumb != tt.wantThumb || gotWidth != tt.wantWidth || gotHeight != tt.wantHeight {
				t.Errorf("SelectPhoto() = (%q, %q, %d, %d), want (%q, %q, %d, %d)",
					gotURL, gotThumb, gotWidth, gotHeight,
					tt.wantURL, tt.wantThumb, tt.wantWidth, tt.wantHeight)
			}
		})
	}
}

func TestMediaPreview(t *testing.T) {
	tests := []struct {
		name     string
		media    *twitterxapi.Media
		wantURL  string
		wantKind string
	}{
		{
			name:     "nil media",
			media:    nil,
			wantURL:  "",
			wantKind: "",
		},
		{
			name:     "empty media",
			media:    &twitterxapi.Media{},
			wantURL:  "",
			wantKind: "",
		},
		{
			name: "video thumbnail",
			media: &twitterxapi.Media{
				Videos: []twitterxapi.Video{
					{ThumbnailURL: "https://thumb/video.jpg"},
				},
			},
			wantURL:  "https://thumb/video.jpg",
			wantKind: "video",
		},
		{
			name: "single photo",
			media: &twitterxapi.Media{
				Photos: []twitterxapi.Photo{
					{URL: "https://img/photo.jpg"},
				},
			},
			wantURL:  "https://img/photo.jpg",
			wantKind: "photo",
		},
		{
			name: "multiple photos with mosaic",
			media: &twitterxapi.Media{
				Photos: []twitterxapi.Photo{
					{URL: "https://img/1.jpg"},
					{URL: "https://img/2.jpg"},
				},
				Mosaic: &twitterxapi.Mosaic{
					Formats: map[string]string{"jpeg": "https://img/mosaic.jpg"},
				},
			},
			wantURL:  "https://img/mosaic.jpg",
			wantKind: "mosaic",
		},
		{
			name: "video takes priority over photos",
			media: &twitterxapi.Media{
				Videos: []twitterxapi.Video{
					{ThumbnailURL: "https://thumb/video.jpg"},
				},
				Photos: []twitterxapi.Photo{
					{URL: "https://img/photo.jpg"},
				},
			},
			wantURL:  "https://thumb/video.jpg",
			wantKind: "video",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotURL, gotKind := MediaPreview(tt.media)
			if gotURL != tt.wantURL || gotKind != tt.wantKind {
				t.Errorf("MediaPreview() = (%q, %q), want (%q, %q)",
					gotURL, gotKind, tt.wantURL, tt.wantKind)
			}
		})
	}
}

func TestMediaHint(t *testing.T) {
	tests := []struct {
		kind string
		want string
	}{
		{"video", "Video"},
		{"mosaic", "Mosaic"},
		{"photo", "Photo"},
		{"", ""},
		{"unknown", ""},
	}

	for _, tt := range tests {
		t.Run(tt.kind, func(t *testing.T) {
			if got := MediaHint(tt.kind); got != tt.want {
				t.Errorf("MediaHint(%q) = %q, want %q", tt.kind, got, tt.want)
			}
		})
	}
}

func TestPickMosaicURL(t *testing.T) {
	tests := []struct {
		name   string
		mosaic *twitterxapi.Mosaic
		want   string
	}{
		{
			name:   "nil mosaic",
			mosaic: nil,
			want:   "",
		},
		{
			name:   "empty formats",
			mosaic: &twitterxapi.Mosaic{},
			want:   "",
		},
		{
			name: "jpeg preferred",
			mosaic: &twitterxapi.Mosaic{
				Formats: map[string]string{
					"jpg":  "https://img/one.jpg",
					"jpeg": "https://img/two.jpg",
				},
			},
			want: "https://img/two.jpg",
		},
		{
			name: "jpg fallback",
			mosaic: &twitterxapi.Mosaic{
				Formats: map[string]string{
					"jpg": "https://img/one.jpg",
				},
			},
			want: "https://img/one.jpg",
		},
		{
			name: "unknown formats",
			mosaic: &twitterxapi.Mosaic{
				Formats: map[string]string{
					"png": "https://img/one.png",
				},
			},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := PickMosaicURL(tt.mosaic); got != tt.want {
				t.Errorf("PickMosaicURL() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestMosaicDimensions(t *testing.T) {
	tests := []struct {
		name   string
		mosaic *twitterxapi.Mosaic
		wantW  int
		wantH  int
	}{
		{
			name:   "nil mosaic",
			mosaic: nil,
			wantW:  0,
			wantH:  0,
		},
		{
			name:   "no sizes",
			mosaic: &twitterxapi.Mosaic{},
			wantW:  0,
			wantH:  0,
		},
		{
			name:   "width only",
			mosaic: &twitterxapi.Mosaic{Width: intPtr(640)},
			wantW:  640,
			wantH:  0,
		},
		{
			name:   "width and height",
			mosaic: &twitterxapi.Mosaic{Width: intPtr(640), Height: intPtr(480)},
			wantW:  640,
			wantH:  480,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotW, gotH := MosaicDimensions(tt.mosaic)
			if gotW != tt.wantW || gotH != tt.wantH {
				t.Errorf("MosaicDimensions() = (%d, %d), want (%d, %d)", gotW, gotH, tt.wantW, tt.wantH)
			}
		})
	}
}

func TestMimeTypeForVideo(t *testing.T) {
	tests := []struct {
		name   string
		format string
		want   string
	}{
		{
			name:   "empty format",
			format: "",
			want:   "video/mp4",
		},
		{
			name:   "has mime type",
			format: "video/quicktime",
			want:   "video/quicktime",
		},
		{
			name:   "simple format",
			format: "webm",
			want:   "video/webm",
		},
		{
			name:   "trimmed format",
			format: "  mp4  ",
			want:   "video/mp4",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MimeTypeForVideo(tt.format); got != tt.want {
				t.Errorf("MimeTypeForVideo() = %q, want %q", got, tt.want)
			}
		})
	}
}

func intPtr(v int) *int {
	return &v
}
