package telegraph

// Response - обгортка відповіді Telegraph API
type Response[T any] struct {
	OK     bool   `json:"ok"`
	Result T      `json:"result,omitempty"`
	Error  string `json:"error,omitempty"`
}

// Account - Telegraph акаунт
type Account struct {
	ShortName   string `json:"short_name"`
	AuthorName  string `json:"author_name,omitempty"`
	AuthorURL   string `json:"author_url,omitempty"`
	AccessToken string `json:"access_token,omitempty"`
	AuthURL     string `json:"auth_url,omitempty"`
	PageCount   int    `json:"page_count,omitempty"`
}

// Page - Telegraph сторінка (стаття)
type Page struct {
	Path        string `json:"path"`
	URL         string `json:"url"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	AuthorName  string `json:"author_name,omitempty"`
	AuthorURL   string `json:"author_url,omitempty"`
	ImageURL    string `json:"image_url,omitempty"`
	Views       int    `json:"views,omitempty"`
	CanEdit     bool   `json:"can_edit,omitempty"`
}

// CreateAccountRequest - запит на створення акаунту
type CreateAccountRequest struct {
	ShortName  string
	AuthorName string
	AuthorURL  string
}

// CreatePageRequest - запит на створення сторінки
type CreatePageRequest struct {
	AccessToken   string
	Title         string
	Content       []any // Node або string
	AuthorName    string
	AuthorURL     string
	ReturnContent bool
}
