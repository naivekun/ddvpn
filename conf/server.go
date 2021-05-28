package conf

import "encoding/json"

type ServerConfig struct {
	ListenAddr  string `json:"listen"`
	ServerTunIP string `json:"server_tun_ip"`
	MTU         int    `json:"mtu"`
}

func DumpDefaultServerConfig() json.RawMessage {
	ret, _ := json.Marshal(ServerConfig{
		ListenAddr:  "0.0.0.0:4433",
		ServerTunIP: "10.10.10.100/24",
		MTU:         1300,
	})
	return ret
}
