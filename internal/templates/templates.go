package templates

import (
	"bytes"
	"embed"
	"text/template"
)

//go:embed prime/*.tmpl
var primeTemplates embed.FS

// PrimeData holds all data needed for prime templates
type PrimeData struct {
	Location     string
	Role         string
	Grove        *GroveData
	Mission      *MissionData
	CurrentEpic  *EpicData
	Epics        []*EpicWithTasks
}

// GroveData represents grove information for templates
type GroveData struct {
	ID        string
	Name      string
	MissionID string
}

// MissionData represents mission information for templates
type MissionData struct {
	ID          string
	Title       string
	Workspace   string
	Description string
}

// EpicData represents epic information for templates
type EpicData struct {
	ID          string
	Title       string
	Status      string
	Description string
}

// EpicWithTasks represents an epic with its ready tasks
type EpicWithTasks struct {
	ID          string
	Title       string
	Status      string
	Description string
	Tasks       []*TaskData
}

// TaskData represents task information for templates
type TaskData struct {
	ID     string
	Title  string
	Status string
}

// RenderPrime renders a prime template with the given data
func RenderPrime(templateName string, data *PrimeData) (string, error) {
	// Create template with custom functions
	funcMap := template.FuncMap{
		"add": func(a, b int) int {
			return a + b
		},
	}

	// Parse the main template and all partials
	tmpl, err := template.New(templateName).Funcs(funcMap).ParseFS(
		primeTemplates,
		"prime/"+templateName,
		"prime/core-rules.tmpl",
		"prime/git-discovery.tmpl",
		"prime/welcome-orc.tmpl",
		"prime/welcome-imp.tmpl",
	)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// GetCoreRules returns the core rules template content
func GetCoreRules() (string, error) {
	content, err := primeTemplates.ReadFile("prime/core-rules.tmpl")
	if err != nil {
		return "", err
	}
	return string(content), nil
}

// GetGitDiscovery returns the git discovery template content
func GetGitDiscovery() (string, error) {
	content, err := primeTemplates.ReadFile("prime/git-discovery.tmpl")
	if err != nil {
		return "", err
	}
	return string(content), nil
}

// GetWelcomeORC returns the ORC welcome message template content
func GetWelcomeORC() (string, error) {
	content, err := primeTemplates.ReadFile("prime/welcome-orc.tmpl")
	if err != nil {
		return "", err
	}
	return string(content), nil
}

// GetWelcomeIMP returns the IMP welcome message template content
func GetWelcomeIMP() (string, error) {
	content, err := primeTemplates.ReadFile("prime/welcome-imp.tmpl")
	if err != nil {
		return "", err
	}
	return string(content), nil
}
