package tweetsvc

import "errors"

var (
	ErrFetchTweet = errors.New("fetch tweet")
	ErrSendTweet  = errors.New("send tweet")
)
