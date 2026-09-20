//go:build windows

// terminal_windows.go — enables Virtual Terminal Processing for Windows console.
package cmd

import (
	"os"
	"sync"

	"golang.org/x/sys/windows"
)

var (
	vtOnce        sync.Once
	isVtSupported bool
)

func initVirtualTerminal() bool {
	vtOnce.Do(func() {
		if os.Getenv("WT_SESSION") != "" {
			isVtSupported = true

			return
		}

		stdoutHandle := windows.Handle(os.Stdout.Fd())

		var mode uint32
		if err := windows.GetConsoleMode(stdoutHandle, &mode); err != nil {
			isVtSupported = false

			return
		}

		if mode&windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING != 0 {
			isVtSupported = true

			return
		}

		newMode := mode | windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING
		if err := windows.SetConsoleMode(stdoutHandle, newMode); err != nil {
			isVtSupported = false

			return
		}

		stderrHandle := windows.Handle(os.Stderr.Fd())

		var errMode uint32
		if err := windows.GetConsoleMode(stderrHandle, &errMode); err == nil {
			_ = windows.SetConsoleMode(stderrHandle, errMode|windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING)
		}

		isVtSupported = true
	})

	return isVtSupported
}
