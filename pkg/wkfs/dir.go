package wkfs

import (
	"os"
)

// EnsureDir creates intermediate directories if needed
func EnsureDir(p string) (err error) {
	if _, err = os.Stat(p); err != nil {
		err = os.MkdirAll(p, os.ModePerm)
		if err != nil {
			return
		}
	}
	return
}

func IsDir(p string) bool {
	fi, err := os.Stat(p)
	if err != nil {
		return false
	}

	return fi.IsDir()
}
