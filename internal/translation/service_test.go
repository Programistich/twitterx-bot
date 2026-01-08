package translation

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestService_Translate_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}

		if r.Header.Get("Content-Type") != "application/x-www-form-urlencoded" {
			t.Errorf("expected form content-type")
		}

		if err := r.ParseForm(); err != nil {
			t.Fatalf("failed to parse form: %v", err)
		}

		if r.Form.Get("tl") != "uk" {
			t.Errorf("expected target language 'uk', got '%s'", r.Form.Get("tl"))
		}

		if r.Form.Get("q") != "Hello world" {
			t.Errorf("expected query 'Hello world', got '%s'", r.Form.Get("q"))
		}

		resp := googleTranslateResponse{
			Sentences: []sentence{
				{Text: "Привіт ", Orig: "Hello "},
				{Text: "світ", Orig: "world"},
			},
			Src: "en",
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	service := NewService(&http.Client{}, server.URL)

	translation, err := service.Translate(context.Background(), "Hello world", LangUkrainian)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if translation.Text != "Привіт світ" {
		t.Errorf("expected 'Привіт світ', got '%s'", translation.Text)
	}

	if translation.From.ISO != "en" {
		t.Errorf("expected source language 'en', got '%s'", translation.From.ISO)
	}

	if translation.To.ISO != "uk" {
		t.Errorf("expected target language 'uk', got '%s'", translation.To.ISO)
	}
}

func TestService_Translate_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	service := NewService(&http.Client{}, server.URL)

	_, err := service.Translate(context.Background(), "Hello", LangUkrainian)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestService_Translate_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	service := NewService(&http.Client{}, server.URL)

	_, err := service.Translate(context.Background(), "Hello", LangUkrainian)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestService_Translate_EmptyResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := googleTranslateResponse{
			Sentences: []sentence{},
			Src:       "en",
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	service := NewService(&http.Client{}, server.URL)

	translation, err := service.Translate(context.Background(), "Hello", LangUkrainian)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if translation.Text != "" {
		t.Errorf("expected empty text, got '%s'", translation.Text)
	}
}

func TestNewService_DefaultBaseURL(t *testing.T) {
	service := NewService(&http.Client{}, "")

	if service.baseURL != DefaultBaseURL {
		t.Errorf("expected default base URL '%s', got '%s'", DefaultBaseURL, service.baseURL)
	}
}

func TestLanguageFromISO(t *testing.T) {
	tests := []struct {
		iso      string
		expected Language
	}{
		{"uk", LangUkrainian},
		{"UK", LangUkrainian},
		{"en", LangEnglish},
		{"ru", LangRussian},
		{"unknown", Language{ISO: "unknown"}},
	}

	for _, tt := range tests {
		t.Run(tt.iso, func(t *testing.T) {
			result := LanguageFromISO(tt.iso)
			if result.ISO != tt.expected.ISO {
				t.Errorf("expected ISO '%s', got '%s'", tt.expected.ISO, result.ISO)
			}
		})
	}
}
