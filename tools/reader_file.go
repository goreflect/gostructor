package tools

import (
	"bytes"
	"io/ioutil"
)

//ReadFromFile - read from file to byte buffer
func ReadFromFile(fileName string) (*bytes.Buffer, error) {
	bts, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, err
	}
	byteBuffer := bytes.NewBuffer(bts)
	return byteBuffer, nil
}
