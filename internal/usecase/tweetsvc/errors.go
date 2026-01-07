package tweetsvc

import "errors"

var (
	ErrFetchTweet  = errors.New("fetch tweet")
	ErrSendTweet   = errors.New("send tweet")
	ErrBuildChain  = errors.New("build chain")
	ErrSendChain   = errors.New("send chain")
	ErrBuildInline = errors.New("build inline result")
)
