package twitterxapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const DefaultBaseURL = "http://127.0.0.1:8080"

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(baseURL string) *Client {
	base := strings.TrimRight(baseURL, "/")
	if base == "" {
		base = DefaultBaseURL
	}

	return &Client{
		baseURL: base,
		httpClient: &http.Client{
			Timeout: 12 * time.Second,
		},
	}
}

type TweetResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Tweet   *Tweet `json:"tweet,omitempty"`
}

type Tweet struct {
	ID     string `json:"id"`
	URL    string `json:"url"`
	Text   string `json:"text"`
	Author Author `json:"author"`
	Media  *Media `json:"media,omitempty"`
}

type Author struct {
	Name       string `json:"name"`
	ScreenName string `json:"screen_name"`
	AvatarURL  string `json:"avatar_url"`
}

type Media struct {
	Photos []Photo `json:"photos,omitempty"`
	Videos []Video `json:"videos,omitempty"`
	Mosaic *Mosaic `json:"mosaic,omitempty"`
}

type Photo struct {
	URL    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

type Video struct {
	URL          string   `json:"url"`
	ThumbnailURL string   `json:"thumbnail_url"`
	Width        int      `json:"width"`
	Height       int      `json:"height"`
	Format       string   `json:"format,omitempty"`
	Type         string   `json:"type,omitempty"`
	Duration     *float64 `json:"duration,omitempty"`
}

type Mosaic struct {
	Type    string            `json:"type"`
	Width   *int              `json:"width,omitempty"`
	Height  *int              `json:"height,omitempty"`
	Formats map[string]string `json:"formats,omitempty"`
}

func (c *Client) GetTweet(ctx context.Context, username, tweetID string) (*Tweet, error) {
	if username == "" {
		return nil, errors.New("username is required")
	}
	if tweetID == "" {
		return nil, errors.New("tweet id is required")
	}

	reqURL := fmt.Sprintf("%s/api/users/%s/tweets/%s", c.baseURL, url.PathEscape(username), url.PathEscape(tweetID))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch tweet: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("api status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var tweetResp TweetResponse
	if err := json.Unmarshal(body, &tweetResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	if tweetResp.Code != http.StatusOK {
		return nil, fmt.Errorf("api error %d: %s", tweetResp.Code, tweetResp.Message)
	}
	if tweetResp.Tweet == nil {
		return nil, errors.New("api response missing tweet")
	}

	return tweetResp.Tweet, nil
}
