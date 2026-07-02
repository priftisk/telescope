package internal

import (
	"os"
	"path/filepath"
)

func GetStaticDir() (string, error) {
	wd, err := os.Getwd()
	if err != nil {

		return "", err
	}
	static := filepath.Join(wd, "static")
	return static, nil
}
