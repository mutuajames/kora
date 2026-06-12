package doctype

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// ParseFile reads a single YAML/JSON file and returns a DocType.
func ParseFile(path string) (*DocType, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading file %s: %w", path, err)
	}

	dt := &DocType{}
	if err := yaml.Unmarshal(data, dt); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", path, err)
	}

	if err := dt.Validate(); err != nil {
		return nil, fmt.Errorf("validating %s: %w", path, err)
	}

	return dt, nil
}

// ParseDirectory reads all YAML/JSON files in a directory and returns the DocTypes found.
func ParseDirectory(path string) ([]*DocType, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("reading directory %s: %w", path, err)
	}

	var doctypes []*DocType
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(entry.Name()))
		if ext != ".yaml" && ext != ".yml" && ext != ".json" {
			continue
		}

		filePath := filepath.Join(path, entry.Name())

		// Skip workflow files (they have "document_type" and "states", not "module" and "fields").
		if isWorkflowFile(filePath) {
			continue
		}

		dt, err := ParseFile(filePath)
		if err != nil {
			return nil, err
		}
		doctypes = append(doctypes, dt)
	}

	return doctypes, nil
}

// isWorkflowFile checks if a YAML file is a workflow definition by peeking at its top-level keys.
func isWorkflowFile(path string) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	// Quick check: workflow files have "document_type" and "states" at top level.
	// DocType files have "module" and "fields".
	hasDocType := containsKey(data, "document_type")
	hasStates := containsKey(data, "states")
	hasModule := containsKey(data, "module")
	hasFields := containsKey(data, "fields")

	// If it has document_type and states but NOT module and fields, it's a workflow.
	if hasDocType && hasStates && !hasModule && !hasFields {
		return true
	}
	return false
}

// containsKey does a simple string check for a top-level YAML key.
func containsKey(data []byte, key string) bool {
	// Look for "key:" at the start of a line.
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, key+":") || strings.HasPrefix(trimmed, key+" :") {
			return true
		}
	}
	return false
}

// ParseConfigTree reads the full config directory structure:
//
//	config/<app>/
//	  app.yaml
//	  roles.yaml
//	  permissions.yaml
//	  scheduler.yaml
//	  doctypes/
//	    *.yaml
//
// Returns DocTypes, Roles, Permissions, and other configs separately.
func ParseConfigTree(basePath string) ([]*DocType, error) {
	// Check for doctypes/ subdirectory.
	doctypesPath := filepath.Join(basePath, "doctypes")
	if _, err := os.Stat(doctypesPath); err == nil {
		return ParseDirectory(doctypesPath)
	}

	// Otherwise parse the directory directly.
	return ParseDirectory(basePath)
}

// Validate checks that the DocType definition is structurally valid.
func (d *DocType) Validate() error {
	if d.Name == "" {
		return fmt.Errorf("doctype name is required")
	}

	if d.Module == "" {
		return fmt.Errorf("doctype %s: module is required", d.Name)
	}

	// Set defaults.
	if d.SortField == "" {
		d.SortField = "modified"
	}
	if d.SortOrder == "" {
		d.SortOrder = "DESC"
	}
	if d.TitleField == "" {
		d.TitleField = "name"
	}

	fieldnames := make(map[string]bool)
	for i := range d.Fields {
		f := &d.Fields[i]

		if f.Fieldname == "" {
			return fmt.Errorf("doctype %s: field %d has no fieldname", d.Name, i)
		}
		if fieldnames[f.Fieldname] {
			return fmt.Errorf("doctype %s: duplicate fieldname %q", d.Name, f.Fieldname)
		}
		fieldnames[f.Fieldname] = true

		if f.Fieldtype == "" {
			return fmt.Errorf("doctype %s: field %s has no fieldtype", d.Name, f.Fieldname)
		}

		// Validate field type.
		if err := validateFieldType(f.Fieldtype); err != nil {
			return fmt.Errorf("doctype %s, field %s: %w", d.Name, f.Fieldname, err)
		}

		// Set default label.
		if f.Label == "" {
			f.Label = fieldnameToLabel(f.Fieldname)
		}

		// Validate constraints.
		for j, c := range f.Constraints {
			if c.Type == "" {
				return fmt.Errorf("doctype %s, field %s: constraint %d has no type", d.Name, f.Fieldname, j)
			}
			if c.Message == "" {
				return fmt.Errorf("doctype %s, field %s: constraint %d has no message", d.Name, f.Fieldname, j)
			}
		}

		// Validate Table field has options (the child DocType name).
		if f.Fieldtype == "Table" && f.Options == "" {
			return fmt.Errorf("doctype %s, field %s: Table field requires options (child DocType name)", d.Name, f.Fieldname)
		}

		// Validate Link field has options (the target DocType name).
		if f.Fieldtype == "Link" && f.Options == "" {
			return fmt.Errorf("doctype %s, field %s: Link field requires options (target DocType name)", d.Name, f.Fieldname)
		}
	}

	return nil
}

func validateFieldType(ft string) error {
	validTypes := map[string]bool{
		"Data": true, "Text": true, "Text Editor": true,
		"Int": true, "Float": true, "Currency": true, "Percent": true,
		"Check": true, "Date": true, "Time": true, "Datetime": true,
		"Select": true, "Link": true, "Dynamic Link": true,
		"Table": true, "Attach": true, "Attach Image": true,
		"JSON": true, "Password": true,
		"Section Break": true, "Column Break": true, "Heading": true,
	}
	if !validTypes[ft] {
		return fmt.Errorf("unknown fieldtype %q", ft)
	}
	return nil
}

func fieldnameToLabel(name string) string {
	// Convert snake_case to Title Case.
	parts := strings.Split(name, "_")
	for i, p := range parts {
		if len(p) > 0 {
			parts[i] = strings.ToUpper(p[:1]) + p[1:]
		}
	}
	return strings.Join(parts, " ")
}
