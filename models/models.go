package models

import (
	"encoding/json"
	"io"
	"os"
)

func ParseModelWithUnmarshal[T any](model *T, filePath string) (err error) {
	file, err := os.Open(filePath)
	if err != nil {
		return
	}
	defer file.Close()
	bytes, err := io.ReadAll(file)
	if err != nil {
		return
	}
	if err = json.Unmarshal(bytes, model); err != nil {
		return
	}
	return
}

func ParseModelWithDecoder[T any](model *T, filePath string) (err error) {
	file, err := os.Open(filePath)
	if err != nil {
		return
	}
	defer file.Close()
	if err = json.NewDecoder(file).Decode(model); err != nil {
		return
	}
	return
}
