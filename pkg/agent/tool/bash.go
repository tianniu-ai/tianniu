package tool

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/shared"
)

const (
	defaultTimeout       = 30 * time.Second
	defaultMaxOutput     = 64 * 1024
	defaultMaxCommandLen = 4096
)

var dangerousPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)\brm\s+(-[a-zA-Z]*f[a-zA-Z]*\s+|.*--no-preserve-root\s+)/`),
	regexp.MustCompile(`(?i)\bsudo\b`),
	regexp.MustCompile(`(?i)\bsu\b\s+`),
	regexp.MustCompile(`(?i)\bmkfs\b`),
	regexp.MustCompile(`(?i)\bdd\s+.*of=/dev/`),
	regexp.MustCompile(`(?i)\bformat\b\s+[A-Za-z]:`),
	regexp.MustCompile(`(?i)\bchmod\s+([0-7]*7[0-7]*[0-7]*\s+/|.*-R\s+/)`),
	regexp.MustCompile(`(?i)\bchown\s+.*-R\s+/`),
	regexp.MustCompile(`(?i)\bshutdown\b`),
	regexp.MustCompile(`(?i)\breboot\b`),
	regexp.MustCompile(`(?i)\binit\s+[06]`),
	regexp.MustCompile(`(?i)\bmount\b.*\b/dev/`),
	regexp.MustCompile(`(?i)\bumount\b`),
	regexp.MustCompile(`(?i)\biptables\b`),
	regexp.MustCompile(`(?i)\bnc\b.*-e`),
	regexp.MustCompile(`(?i)\bncat\b.*-e`),
	regexp.MustCompile(`(?i)\bcurl\b.*\|\s*(ba)?sh`),
	regexp.MustCompile(`(?i)\bwget\b.*\|\s*(ba)?sh`),
	regexp.MustCompile(`(?i)\bchmod\s+[0-7]*777`),
	regexp.MustCompile(`(?i)\bkill\s+-9\s+1\b`),
	regexp.MustCompile(`(?i)\bkillall\b`),
	regexp.MustCompile(`(?i)\b:\(\)\s*\{\s*:\|:\&\s*\}`),
}

var blockedEnvPrefixes = []string{
	"JWT_",
	"API_KEY",
	"SECRET",
	"TOKEN",
	"PASSWORD",
	"CREDENTIAL",
	"AWS_",
	"DATABASE_",
}

type BashToolConfig struct {
	Timeout        time.Duration `json:"timeout"`
	MaxOutput      int           `json:"max_output"`
	WorkDir        string        `json:"work_dir"`
	Disabled       bool          `json:"disabled"`
	AllowDangerous bool          `json:"allow_dangerous"`
}

type BashTool struct {
	config BashToolConfig
}

func NewBashTool(config BashToolConfig) *BashTool {
	if config.Timeout <= 0 {
		config.Timeout = defaultTimeout
	}
	if config.MaxOutput <= 0 {
		config.MaxOutput = defaultMaxOutput
	}
	return &BashTool{config: config}
}

type BashToolParam struct {
	Command string `json:"command"`
}

func (t *BashTool) ToolName() AgentTool {
	return AgentToolBash
}

func (t *BashTool) Info() openai.ChatCompletionToolUnionParam {
	return openai.ChatCompletionFunctionTool(shared.FunctionDefinitionParam{
		Name:        string(AgentToolBash),
		Description: openai.String("execute a bash command in a sandboxed environment with security restrictions"),
		Parameters: openai.FunctionParameters{
			"type": "object",
			"properties": map[string]any{
				"command": map[string]any{
					"type":        "string",
					"description": "the bash command to execute (subject to security restrictions)",
				},
			},
			"required": []string{"command"},
		},
	})
}

func (t *BashTool) Execute(ctx context.Context, argumentsInJSON string) (string, error) {
	if t.config.Disabled {
		return "", fmt.Errorf("bash tool is disabled")
	}

	p := BashToolParam{}
	if err := json.Unmarshal([]byte(argumentsInJSON), &p); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	if len(p.Command) > defaultMaxCommandLen {
		return "", fmt.Errorf("command too long (max %d characters)", defaultMaxCommandLen)
	}

	if !t.config.AllowDangerous {
		if err := validateCommand(p.Command); err != nil {
			return "", err
		}
	}

	timeout := t.config.Timeout
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(ctx, "cmd", "/C", p.Command)
	} else {
		cmd = exec.CommandContext(ctx, "sh", "-c", p.Command)
	}

	if t.config.WorkDir != "" {
		absDir, err := filepath.Abs(t.config.WorkDir)
		if err != nil {
			return "", fmt.Errorf("invalid work_dir: %w", err)
		}
		cmd.Dir = absDir
	}

	cmd.Env = filterEnv()

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	if ctx.Err() == context.DeadlineExceeded {
		return "", fmt.Errorf("command timed out after %v", timeout)
	}

	out := stdout.String()
	errOut := stderr.String()
	combined := out
	if errOut != "" {
		if combined != "" {
			combined += "\n"
		}
		combined += errOut
	}

	if len(combined) > t.config.MaxOutput {
		combined = combined[:t.config.MaxOutput] + fmt.Sprintf("\n... [output truncated, max %d bytes]", t.config.MaxOutput)
	}

	if err != nil {
		return combined, fmt.Errorf("command failed: %w", err)
	}
	return combined, nil
}

func validateCommand(cmd string) error {
	for _, pattern := range dangerousPatterns {
		if pattern.MatchString(cmd) {
			return fmt.Errorf("command contains dangerous pattern and is blocked for security reasons")
		}
	}
	return nil
}

func filterEnv() []string {
	env := make([]string, 0, len(os.Environ()))
	for _, e := range os.Environ() {
		parts := strings.SplitN(e, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.ToUpper(parts[0])
		blocked := false
		for _, prefix := range blockedEnvPrefixes {
			if strings.HasPrefix(key, prefix) {
				blocked = true
				break
			}
		}
		if !blocked {
			env = append(env, e)
		}
	}
	return env
}
