package translation

// Language представляє мову для перекладу
type Language struct {
	ISO  string
	Name string
}

// Підтримувані мови
var (
	LangUkrainian = Language{ISO: "uk", Name: "Ukrainian"}
	LangEnglish   = Language{ISO: "en", Name: "English"}
	LangRussian   = Language{ISO: "ru", Name: "Russian"}
	LangGerman    = Language{ISO: "de", Name: "German"}
	LangFrench    = Language{ISO: "fr", Name: "French"}
	LangSpanish   = Language{ISO: "es", Name: "Spanish"}
	LangItalian   = Language{ISO: "it", Name: "Italian"}
	LangPolish    = Language{ISO: "pl", Name: "Polish"}
	LangJapanese  = Language{ISO: "ja", Name: "Japanese"}
	LangChinese   = Language{ISO: "zh", Name: "Chinese"}
)

// Translation представляє результат перекладу
type Translation struct {
	From Language
	To   Language
	Text string
}

// googleTranslateResponse - відповідь від Google Translate API
type googleTranslateResponse struct {
	Sentences []sentence `json:"sentences"`
	Src       string     `json:"src"`
}

// sentence - речення з перекладом
type sentence struct {
	Text string `json:"trans"`
	Orig string `json:"orig"`
}
