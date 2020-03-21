package pipeline

import (
	"fmt"
	"os"
)

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
