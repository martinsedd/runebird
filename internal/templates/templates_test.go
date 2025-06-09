package templates

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"runebird/internal/config"
)

func TestTemplateManager(t *testing.T) {
	tmpDir := t.TempDir()

	testTemplates := map[string]string{
		"welcome.html": `
<html>
<body>
	<h1>Welcome, {{ .Name }}!</h1>
</body>
</html>
`,
		"notification.html": `
{{ define "subject" }}Notification for {{ .User }}{{ end }}
<html>
<body>
	<p>You have a new notification, {{ .User }}.</p>
</body>
</html>
`,
	}

	for name, content := range testTemplates {
		err := os.WriteFile(filepath.Join(tmpDir, name), []byte(content), 0644)
		if err != nil {
			t.Fatalf("failed to write test template %s: %v", name, err)
		}
	}

	cfg := &config.TemplatesConfig{
		Path: tmpDir,
	}

	t.Run("LoadTemplates", func(t *testing.T) {
		tm, err := New(cfg)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		names := tm.ListTemplates()
		if len(names) != 2 {
			t.Errorf("expected 2 templates, got: %d", len(names))
		}
	})

	t.Run("RenderTemplateWithoutSubject", func(t *testing.T) {
		tm, err := New(cfg)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		data := map[string]string{"Name": "Alice"}
		body, subject, err := tm.Render("welcome", data)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		// Debug output
		t.Logf("Rendered body: %q", body)

		trimmedBody := strings.TrimSpace(body)
		if !strings.Contains(trimmedBody, "Welcome, Alice!") {
			t.Errorf("expected body to contain 'Welcome, Alice!', got: %s", body)
		}
		if subject != "" {
			t.Errorf("expected empty subject, got: %s", subject)
		}
	})

	t.Run("RenderTemplateWithSubject", func(t *testing.T) {
		tm, err := New(cfg)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		data := map[string]string{"User": "Bob"}
		body, subject, err := tm.Render("notification", data)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		// Debug output
		t.Logf("Rendered body: %q", body)
		t.Logf("Rendered subject: %q", subject)

		trimmedBody := strings.TrimSpace(body)
		if !strings.Contains(trimmedBody, "You have a new notification, Bob.") {
			t.Errorf("expected body to contain notification message, got: %s", body)
		}
		if subject != "Notification for Bob" {
			t.Errorf("expected subject 'Notification for Bob', got: %s", subject)
		}
	})

	t.Run("RenderNonExistentTemplate", func(t *testing.T) {
		tm, err := New(cfg)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		_, _, err = tm.Render("nonexistent", nil)
		if err == nil {
			t.Fatal("expected error for nonexistent template, got none")
		}
	})

	t.Run("EmptyTemplateDirectory", func(t *testing.T) {
		emptyDir := t.TempDir()
		emptyCfg := &config.TemplatesConfig{
			Path: emptyDir,
		}
		_, err := New(emptyCfg)
		if err == nil {
			t.Fatal("expected error for empty template directory, got none")
		}
	})
}
