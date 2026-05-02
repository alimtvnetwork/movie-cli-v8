// self_replace.go — `movie self-replace` command.
// One-shot bootstrap that atomically copies the freshly deployed binary over
// the active PATH binary, breaking out of stale-handoff loops where the active
// binary points at a different drive than powershell.json's deployPath.
//
// Usage:
//
//	movie self-replace                           # auto: from deployPath, to active PATH
//	movie self-replace --from D:\bin-run\movie.exe
//	movie self-replace --from <src> --to <dst>
//
// See spec/09-app-issues/07-updater-deploypath-vs-path-mismatch.md.
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/alimtvnetwork/movie-cli-v8/updater"
)

var (
	selfReplaceFrom string
	selfReplaceTo   string
)

var selfReplaceCmd = &cobra.Command{
	Use:   "self-replace",
	Short: "Atomically replace the active PATH binary with the freshly deployed one",
	Long: `Atomically copies a deployed binary over the active PATH 'movie',
using rename-first semantics so it works on Windows even while the binary
is loaded by another process.

Defaults:
  --from   deployPath + binaryName from powershell.json (e.g. D:\bin-run\movie.exe)
  --to     the active 'movie' resolved from $PATH (e.g. E:\bin-run\movie.exe)

This is a one-shot bootstrap to break out of a stuck-update loop where the
deploy target and the PATH-resolved binary live in different directories,
leaving the active binary frozen on an old version forever.

Examples:
  movie self-replace
  movie self-replace --from D:\bin-run\movie.exe
  movie self-replace --from D:\bin-run\movie.exe --to E:\bin-run\movie.exe`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("self-replace: bootstrapping active PATH binary")
		if err := updater.SelfReplace(selfReplaceFrom, selfReplaceTo); err != nil {
			fmt.Fprintf(os.Stderr, "self-replace failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("self-replace: done")
	},
}

func init() {
	selfReplaceCmd.Flags().StringVar(&selfReplaceFrom, "from", "", "Source binary (default: deployPath/binaryName from powershell.json)")
	selfReplaceCmd.Flags().StringVar(&selfReplaceTo, "to", "", "Target binary to replace (default: active 'movie' on PATH)")
}
