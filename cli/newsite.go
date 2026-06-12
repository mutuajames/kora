package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/spf13/cobra"
)

var newSiteCmd = &cobra.Command{
	Use:   "new-site <hostname>",
	Short: "Create a new site and database",
	Long: `Create a new site directory with a site_config.yaml template.
The database must already exist.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		hostname := args[0]

		sitesDir := "sites"
		siteDir := filepath.Join(sitesDir, hostname)

		// Create site directory if it doesn't exist.
		if err := os.MkdirAll(siteDir, 0755); err != nil {
			return fmt.Errorf("creating site directory: %w", err)
		}

		// Write site_config.yaml template.
		configPath := filepath.Join(siteDir, "site_config.yaml")
		if _, err := os.Stat(configPath); err == nil {
			return fmt.Errorf("site config already exists at %s", configPath)
		}

		tmpl := template.Must(template.New("site_config").Parse(siteConfigTemplate))
		f, err := os.Create(configPath)
		if err != nil {
			return fmt.Errorf("creating site config: %w", err)
		}
		defer f.Close()

		data := map[string]string{
			"Hostname": hostname,
			"DBName":   hostnameToDBName(hostname),
		}
		if err := tmpl.Execute(f, data); err != nil {
			return fmt.Errorf("writing site config: %w", err)
		}

		// Create files directory.
		filesDir := filepath.Join(siteDir, "files")
		if err := os.MkdirAll(filesDir, 0755); err != nil {
			return fmt.Errorf("creating files directory: %w", err)
		}

		fmt.Printf("Site created at %s\n", siteDir)
		fmt.Printf("Edit %s to configure database credentials.\n", configPath)
		return nil
	},
}

func hostnameToDBName(hostname string) string {
	// Replace dots with underscores for a valid MySQL database name.
	name := ""
	for _, c := range hostname {
		if c == '.' {
			name += "_"
		} else {
			name += string(c)
		}
	}
	return name
}

var siteConfigTemplate = `# Site configuration for {{.Hostname}}
db_host: 127.0.0.1
db_port: 3306
db_name: {{.DBName}}
db_user: kora
db_password: changeme

redis_url: redis://localhost:6379/0

file_storage: local
files_path: sites/{{.Hostname}}/files

apps:
  - core

hostname: {{.Hostname}}
`
