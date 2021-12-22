package web

import (
	"fmt"
	"os"
	"path/filepath"
)

// DeleteMediaObject purges object on Media
func (app *WebApp) DeleteMediaObject(fileName string) error {
	fPath := filepath.Join(app.getMediaPath(), "original", fileName)
	err := os.Remove(fPath) // remove a single file
	if err != nil {
		return err
	}
	// smaller file will be deleted when they are expired.
	return nil
}

func (app *WebApp) getMediaPath() string {
	return filepath.Join(app.basePath, "media")
}

// DoesThisMediaExist tells whether the requested file exist in storage or not
func (app *WebApp) DoesThisMediaExist(fileName string) bool {
	fPath := filepath.Join(app.getMediaPath(), "original", fileName)
	_, err := os.Stat(fPath)
	return !os.IsNotExist(err)
}

// UploadToMedia to upload file to Media and return with output
func (app *WebApp) UploadToMedia(buff []byte, fileName string) (string, error) {
	fPath := filepath.Join(app.getMediaPath(), "original", fileName)
	dir := filepath.Dir(fPath)
	_, err := CheckOrCreateDir(dir)
	if err != nil {
		return "", err
	}
	file, err := os.Create(fPath)
	if err != nil {
		errMsg := fmt.Sprintf("Unable to open file %q", fPath)
		return errMsg, err
	}
	file.Write(buff)
	defer file.Close()
	return fPath, nil
}
