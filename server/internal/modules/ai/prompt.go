package ai

import (
	"bytes"
	"embed"
	"fmt"
	"strings"
	"text/template"
)

//go:embed prompts
var FS embed.FS

var prompts *template.Template

const (
	ChatPrompt  = "chat"
	TitlePrompt = "title"
)

// Parses all templates at once
func LoadPrompts() error {
	var err error
	prompts, err = template.ParseFS(FS, "prompts/*.tmpl")
	if err != nil {
		return fmt.Errorf("load prompt templates: %w", err)
	}
	return nil
}

// Fill a template with passed data
func RenderPrompt(name string, data any) (string, error) {
	var promptBuf bytes.Buffer
	if err := prompts.ExecuteTemplate(&promptBuf, name, data); err != nil {
		return "", fmt.Errorf("render Prompt %q: %w", name, err)
	}
	return strings.TrimSpace(promptBuf.String()), nil
}
