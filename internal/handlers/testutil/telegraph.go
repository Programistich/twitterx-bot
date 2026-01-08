package testutil

import "context"

// FakeTelegraph is a test double for Telegraph article creation.
type FakeTelegraph struct {
	Called   bool
	GotText  string
	GotTitle string
	URL      string
	Err      error
}

func (f *FakeTelegraph) CreateArticle(_ context.Context, text, title string) (string, error) {
	f.Called = true
	f.GotText = text
	f.GotTitle = title
	return f.URL, f.Err
}
