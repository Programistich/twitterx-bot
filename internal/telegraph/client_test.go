package telegraph

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_CreateAccount_Success(t *testing.T) {
	// Mock сервер
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Перевіряємо метод та шлях
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/createAccount" {
			t.Errorf("expected /createAccount, got %s", r.URL.Path)
		}

		// Перевіряємо content-type
		if r.Header.Get("Content-Type") != "application/x-www-form-urlencoded" {
			t.Errorf("expected form content-type")
		}

		// Відповідь
		resp := Response[Account]{
			OK: true,
			Result: Account{
				ShortName:   "TestBot",
				AccessToken: "test-token-123",
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(&http.Client{}, server.URL)

	account, err := client.CreateAccount(context.Background(), CreateAccountRequest{
		ShortName:  "TestBot",
		AuthorName: "Test Author",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if account.ShortName != "TestBot" {
		t.Errorf("expected ShortName 'TestBot', got '%s'", account.ShortName)
	}

	if account.AccessToken != "test-token-123" {
		t.Errorf("expected AccessToken 'test-token-123', got '%s'", account.AccessToken)
	}
}

func TestClient_CreateAccount_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := Response[Account]{
			OK:    false,
			Error: "SHORT_NAME_REQUIRED",
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(&http.Client{}, server.URL)

	_, err := client.CreateAccount(context.Background(), CreateAccountRequest{})

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_CreatePage_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/createPage" {
			t.Errorf("expected /createPage, got %s", r.URL.Path)
		}

		resp := Response[Page]{
			OK: true,
			Result: Page{
				Path:  "test-article-01-08",
				URL:   "https://telegra.ph/test-article-01-08",
				Title: "Test Article",
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(&http.Client{}, server.URL)

	page, err := client.CreatePage(context.Background(), CreatePageRequest{
		AccessToken: "test-token",
		Title:       "Test Article",
		Content:     []any{"Test content"},
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if page.URL != "https://telegra.ph/test-article-01-08" {
		t.Errorf("unexpected URL: %s", page.URL)
	}
}

func TestClient_CreatePage_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := Response[Page]{
			OK:    false,
			Error: "ACCESS_TOKEN_INVALID",
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(&http.Client{}, server.URL)

	_, err := client.CreatePage(context.Background(), CreatePageRequest{
		AccessToken: "invalid-token",
		Title:       "Test",
		Content:     []any{"content"},
	})

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
