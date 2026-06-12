package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	configSiteFlag string
	configPathFlag string
)

var configImportCmd = &cobra.Command{
	Use:   "import",
	Short: "Import YAML config into the database",
	RunE: func(cmd *cobra.Command, args []string) error {
		if configSiteFlag == "" {
			return fmt.Errorf("--site flag is required")
		}
		if configPathFlag == "" {
			return fmt.Errorf("--path flag is required")
		}
		return runConfigImport(configSiteFlag, configPathFlag)
	},
}

func init() {
	configCmd.AddCommand(configImportCmd)
	configImportCmd.Flags().StringVar(&configSiteFlag, "site", "", "Target site hostname")
	configImportCmd.Flags().StringVar(&configPathFlag, "path", "", "Path to config directory")
}
