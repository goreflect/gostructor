package tools

import "strings"

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
func ConvertStringIntoArray(value string, configuration ConfigureConverts) []string {
	return strings.Split(value, string(configuration.Separator))
}
