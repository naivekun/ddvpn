package server

import (
	"crypto/tls"
	"ddvpn/common"
	"ddvpn/conf"
	"ddvpn/tunnel"
	"encoding/binary"
	"encoding/json"
	"io"
	"log"
	"net"
	"strconv"

	"github.com/songgao/water"
)

const (
	BUFFERSIZE = 1600
)

type Server struct {
	config        conf.ServerConfig
	tlsDialConfig *tls.Config
	Tun           *water.Interface
	connEndpoint  map[[4]byte]net.Conn
}

type ClientAuthMessage struct {
	Username      string `json:"username"`
	Password      string `json:"password"`
	ClientAddress string `json:"client_addr"`
}

func Config(c conf.Config) (*Server, error) {
	ret := &Server{}
	err := json.Unmarshal(c.ServerConfig, &ret.config)
	if err != nil {
		return nil, err
	}

	ret.tlsDialConfig = conf.ParseTLSServerConfig(c.TLSConfig)
	ret.connEndpoint = make(map[[4]byte]net.Conn)
	return ret, nil
}

func Run(s *Server) error {
	listener, err := tls.Listen("tcp", s.config.ListenAddr, s.tlsDialConfig)
	if err != nil {
		return common.Raise("listen to address " + s.config.ListenAddr + " failed: " + err.Error())
	}
	s.Tun, err = tunnel.CreateNewTunnel("tun", s.config.ServerTunIP, 1300)
	if err != nil {
		return common.Raise("failed to create tun device: " + err.Error())
	}

	go func() {
		err := s.HandleInput()
		log.Fatalln(err.Error())
	}()
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("accept connection failed: " + err.Error())
			continue
		}
		err = s.Handshake(conn)
		if err != nil {
			log.Println("connection handshake failed: " + err.Error())
			continue
		}

	}
}

func (s *Server) Handshake(conn net.Conn) error {
	authMsgLengthRecvBuf := [2]byte{}
	_, err := io.ReadFull(conn, authMsgLengthRecvBuf[:])
	if err != nil {
		return common.Raise("read auth message length failed: " + err.Error())
	}
	authMsgLen := int(binary.BigEndian.Uint16(authMsgLengthRecvBuf[:]))
	log.Println("auth message len: " + strconv.Itoa(int(authMsgLen)))
	if authMsgLen > 0x7fff {
		return common.Raise("invalid auth message length: " + strconv.Itoa(int(authMsgLen)))
	}

	authMsgRecvBuf := make([]byte, authMsgLen)
	nRead, err := io.ReadFull(conn, authMsgRecvBuf)
	if err != nil {
		return common.Raise("error read auth message: " + err.Error())
	}
	if nRead != authMsgLen {
		return common.Raise("invalid auth message read: " + strconv.Itoa(nRead))
	}

	log.Println("auth msg length rcvd: " + strconv.Itoa(nRead))
	authMsg := &ClientAuthMessage{}
	err = json.Unmarshal(authMsgRecvBuf, authMsg)
	if err != nil {
		return common.Raise("error decode auth message: " + err.Error())
	}
	authResult := common.PAMAuth("passwd", authMsg.Username, authMsg.Password)
	if authResult != nil {
		_, err := conn.Write([]byte{common.AUTH_FAILED})
		if err != nil {
			return common.Raise("error send auth status: " + err.Error())
		}
		//auth failed
		return common.Raise("auth failed from " + conn.RemoteAddr().String())
	}
	_, err = conn.Write([]byte{common.AUTH_SUCCESS})
	if err != nil {
		return common.Raise("error send auth status: " + err.Error())
	}

	//auth complete

	ip, _, err := net.ParseCIDR(authMsg.ClientAddress)
	ip = ip.To4()
	if ip == nil || len([]byte(ip)) != 4 {
		return common.Raise("invalid client tunnel IP: " + authMsg.ClientAddress)
	}
	s.connEndpoint[[4]byte{([]byte)(ip)[0], ([]byte)(ip)[1], ([]byte)(ip)[2], ([]byte)(ip)[3]}] = conn
	go func() {
		_, err := common.FastCopy(conn, s.Tun)
		if err != nil {
			log.Println("server copy tls -> tun exited: " + err.Error())
		}
	}()
	return nil
}

func (s *Server) HandleInput() error {

	var packet common.IPPacket = make([]byte, BUFFERSIZE)
	for {
		packetLen, err := s.Tun.Read(packet[:s.config.MTU])
		if err != nil {
			return common.Raise("error read packet from tunnel")
		}
		if packet.IPver() != 4 {
			log.Println("non IPv4 packet received")
			continue
		}
		isVpnNetworkPacket := false
		dst := packet.Dst()
		epConn, ok := s.connEndpoint[dst]
		if ok {
			isVpnNetworkPacket = true
		}
		if packet.IsMulticast() {
			isVpnNetworkPacket = true
		}

		if isVpnNetworkPacket {
			if packet.IsMulticast() {
				for _, ep := range s.connEndpoint {
					_, err := ep.Write(packet[:packetLen])
					if err != nil {
						log.Println("error write packet to tunnel: " + err.Error())
					}
				}
			} else {
				_, err := epConn.Write(packet[:packetLen])
				if err != nil {
					log.Println("error write packet to tunnel: " + err.Error())
				}
			}
		}

	}
}
