package pipeline

import (
	"fmt"
	"os"
)

// author: artemkaxboy
func checkFileAccessibility(filename string) error {
	info, err := os.Stat(filename)
	if err != nil {
		return err
	}
	if info.Mode()&(1<<8) == 0 {
		return fmt.Errorf("%s permission denied", filename)
	}
	if info.IsDir() {
		return fmt.Errorf("%s is a directory", filename)
	}
	return nil
}

// // get type configuration file by name: configuration file can be ended by .hocon, .json, .yml, .yaml. In version > 1.0 this will be moved into another lib
// func getTypeConfigurationFile(fileName string) error {
// 	return nil
// }
