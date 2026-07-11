package pipeline

import (
	"errors"
	"reflect"
	"strings"

	"github.com/goreflect/gostructor/converters"
	"github.com/goreflect/gostructor/infra"
	"github.com/goreflect/gostructor/properties"
	"github.com/goreflect/gostructor/tags"
	vault "github.com/mittwald/vaultgo"
	"github.com/sirupsen/logrus"
)

/*VaultConfig - source vault configuring*/
type VaultConfig struct {
	Config     *properties.VaultConfiguration
	connection *vault.Client
}

func (config *VaultConfig) configureVault() error {
	configured, errConfigure := Configure(&properties.VaultConfiguration{}, []infra.FuncType{infra.FunctionSetupEnvironment}, "", true)
	if errConfigure != nil {
		logrus.Error("Can not initialize vault properties. Please setup VAULT_ADDRESS & VAULT_TOKEN for working with cf_vault")
		return errConfigure
	}
	config.Config = configured.(*properties.VaultConfiguration)
	return nil
}

func (config *VaultConfig) connect() error {
	conn, errConnection := vault.NewClient(config.Config.VaultAddress,
		vault.WithCaPath(""),
		vault.WithAuthToken(config.Config.VaultToken))
	if errConnection != nil {
		return errConnection
	}
	conn.SetToken(config.Config.VaultToken)
	config.connection = conn
	return nil
}

func (config *VaultConfig) vaultAvailable() error {
	if config.Config == nil {
		if err := config.configureVault(); err != nil {
			logrus.Error("Configure Vault Error: ", err)
			return err
		}
	}

	if config.connection == nil {
		return config.connect()
	}
	return nil
}

func (config *VaultConfig) prepareLayer(context *structContext) error {
	if errConn := config.vaultAvailable(); errConn != nil {
		logrus.Error("Error while connect to vault: ", errConn)
		return errConn
	}
	return nil
}

// parseTag splits a cf_vault tag ("secret/path#key") into a vault path and
// secret key, returning an error instead of panicking on a malformed tag.
func parseVaultTag(nameField string) (path string, secretName string, err error) {
	parts := strings.SplitN(nameField, "#", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", errors.New("cf_vault tag '" + nameField + "' is malformed, expected format 'path/to/secret#key'")
	}
	return parts[0], parts[1], nil
}

func (config VaultConfig) readSecret(context *structContext, nameField string) (interface{}, error) {
	path, secretName, err := parseVaultTag(nameField)
	if err != nil {
		return nil, err
	}
	secret, err := config.connection.Logical().Read(path)
	if err != nil {
		logrus.Error("Error while reading config from vault: ", err)
		return nil, err
	}
	if secret == nil {
		return nil, errors.New("no secret found in vault at path '" + path + "'")
	}
	secretValue, found := secret.Data[secretName]
	if !found {
		return nil, errors.New("key '" + secretName + "' was not found in vault secret at path '" + path + "'")
	}
	logrus.Debug("Secret Vault: ", secretValue)
	return secretValue, nil
}

func (config VaultConfig) GetBaseType(context *structContext) infra.GoStructorValue {
	if err := config.prepareLayer(context); err != nil {
		return infra.NewGoStructorNoValue(context.Value, err)
	}
	nameField := context.StructField.Tag.Get(tags.TagHashiCorpVault)
	secretValue, err := config.readSecret(context, nameField)
	if err != nil {
		return infra.NewGoStructorNoValue(context.Value, err)
	}
	return converters.ConvertBetweenPrimitiveTypes(reflect.ValueOf(secretValue), reflect.Indirect(context.Value))
}

func (config VaultConfig) GetComplexType(context *structContext) infra.GoStructorValue {
	if err := config.prepareLayer(context); err != nil {
		return infra.NewGoStructorNoValue(context.Value, err)
	}
	nameField := context.StructField.Tag.Get(tags.TagHashiCorpVault)
	secretValue, err := config.readSecret(context, nameField)
	if err != nil {
		return infra.NewGoStructorNoValue(context.Value, err)
	}
	kind := reflect.Indirect(context.Value).Kind()
	if kind == reflect.Slice {
		secretString, ok := secretValue.(string)
		if !ok {
			return infra.NewGoStructorNoValue(context.Value, errors.New("cf_vault slice fields require a comma-separated string secret value"))
		}
		return converters.ConvertBetweenComplexTypes(reflect.ValueOf(strings.Split(secretString, ",")), reflect.Indirect(context.Value))
	}
	return infra.NewGoStructorNoValue(context.Value, errors.New("not supported complex type"))
}
