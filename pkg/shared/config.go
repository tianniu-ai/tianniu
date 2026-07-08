package shared

type ModelConfig struct {
	BaseURL string `yaml:"base_url"`
	ApiKey  string `yaml:"api_key"`
	Model   string `yaml:"model"`

	ContextWindow int `yaml:"context_window"`
}
