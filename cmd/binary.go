// binary.go — implements the `movie binary` command to output the GitMap identity block.
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var binaryCmd = &cobra.Command{
	Use:     "binary",
	Aliases: []string{"bin"},
	Short:   "Display movie binary identity and runtime environment metadata",
	Long: `Display movie binary identity and runtime environment metadata.

Outputs the name, Git repository URL, version, commit SHA, primary database path,
installed binary path, and build date following the GitMap terminal design system.`,
	Run: func(cmd *cobra.Command, args []string) {
		isColor := isColorEnabled()
		renderMovieIdentityBlock(os.Stdout, isColor)
	},
}
