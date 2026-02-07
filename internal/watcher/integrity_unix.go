//go:build !windows

package watcher

import (
	"os"
	"syscall"
)

// getInode returns the inode number of the file
func getInode(info os.FileInfo) uint64 {
	if stat, ok := info.Sys().(*syscall.Stat_t); ok {
		return stat.Ino
	}
	return 0
}
