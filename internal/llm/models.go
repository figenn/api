package llm

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Prompt struct {
	Input string `json:"input" form:"input"`
}

type Completion struct {
	Output string `json:"output"`
}
