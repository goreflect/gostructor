package tools

import (
	"errors"
	"strings"
)

type (
	// ConfigureConverts - configuration structure for converters tools
	ConfigureConverts struct {
		Separator SeparatorType
	}

	// SeparatorType - type for configuring separators
	SeparatorType string
)

const (
	// COMMA - separator in value is ,
	COMMA SeparatorType = ","
)

// ConvertStringIntoArray - converting from string into array of string
func ConvertStringIntoArray(value string, configuration ConfigureConverts) ([]string, error) {
	if err := configuration.validation(); err != nil {
		return nil, err
	}
	return strings.Split(value, string(configuration.Separator)), nil
}

func (configure ConfigureConverts) validation() error {
	if configure.Separator == "" {
		return errors.New("separator can not be empty! ")
	}
	return nil
}
