package wkfs

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"
)

type scanMode int

const (
	scanModeFilesOnly scanMode = iota
	scanModeFoldersOnly
	scanModeBoth
)

type ScanOpt func(options *ScanOptions)

// ScanOptions holds folder scan options
type ScanOptions struct {
	patterns []*regexp.Regexp
	types    []string
	mode     scanMode
	// TODO
	// IgnorePrefix []string
	// IgnoreType []string
}

// WithFilesOnly specifies only scan for files
func WithFilesOnly() ScanOpt {
	return func(op *ScanOptions) {
		op.mode = scanModeFilesOnly
	}
}

// WithFoldersOnly specifies only scan for folders
func WithFoldersOnly() ScanOpt {
	return func(op *ScanOptions) {
		op.mode = scanModeFoldersOnly
	}
}

// WithBothFileAndFolder specifies scan both files and folders
func WithBothFileAndFolder() ScanOpt {
	return func(op *ScanOptions) {
		op.mode = scanModeBoth
	}
}

// WithPatterns specifies regex pattern to scan for
func WithPatterns(patterns ...string) ScanOpt {
	return func(op *ScanOptions) {
		for _, raw := range patterns {
			r, err := regexp.Compile(raw)
			if err == nil {
				op.patterns = append(op.patterns, r)
			} else {
				log.Debugf("unable to compile regexp: %s, err: %s", raw, err)
			}
		}
	}
}

// WithTypes specifies file types to scan for
func WithTypes(types ...string) ScanOpt {
	return func(op *ScanOptions) {
		op.types = append(op.types, types...)
	}
}

func applyScanOptions(opts []ScanOpt) *ScanOptions {
	options := &ScanOptions{
		patterns: nil,
		types:    []string{".txt"},
		mode:     scanModeFilesOnly,
	}

	for _, opt := range opts {
		opt(options)
	}

	return options
}

// Scan folder recursive, find all the files match with scan options
func Scan(dir string, opts ...ScanOpt) (files, folders []string, err error) {
	options := applyScanOptions(opts)

	err = filepath.Walk(dir, func(path string, f os.FileInfo, e error) error {
		continueFlag := false

		f, e = os.Stat(path)
		if e != nil {
			return e
		}

		fm := f.Mode()

		// scan mode
		if options.mode == scanModeFilesOnly && !fm.IsRegular() {
			return nil
		} else if options.mode == scanModeFoldersOnly && !fm.IsDir() {
			return nil
		}

		if !fm.IsRegular() && !fm.IsDir() {
			return nil
		}

		// only file needs type checking
		if fm.IsRegular() && len(options.types) > 0 {
			// check type first
			ok := false
			ext := filepath.Ext(path)
			for _, t := range options.types {
				if strings.Index(ext, t) != -1 {
					ok = true
					break
				}
			}

			if !ok {
				return nil
			}
		}

		// pattern match
		if len(options.patterns) > 0 {
			baseName := filepath.Base(path)
			for _, reg := range options.patterns {
				if reg.MatchString(baseName) {
					continueFlag = true
					break
				}
			}
		} else {
			continueFlag = true
		}

		if continueFlag {
			if fm.IsDir() {
				folders = append(folders, path)
			} else if fm.IsRegular() {
				files = append(files, path)
			}
		}

		return e
	})

	return
}
