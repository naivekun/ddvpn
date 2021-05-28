package conf

import (
	"crypto/tls"
	"crypto/x509"
	"ddvpn/common"
	"encoding/json"
	"io/ioutil"
	"log"
)

type TLSConfig struct {
	UseCustomCA bool   `json:"useCustomCA"`
	CaCert      string `json:"caCert"`

	UseClientCert bool   `json:"useClientCert"`
	ClientCert    string `json:"clientCert"`
	ClientKey     string `json:"clientKey"`

	VerifyClientCert bool   `json:"verifyClientCert"`
	ServerCert       string `json:"serverCert"`
	ServerKey        string `json:"serverKey"`
}

func DumpDefaultTlsConfig() json.RawMessage {
	ret, _ := json.Marshal(&TLSConfig{
		UseCustomCA:      true,
		CaCert:           "certs/ca/ca_cert.pem",
		UseClientCert:    true,
		ClientCert:       "certs/client/client_cert.pem",
		ClientKey:        "certs/client/private/client_key.pem",
		VerifyClientCert: true,
		ServerCert:       "certs/server/server_cert.pem",
		ServerKey:        "certs/server/private/server_key.pem",
	})
	return ret
}

func ParseTLSClientConfig(c json.RawMessage) *tls.Config {
	tlsConfig := TLSConfig{}
	common.Must(json.Unmarshal(c, &tlsConfig))

	tlsDialConfig := &tls.Config{}
	if tlsConfig.UseCustomCA {
		pool := x509.NewCertPool()
		caCert, err := ioutil.ReadFile(common.JoinCurrentPath(tlsConfig.CaCert))
		if err != nil {
			log.Fatalln("read ca failed")
			return nil
		}
		pool.AppendCertsFromPEM(caCert)
		tlsDialConfig.RootCAs = pool
	}

	if tlsConfig.UseClientCert {
		clientCrt, err := tls.LoadX509KeyPair(tlsConfig.ClientCert, tlsConfig.ClientKey)
		if err != nil {
			log.Fatalln("read client key failed")
			return nil
		}
		tlsDialConfig.Certificates = []tls.Certificate{clientCrt}
	}
	return tlsDialConfig
}

func ParseTLSServerConfig(c json.RawMessage) *tls.Config {
	tlsConfig := &TLSConfig{}
	common.Must(json.Unmarshal(c, &tlsConfig))

	pool := x509.NewCertPool()
	caCrt, err := ioutil.ReadFile(common.JoinCurrentPath(tlsConfig.CaCert))
	if err != nil {
		log.Fatalln("read ca failed")
		return nil
	}
	pool.AppendCertsFromPEM(caCrt)

	cert, err := tls.LoadX509KeyPair(tlsConfig.ServerCert, tlsConfig.ServerKey)
	if err != nil {
		log.Fatalln("tls load server key failed")
		return nil
	}

	tlsDialConfig := &tls.Config{
		Certificates: []tls.Certificate{
			cert,
		},
		ClientCAs: pool,
	}

	if tlsConfig.VerifyClientCert {
		tlsDialConfig.ClientAuth = tls.RequireAndVerifyClientCert
	}
	return tlsDialConfig
}
