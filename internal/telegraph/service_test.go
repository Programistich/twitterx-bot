package telegraph

import (
	"context"
	"testing"
)

// fakeClient для тестування
type fakeClient struct {
	createAccountResp *Account
	createAccountErr  error
	createPageResp    *Page
	createPageErr     error

	// Для перевірки викликів
	createAccountCalls int
	createPageCalls    int
}

func (f *fakeClient) CreateAccount(ctx context.Context, req CreateAccountRequest) (*Account, error) {
	f.createAccountCalls++
	return f.createAccountResp, f.createAccountErr
}

func (f *fakeClient) CreatePage(ctx context.Context, req CreatePageRequest) (*Page, error) {
	f.createPageCalls++
	return f.createPageResp, f.createPageErr
}

func TestService_CreateArticle_Success(t *testing.T) {
	client := &fakeClient{
		createAccountResp: &Account{
			ShortName:   "TestBot",
			AccessToken: "test-token",
		},
		createPageResp: &Page{
			URL:   "https://telegra.ph/test-article",
			Title: "Test Title",
		},
	}

	service := NewService(client)

	url, err := service.CreateArticle(context.Background(), "Test content", "Test Title")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if url != "https://telegra.ph/test-article" {
		t.Errorf("unexpected URL: %s", url)
	}

	// Перевіряємо що акаунт створено
	if client.createAccountCalls != 1 {
		t.Errorf("expected 1 createAccount call, got %d", client.createAccountCalls)
	}
}

func TestService_CreateArticle_CachesAccount(t *testing.T) {
	client := &fakeClient{
		createAccountResp: &Account{
			ShortName:   "TestBot",
			AccessToken: "test-token",
		},
		createPageResp: &Page{
			URL: "https://telegra.ph/test-article",
		},
	}

	service := NewService(client)

	// Перший виклик
	_, err := service.CreateArticle(context.Background(), "Content 1", "Title 1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Другий виклик - акаунт має бути з кешу
	_, err = service.CreateArticle(context.Background(), "Content 2", "Title 2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// createAccount має викликатись тільки 1 раз
	if client.createAccountCalls != 1 {
		t.Errorf("expected 1 createAccount call (cached), got %d", client.createAccountCalls)
	}

	// createPage має викликатись 2 рази
	if client.createPageCalls != 2 {
		t.Errorf("expected 2 createPage calls, got %d", client.createPageCalls)
	}
}

func TestService_CreateArticle_InvalidTitle(t *testing.T) {
	client := &fakeClient{}
	service := NewService(client)

	_, err := service.CreateArticle(context.Background(), "Content", "")
	if err != ErrTitleEmpty {
		t.Errorf("expected ErrTitleEmpty, got %v", err)
	}
}

func TestService_CreateArticle_EmptyContent(t *testing.T) {
	client := &fakeClient{}
	service := NewService(client)

	_, err := service.CreateArticle(context.Background(), "", "Title")
	if err != ErrContentEmpty {
		t.Errorf("expected ErrContentEmpty, got %v", err)
	}
}

func TestService_CreateArticle_NoAccessToken(t *testing.T) {
	client := &fakeClient{
		createAccountResp: &Account{
			ShortName: "TestBot",
			// AccessToken відсутній!
		},
	}

	service := NewService(client)

	_, err := service.CreateArticle(context.Background(), "Content", "Title")
	if err != ErrNoAccessToken {
		t.Errorf("expected ErrNoAccessToken, got %v", err)
	}
}

func TestService_WithOptions(t *testing.T) {
	client := &fakeClient{
		createAccountResp: &Account{
			ShortName:   "TestBot",
			AccessToken: "test-token",
		},
		createPageResp: &Page{
			URL: "https://telegra.ph/test-article",
		},
	}

	service := NewService(client,
		WithShortName("MyBot"),
		WithAuthorName("Custom Author"),
		WithAuthorURL("https://example.com"),
	)

	if service.shortName != "MyBot" {
		t.Errorf("expected shortName 'MyBot', got '%s'", service.shortName)
	}
	if service.authorName != "Custom Author" {
		t.Errorf("expected authorName 'Custom Author', got '%s'", service.authorName)
	}
	if service.authorURL != "https://example.com" {
		t.Errorf("expected authorURL 'https://example.com', got '%s'", service.authorURL)
	}
}
