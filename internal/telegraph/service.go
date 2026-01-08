package telegraph

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// TelegraphClient - інтерфейс для HTTP клієнта
type TelegraphClient interface {
	CreateAccount(ctx context.Context, req CreateAccountRequest) (*Account, error)
	CreatePage(ctx context.Context, req CreatePageRequest) (*Page, error)
}

// Service - сервіс для роботи з Telegraph
type Service struct {
	client     TelegraphClient
	converter  *Converter
	shortName  string
	authorName string
	authorURL  string

	cachedAccount *Account
	mu            sync.Mutex
}

// Option - функція для налаштування сервісу
type Option func(*Service)

// WithShortName встановлює short_name для акаунту
func WithShortName(name string) Option {
	return func(s *Service) {
		s.shortName = name
	}
}

// WithAuthorName встановлює author_name
func WithAuthorName(name string) Option {
	return func(s *Service) {
		s.authorName = name
	}
}

// WithAuthorURL встановлює author_url
func WithAuthorURL(url string) Option {
	return func(s *Service) {
		s.authorURL = url
	}
}

// NewService створює новий сервіс
func NewService(client TelegraphClient, opts ...Option) *Service {
	s := &Service{
		client:     client,
		converter:  NewConverter(),
		shortName:  fmt.Sprintf("TwitterX-%d", time.Now().UnixNano()%100000),
		authorName: "TwitterX Bot",
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

// CreateArticle створює статтю в Telegraph та повертає URL
func (s *Service) CreateArticle(ctx context.Context, text, title string) (string, error) {
	// Валідуємо заголовок
	validTitle, err := s.converter.ValidateTitle(title)
	if err != nil {
		return "", err
	}

	// Конвертуємо текст в DOM
	content, err := s.converter.TextToNodes(text)
	if err != nil {
		return "", err
	}

	// Отримуємо або створюємо акаунт
	account, err := s.getOrCreateAccount(ctx)
	if err != nil {
		return "", err
	}

	// Створюємо сторінку
	page, err := s.client.CreatePage(ctx, CreatePageRequest{
		AccessToken:   account.AccessToken,
		Title:         validTitle,
		Content:       content,
		AuthorName:    s.authorName,
		AuthorURL:     s.authorURL,
		ReturnContent: false,
	})
	if err != nil {
		return "", err
	}

	return page.URL, nil
}

// getOrCreateAccount отримує акаунт з кешу або створює новий
func (s *Service) getOrCreateAccount(ctx context.Context) (*Account, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Повертаємо з кешу якщо є
	if s.cachedAccount != nil && s.cachedAccount.AccessToken != "" {
		return s.cachedAccount, nil
	}

	// Створюємо новий акаунт
	account, err := s.client.CreateAccount(ctx, CreateAccountRequest{
		ShortName:  s.shortName,
		AuthorName: s.authorName,
		AuthorURL:  s.authorURL,
	})
	if err != nil {
		return nil, err
	}

	if account.AccessToken == "" {
		return nil, ErrNoAccessToken
	}

	// Кешуємо
	s.cachedAccount = account

	return account, nil
}
