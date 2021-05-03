package tags

// Main specification tags in out configurator
const (
	prefixLibTag      = "cf_"
	TagHocon          = prefixLibTag + "hocon"
	TagJSON           = prefixLibTag + "json"
	TagYaml           = prefixLibTag + "yaml"
	TagIni            = prefixLibTag + "ini"
	TagToml           = prefixLibTag + "toml"
	TagDefault        = prefixLibTag + "default"
	TagEnvironment    = prefixLibTag + "env"
	TagServerKeyValue = prefixLibTag + "server_kv"
	TagServerFile     = prefixLibTag + "server_file"
	TagHashiCorpVault = prefixLibTag + "vault"
	TagPriority       = prefixLibTag + "priority"

	AmountTags = 11

	HoconFile = "GOSTRUCTOR_HOCON"
	YamlFile  = "GOSTRUCTOR_YAML"
	JSONFile  = "GOSTRUCTOR_JSON"

	Priority = "GOSTRUCTOR_PRIORITY"
)
