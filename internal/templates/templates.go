package templates

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"path/filepath"

	"runebird/internal/config"
)

type TemplateManager struct {
	Templates map[string]*template.Template
}

func New(cfg *config.TemplatesConfig) (*TemplateManager, error) {
	tm := &TemplateManager{
		Templates: make(map[string]*template.Template),
	}

	err := filepath.Walk(cfg.Path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || filepath.Ext(path) != ".html" {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read template file %s: %v", path, err)
		}

		name := filepath.Base(path[:len(path)-len(".html")])
		tmpl, err := template.New(name).Parse(string(content))
		if err != nil {
			return fmt.Errorf("failed to parse template %s: %v", name, err)
		}

		tm.Templates[name] = tmpl
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to load templates from %s: %v", cfg.Path, err)
	}

	if len(tm.Templates) == 0 {
		return nil, fmt.Errorf("no templates found in directory %s", cfg.Path)
	}

	return tm, nil
}

func (tm *TemplateManager) Render(name string, data interface{}) (body string, subject string, err error) {
	tmpl, ok := tm.Templates[name]
	if !ok {
		return "", "", fmt.Errorf("template %s not found", name)
	}

	var bodyBuf bytes.Buffer
	if err := tmpl.Execute(&bodyBuf, data); err != nil {
		return "", "", fmt.Errorf("failed to render template %s: %v", name, err)
	}
	body = bodyBuf.String()

	subjectTmpl := tmpl.Lookup("subject")
	if subjectTmpl != nil {
		var subjectBuf bytes.Buffer
		if err := subjectTmpl.Execute(&subjectBuf, data); err != nil {
			return "", "", fmt.Errorf("failed to render subject for template %s: %v", name, err)
		}
		subject = subjectBuf.String()
	}

	return body, subject, nil
}

func (tm *TemplateManager) ListTemplates() []string {
	names := make([]string, 0, len(tm.Templates))
	for name := range tm.Templates {
		names = append(names, name)
	}
	return names
}
