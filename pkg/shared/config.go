package shared

type ModelConfig struct {
	BaseURL string `json:"base_url"`
	ApiKey  string `json:"api_key"`
	Model   string `json:"model"`

	ContextWindow int `json:"context_window"`
}
