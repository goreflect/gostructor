package tools

import (
	"bytes"
	"os"
)

// ReadFromFile reads fileName into a byte buffer.
func ReadFromFile(fileName string) (*bytes.Buffer, error) {
	bts, err := os.ReadFile(fileName)
	if err != nil {
		return nil, err
	}
	byteBuffer := bytes.NewBuffer(bts)
	return byteBuffer, nil
}
