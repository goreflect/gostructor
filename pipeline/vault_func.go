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
	configured, errConfigure := Configure(&properties.VaultConfiguration{}, "", []infra.FuncType{infra.FunctionSetupEnvironment}, "", true)
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
	if err := config.configureVault(); err != nil {
		return err
	}

	if errConn := config.vaultAvailable(); errConn != nil {
		logrus.Error("Error while connect to vault: ", errConn)
		return errConn
	}
	return nil
}

func (config VaultConfig) GetBaseType(context *structContext) infra.GoStructorValue {
	if err := config.prepareLayer(context); err != nil {
		return infra.NewGoStructorNoValue(context.Value, err)
	}
	nameField := context.StructField.Tag.Get(tags.TagHashiCorpVault)
	path := strings.Split(nameField, "#")[0]
	secretName := strings.Split(nameField, "#")[1]
	secret, err := config.connection.Logical().Read(path)
	secretValue := secret.Data[secretName]
	logrus.Debug("Secret Vault: ", secretValue)
	if err != nil {
		logrus.Error("Error while reading config from vault: ", err)
		return infra.NewGoStructorNoValue(context.Value, err)
	}
	return converters.ConvertBetweenPrimitiveTypes(reflect.ValueOf(secret.Data[secretName]), reflect.Indirect(context.Value))
}

func (config VaultConfig) GetComplexType(context *structContext) infra.GoStructorValue {
	if err := config.prepareLayer(context); err != nil {
		return infra.NewGoStructorNoValue(context.Value, err)
	}
	nameField := context.StructField.Tag.Get(tags.TagHashiCorpVault)
	path := strings.Split(nameField, "#")[0]
	secretName := strings.Split(nameField, "#")[1]
	secret, err := config.connection.Logical().Read(path)
	secretValue := secret.Data[secretName]
	logrus.Debug("Secret Vault: ", secretValue)
	if err != nil {
		logrus.Error("Error while reading config from vault: ", err)
		return infra.NewGoStructorNoValue(context.Value, err)
	}
	kind := reflect.Indirect(context.Value).Kind()
	if kind == reflect.Slice {
		return converters.ConvertBetweenComplexTypes(reflect.ValueOf(strings.Split(secret.Data[secretName].(string), ",")), reflect.Indirect(context.Value))
	}
	return infra.NewGoStructorNoValue(context.Value, errors.New("not supported complex type"))
}
