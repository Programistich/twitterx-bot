package tweet

import "context"

// ArticleCreator - інтерфейс для створення статей (Telegraph)
type ArticleCreator interface {
	CreateArticle(ctx context.Context, text, title string) (string, error)
}
