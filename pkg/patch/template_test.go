package patch

import (
	"testing"
)

// TestRenderTemplate tests the RenderTemplate function
func TestRenderTemplate(t *testing.T) {
	tests := []struct {
		name     string
		tmpl     string
		data     map[string]string
		expected string
		wantErr  bool
	}{
		{
			name:     "Valid Template",
			tmpl:     "Hello, {{.Name}}!",
			data:     map[string]string{"Name": "World"},
			expected: "Hello, World!",
			wantErr:  false,
		},
		{
			name:     "Missing Key",
			tmpl:     "Hello, {{.Name}}!",
			data:     map[string]string{},  // Missing "Name"
			expected: "Hello, <no value>!", // Go templates don't error on missing keys they express no value
			wantErr:  false,
		},
		{
			name:     "Invalid Template Syntax",
			tmpl:     "Hello, {{.Name", // Missing closing `}}`
			data:     map[string]string{"Name": "World"},
			expected: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := RenderTemplate(tt.tmpl, tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("RenderTemplate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if result != tt.expected {
				t.Errorf("RenderTemplate() = %q, expected %q", result, tt.expected)
			}
		})
	}
}
