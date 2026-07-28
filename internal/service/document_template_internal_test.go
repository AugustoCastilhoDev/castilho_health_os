package service

import "testing"

func TestResolvePlaceholders(t *testing.T) {
	content := "Atesto que {{patient_name}} esteve em consulta com {{professional_name}} em {{date}}. CID: {{cid}}"
	vars := map[string]string{
		"patient_name":      "Mariana Costa",
		"professional_name": "Dra. Ana Souza",
		"date":              "28/07/2026",
	}

	got := resolvePlaceholders(content, vars)

	want := "Atesto que Mariana Costa esteve em consulta com Dra. Ana Souza em 28/07/2026. CID: {{cid}}"
	if got != want {
		t.Errorf("resolvePlaceholders() = %q, want %q", got, want)
	}
}
