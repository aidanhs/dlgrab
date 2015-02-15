package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func writeFile(path string, r io.ReadCloser) (err error) {
	logger.Info("Writing file: ", filepath.Base(path))
	err = os.MkdirAll(filepath.Dir(path), 0755)
	if err != nil {
		return
	}
	out, err := os.Create(path)
	if err != nil {
		return
	}
	defer out.Close()
	cnt, err := io.Copy(out, r)
	if err != nil {
		return
	}
	logger.Debug(fmt.Sprintf("Wrote %d bytes", cnt))
	return
}
