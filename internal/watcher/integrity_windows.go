//go:build windows

package watcher

import (
	"os"
)

// getInode returns a file identifier on Windows
// Windows doesn't have inodes, so we use a hash of the path and mod time
func getInode(info os.FileInfo) uint64 {
	// On Windows, we can't easily get a stable file ID
	// Use modification time as a proxy to detect file replacement
	return uint64(info.ModTime().UnixNano())
}
