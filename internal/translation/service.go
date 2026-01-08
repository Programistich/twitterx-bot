package translation

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
)

const (
	DefaultBaseURL = "https://translate.google.com/translate_a/single"
	DefaultTimeout = 10
)

// Translator defines the interface for the translation service.
type Translator interface {
	Translate(ctx context.Context, text string, to Language) (*Translation, error)
}

// Service provides translations via Google Translate.
type Service struct {
	httpClient *http.Client
	baseURL    string
	log        *slog.Logger
}

// NewService creates a new translation service.
func NewService(httpClient *http.Client, baseURL string) *Service {
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}
	return &Service{
		httpClient: httpClient,
		baseURL:    baseURL,
		log:        slog.Default().With("component", "translation"),
	}
}

// Translate translates text into the specified language.
func (s *Service) Translate(ctx context.Context, text string, to Language) (*Translation, error) {
	s.log.Debug("starting translation request",
		"text_length", len(text),
		"target_language", to.ISO,
	)

	form := url.Values{}
	form.Set("client", "it")
	form.Add("dt", "qca")
	form.Add("dt", "t")
	form.Add("dt", "rmt")
	form.Add("dt", "bd")
	form.Add("dt", "rms")
	form.Add("dt", "sos")
	form.Add("dt", "md")
	form.Add("dt", "gt")
	form.Add("dt", "ld")
	form.Add("dt", "ss")
	form.Add("dt", "ex")
	form.Set("otf", "2")
	form.Set("dj", "1")
	form.Set("hl", "en")
	form.Set("ie", "UTF-8")
	form.Set("oe", "UTF-8")
	form.Set("sl", "auto")
	form.Set("tl", to.ISO)
	form.Set("q", text)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.baseURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		s.log.Error("translation request failed", "err", err)
		return nil, fmt.Errorf("translation request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		s.log.Error("translation API returned error", "status", resp.StatusCode)
		return nil, fmt.Errorf("translation API returned status %d", resp.StatusCode)
	}

	var googleResp googleTranslateResponse
	if err := json.NewDecoder(resp.Body).Decode(&googleResp); err != nil {
		s.log.Error("failed to decode translation response", "err", err)
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	var translatedText strings.Builder
	for _, sentence := range googleResp.Sentences {
		translatedText.WriteString(sentence.Text)
	}

	translation := &Translation{
		From: Language{ISO: googleResp.Src},
		To:   to,
		Text: translatedText.String(),
	}

	s.log.Info("translation completed successfully",
		"from", googleResp.Src,
		"to", to.ISO,
	)
	s.log.Debug("translation result",
		"original_length", len(text),
		"translated_length", len(translation.Text),
	)

	return translation, nil
}

// LanguageFromISO returns a Language for the given ISO code.
func LanguageFromISO(iso string) Language {
	switch strings.ToLower(iso) {
	case "uk":
		return LangUkrainian
	case "en":
		return LangEnglish
	case "ru":
		return LangRussian
	default:
		return Language{ISO: iso}
	}
}
