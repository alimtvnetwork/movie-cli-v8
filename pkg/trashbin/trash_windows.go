//go:build windows

package trashbin

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"unsafe"

	"github.com/alimtvnetwork/movie-cli-v8/pkg/appfault"
)

const (
	foDelete          = 0x0003
	fofAllowUndo      = 0x0040
	fofNoConfirmation = 0x0010
	fofSilent         = 0x0004
	fofNoErrorUI      = 0x0400
)

type shFileOpStructW struct {
	hwnd                  uintptr
	wFunc                 uint32
	pFrom                 *uint16
	pTo                   *uint16
	fFlags                uint16
	fAnyOperationsAborted int32
	hNameMappings         uintptr
	lpszProgressTitle     *uint16
}

var (
	shell32          = syscall.NewLazyDLL("shell32.dll")
	shFileOperationW = shell32.NewProc("SHFileOperationW")
)

func moveToTrashOS(absPath string) error {
	fromWide, err := syscall.UTF16FromString(absPath)
	if err == nil {
		fromWide = append(fromWide, 0)
		op := shFileOpStructW{
			wFunc:  foDelete,
			pFrom:  &fromWide[0],
			fFlags: fofAllowUndo | fofNoConfirmation | fofSilent | fofNoErrorUI,
		}

		ret, _, _ := shFileOperationW.Call(uintptr(unsafe.Pointer(&op)))
		if ret == 0 && op.fAnyOperationsAborted == 0 {
			return nil
		}
	}

	return moveToRecycleBinPowerShell(absPath)
}

func moveToRecycleBinPowerShell(absPath string) error {
	info, err := os.Stat(absPath)
	if err != nil {
		return appfault.Wrapf(err, "stat path %s", absPath)
	}

	escaped := strings.ReplaceAll(absPath, "'", "''")
	var psCmd string
	if info.IsDir() {
		psCmd = fmt.Sprintf("Add-Type -AssemblyName Microsoft.VisualBasic; [Microsoft.VisualBasic.FileIO.FileSystem]::DeleteDirectory('%s', 'OnlyErrorDialogs', 'SendToRecycleBin')", escaped)
	} else {
		psCmd = fmt.Sprintf("Add-Type -AssemblyName Microsoft.VisualBasic; [Microsoft.VisualBasic.FileIO.FileSystem]::DeleteFile('%s', 'OnlyErrorDialogs', 'SendToRecycleBin')", escaped)
	}

	cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", psCmd)
	out, runErr := cmd.CombinedOutput()
	if runErr != nil {
		return appfault.Wrapf(runErr, "powershell recycle bin failed: %s (output: %s)", absPath, string(out))
	}

	return nil
}
