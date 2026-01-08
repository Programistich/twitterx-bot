package telegraph

import "errors"

var (
	ErrAPIError       = errors.New("telegraph api error")
	ErrContentTooLong = errors.New("content exceeds 64KB limit")
	ErrContentEmpty   = errors.New("content is empty")
	ErrTitleTooLong   = errors.New("title exceeds 256 characters")
	ErrTitleEmpty     = errors.New("title is empty")
	ErrNoAccessToken  = errors.New("account has no access token")
)
