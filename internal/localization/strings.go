package localization

// StringKey identifies a localizable string.
type StringKey string

const (
	// Help text
	KeyHelpText StringKey = "help_text"

	// Lang command
	KeySelectLanguage  StringKey = "select_language"
	KeyLanguageChanged StringKey = "language_changed"

	// Callback answers
	KeyInvalidCallbackData StringKey = "invalid_callback_data"
	KeyFetchingFullChain   StringKey = "fetching_full_chain"
	KeyCannotDeleteMessage StringKey = "cannot_delete_message"
	KeyDeleted             StringKey = "deleted"

	// Error messages
	KeyInvalidLanguage    StringKey = "invalid_language"
	KeyErrorSavingLanguage StringKey = "error_saving_language"

	// Button texts
	KeySendFullChain  StringKey = "send_full_chain"
	KeyDeleteOriginal StringKey = "delete_original"

	// Media hints
	KeyVideo  StringKey = "video"
	KeyMosaic StringKey = "mosaic"
	KeyPhoto  StringKey = "photo"

	// Tweet headers
	KeyTweet        StringKey = "tweet"
	KeyTweetBy      StringKey = "tweet_by"
	KeyFrom         StringKey = "from"
	KeyBy           StringKey = "by"
	KeyLinkFallback StringKey = "link_fallback"
)
