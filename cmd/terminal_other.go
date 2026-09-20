//go:build !windows

// terminal_other.go — no-op VT initialization for non-Windows systems.
package cmd

func initVirtualTerminal() bool {
	return true
}
