//go:build !windows && !darwin && !linux

package trashbin

func moveToTrashOS(absPath string) error {
	return fallbackQuarantine(absPath)
}
