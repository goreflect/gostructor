package tags

type token int

const (
	wrong token = iota
	eof

	value       // value which recognized by name field in environment or any other sources
	customParam // node, path

	// MISC characters
	comma
)
