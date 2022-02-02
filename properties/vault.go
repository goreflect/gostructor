package properties

type VaultConfiguration struct {
	VaultAddress string `cf_env:"VAULT_URL"`
	VaultToken   string `cf_env:"VAULT_TOKEN"`
}
