package translates

type PapagoMessage struct {
	Message TranslatedMessage `json:"message"`
}

type TranslatedMessage struct {
	Result TranslatedResult `json:"result"`
}

type TranslatedResult struct {
	Text string `json:"translatedText"`
}
