package translation

// Language represents a language for translation.
type Language struct {
	ISO  string
	Name string
}

// Supported languages.
var (
	LangUkrainian = Language{ISO: "uk", Name: "Ukrainian"}
	LangEnglish   = Language{ISO: "en", Name: "English"}
	LangRussian   = Language{ISO: "ru", Name: "Russian"}
)

// Translation represents a translation result.
type Translation struct {
	From Language
	To   Language
	Text string
}

// googleTranslateResponse is the response from Google Translate API.
type googleTranslateResponse struct {
	Sentences []sentence `json:"sentences"`
	Src       string     `json:"src"`
}

// sentence represents a translated sentence.
type sentence struct {
	Text string `json:"trans"`
	Orig string `json:"orig"`
}
