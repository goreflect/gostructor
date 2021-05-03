package extracting

// TODO: In future for v1+
type ExtractorSource interface {
	readSource() map[string]interface{}
}
