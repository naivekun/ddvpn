package tunnel

import (
	"ddvpn/common"
	"log"
	"net"

	"github.com/milosgajdos/tenus"
	"github.com/songgao/water"
)

type TunnelEndpoint struct {
	IPAddr string
}

func CreateNewTunnel(tunnel_name string, ip string, MTU int) (*water.Interface, error) {
	tunnelIP, tunnelSubnet, err := net.ParseCIDR(ip)
	if err != nil {
		return nil, common.Raise("parse tunnel address failed: " + err.Error())
	}
	iface, err := water.New(
		water.Config{
			DeviceType: water.TUN,
		},
	)

	if err != nil {
		return nil, common.Raise("create tun device failed: " + err.Error())
	}
	log.Println("interface created: " + iface.Name())

	link, err := tenus.NewLinkFrom(iface.Name())
	if err != nil {
		return nil, common.Raise("create tun link failed: " + err.Error())
	}
	err = link.SetLinkMTU(MTU)
	if err != nil {
		return nil, common.Raise("set link MTU failed: " + err.Error())
	}
	err = link.SetLinkIp(tunnelIP, tunnelSubnet)
	if err != nil {
		return nil, common.Raise("set interface IP failed: " + err.Error())
	}
	err = link.SetLinkUp()
	if err != nil {
		return nil, common.Raise("unable to set link up" + err.Error())
	}
	return iface, nil
}

func CreateTunnelRoute(iface *water.Interface) error {
	log.Println("create route for interface " + iface.Name())

	return nil
}
