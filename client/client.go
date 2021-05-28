package client

import (
	"crypto/tls"
	"ddvpn/common"
	"ddvpn/conf"
	"ddvpn/tunnel"
	"encoding/binary"
	"encoding/json"
	"io"
	"log"
)

type Client struct {
	config        conf.ClientConfig
	tlsDialConfig *tls.Config
}

type ClientAuthMessage struct {
	Username      string `json:"username"`
	Password      string `json:"password"`
	ClientAddress string `json:"client_addr"`
}

func Config(c conf.Config) (*Client, error) {
	ret := &Client{}
	err := json.Unmarshal(c.ClientConfig, &ret.config)
	if err != nil {
		return nil, err
	}

	ret.tlsDialConfig = conf.ParseTLSClientConfig(c.TLSConfig)
	return ret, nil
}

func Run(c *Client) error {
	conn, err := tls.Dial("tcp", c.config.ServerAddr, c.tlsDialConfig)
	if err != nil {
		return common.Raise("tls dial to " + c.config.ServerAddr + " failed: " + err.Error())
	}
	defer conn.Close()

	tun, err := tunnel.CreateNewTunnel("tun", c.config.ClientTunIP, c.config.MTU)
	if err != nil {
		return common.Raise("create tunnel failed: " + err.Error())
	}
	defer tun.Close()

	log.Println("connection established...")
	// auth: 2 byte message len + json auth message
	authMsg := &ClientAuthMessage{
		Username:      c.config.Username,
		Password:      c.config.Passsowrd,
		ClientAddress: c.config.ClientTunIP,
	}
	authMsgBytes, _ := json.Marshal(&authMsg)
	authMsgLen := len(authMsgBytes)
	if authMsgLen > 0x7fff || authMsgLen < 0 {
		return common.Raise("auth message too long")
	}
	authMsgBuf := make([]byte, 2+authMsgLen)
	binary.BigEndian.PutUint16(authMsgBuf, uint16(authMsgLen))
	copy(authMsgBuf[2:], authMsgBytes)

	authMsgLen += 2
	if err != nil {
		return common.Raise("error write auth message: " + err.Error())
	}
	conn.Write(authMsgBuf[:authMsgLen])

	authStatusBuf := make([]byte, 1)
	_, err = io.ReadFull(conn, authStatusBuf)
	if err != nil {
		return common.Raise("error read auth result: " + err.Error())
	}
	if authStatusBuf[0] == common.AUTH_FAILED {
		return common.Raise("server: auth failed")
	} else {
		//auth success
		log.Println("auth success")
	}

	go func() {
		_, err := common.FastCopy(conn, tun)
		if err != nil {
			log.Println("client copy tls -> tun exited: " + err.Error())
		}
		if tun != nil {
			tun.Close()
		}
	}()

	_, err = common.FastCopy(tun, conn)
	if err != nil {
		log.Println("client copy tun -> tls exited: " + err.Error())
	}
	return nil
}
