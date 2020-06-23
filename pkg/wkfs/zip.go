package wkfs

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/master-g/i18n/pkg/wkio"
)

// Unzip will decompress a zip archive, moving all files and folders
// within the zip file to an output directory
func Unzip(src string, dest string) (inflated []string, err error) {
	var r *zip.ReadCloser
	r, err = zip.OpenReader(src)
	if err != nil {
		return
	}
	defer wkio.SafeClose(r)

	for _, f := range r.File {
		// store filename/path for returning and using later on
		fpath := filepath.Join(dest, f.Name)

		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
			err = fmt.Errorf("%s: illegal file path", fpath)
			return
		}

		inflated = append(inflated, fpath)

		if f.FileInfo().IsDir() {
			//  make folder
			err = os.MkdirAll(fpath, os.ModePerm)
			if err != nil {
				return
			}
			continue
		}

		// make file
		if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return
		}

		var outFile *os.File
		outFile, err = os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return
		}

		var rc io.ReadCloser
		rc, err = f.Open()
		if err != nil {
			return
		}

		_, err = io.Copy(outFile, rc)
		wkio.SafeClose(outFile)
		wkio.SafeClose(rc)

		if err != nil {
			return
		}
	}

	return
}
