package conf

import "encoding/json"

type ClientConfig struct {
	ServerAddr  string `json:"server"`
	Username    string `json:"username"`
	Passsowrd   string `json:"password"`
	ClientTunIP string `json:"client_tun_ip"`
	MTU         int    `json:"mtu"`
}

func DumpDefaultClientConfig() json.RawMessage {
	ret, _ := json.Marshal(ClientConfig{
		ServerAddr:  "10.6.6.6:4433",
		Username:    "admin",
		Passsowrd:   "admin",
		ClientTunIP: "10.10.10.10/24",
		MTU:         1300,
	})
	return ret
}
