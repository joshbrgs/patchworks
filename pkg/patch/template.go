package patch

import (
	"bytes"
	"fmt"
	"text/template"
)

// RenderTemplate applies Go templating to the patch spec.
func RenderTemplate(tmpl string, data map[string]string) (string, error) {
	tpl, err := template.New("patch").Parse(tmpl)
	if err != nil {
		return "", fmt.Errorf("failed to parse: %w", err)
	}

	var rendered bytes.Buffer
	if err = tpl.Execute(&rendered, data); err != nil {
		return "", err
	}

	return rendered.String(), nil
}
