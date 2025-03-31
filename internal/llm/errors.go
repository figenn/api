package llm

import "errors"

var (
	ErrEmptyPrompt = errors.New("llm: prompt cannot be empty")
)
