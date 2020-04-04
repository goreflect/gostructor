package tags

// Main specification tags in out configurator
const (
	prefixLibTag      = "cf_"
	TagHocon          = prefixLibTag + "hocon"
	TagJSON           = prefixLibTag + "json"
	TagYaml           = prefixLibTag + "yaml"
	TagDefault        = prefixLibTag + "default"
	TagEnvironment    = prefixLibTag + "env"
	TagConfigServer   = prefixLibTag + "server"
	TagHashiCorpVault = prefixLibTag + "vault"

	AmountTags = 8

	TagCustomerNode = "node"
	TagCustomerPath = "path"
)
