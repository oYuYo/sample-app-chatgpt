package main

type OpenaiConfig struct {
	ChatUrl        string
	WhisperUrl     string
	ApiKey         string
	ChatUseMode    string
	WhisperUseMode string
}

type Role struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Roles []Role

type oaRes struct {
	Id      string         `json:"id"`
	Object  string         `json:"object"`
	Created int32          `json:"created"`
	Usage   oaResUsage     `json:"usage"`
	Choices []oaResChoices `json:"choices"`
}

type oaResUsage struct {
	PromptTokens     int32 `json:"prompt_tokens"`
	CompletionTokens int32 `json:"completion_tokens"`
	TotalTokens      int32 `json:"total_tokens"`
}

type oaResChoices struct {
	Message      Role   `json:"message"`
	FinishReason string `json:"finish_reason"`
	Index        int32  `json:"index"`
}

type oaResWhisper struct {
	Text string `json:"text"`
}
