package tools

import (
	"bytes"
	"io/ioutil"
)

// ReadFromFile reads fileName into a byte buffer.
func ReadFromFile(fileName string) (*bytes.Buffer, error) {
	bts, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, err
	}
	byteBuffer := bytes.NewBuffer(bts)
	return byteBuffer, nil
}
