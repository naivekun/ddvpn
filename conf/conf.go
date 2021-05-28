package conf

import (
	"ddvpn/common"
	"encoding/json"
	"io/ioutil"
	"os"
)

type Config struct {
	Mode         string          `json:"mode"`
	ClientConfig json.RawMessage `json:"client_config"`
	ServerConfig json.RawMessage `json:"server_config"`
	TLSConfig    json.RawMessage `json:"tls_config"`
}

func DumpDefaultConfig() Config {
	return Config{
		Mode:         "client",
		ClientConfig: DumpDefaultClientConfig(),
		ServerConfig: DumpDefaultServerConfig(),
		TLSConfig:    DumpDefaultTlsConfig(),
	}
}

func CreateNewConfigFile(filename string) error {
	if common.FileExists(filename) {
		return common.Raise("file exists: " + filename)
	}
	configBytes, err := json.MarshalIndent(DumpDefaultConfig(), "", "\t")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filename, configBytes, 0644)
}

func (config *Config) ReadConfigFile(path2config string) error {
	file, err := os.Open(path2config)
	if err != nil {
		return err
	}
	defer file.Close()

	return json.NewDecoder(file).Decode(config)
}
