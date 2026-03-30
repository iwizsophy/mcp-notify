//go:build windows

package player

import "syscall"

const (
	createNewProcessGroup = 0x00000200
	createNoWindow        = 0x08000000
	detachedProcess       = 0x00000008
)

func detachedProcessAttributes() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{
		CreationFlags: detachedProcess | createNewProcessGroup | createNoWindow,
	}
}
