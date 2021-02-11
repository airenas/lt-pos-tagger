package service

import (
	"strings"
)

//PreprocessorImpl implement text preprocessing
type PreprocessorImpl struct {
}

// NewPreprocessorImpl will instantiate a preprocessor
func NewPreprocessorImpl() (*PreprocessorImpl, error) {
	res := &PreprocessorImpl{}
	return res, nil
}

// Process does initial text processing
func (pr *PreprocessorImpl) Process(text string) (string, error) {
	text = strings.ReplaceAll(text, "‘", "`")

	//replace dash to space as it does not get parsed
	text = strings.ReplaceAll(text, "–", "-")
	text = strings.ReplaceAll(text, "‑", "-")
	text = strings.ReplaceAll(text, "-", " ")
	return text, nil
}
