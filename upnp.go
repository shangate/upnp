package upnp

import (
	"errors"
	"fmt"
	"log"
	"net"
	"strings"
)

type Upnp struct {
	Gateway          Gateway
	GatewayInsideIP  string
	GatewayOutsideIP string
	CtrlUrl          string
}

func (upnp *Upnp) AddPortMapping(localPort, remotePort int, protocol string) (err error) {
	if upnp.GatewayOutsideIP == "" {
		if err := upnp.checkExternalIP(); err != nil {
			return err
		}
	}
	if ok := SendRequestForAddPortMapping("http://"+upnp.Gateway.Host+upnp.CtrlUrl, upnp.Gateway.LocalHost, localPort, remotePort, protocol); ok {
		log.Println("add port mapping successfully：protocol:", protocol, "local:", localPort, "remote:", remotePort)
	} else {
		upnp.Gateway.Active = false
		return fmt.Errorf("add port mapping failed")
	}
	return nil
}

func (upnp *Upnp) RemovePortMapping(remotePort int, protocol string) (err error) {
	if upnp.GatewayOutsideIP == "" {
		if err := upnp.checkExternalIP(); err != nil {
			return err
		}
	}
	return SendRequestForRemovePortMapping("http://"+upnp.Gateway.Host+upnp.CtrlUrl, remotePort, protocol)
}

func (upnp *Upnp) checkExternalIP() (err error) {
	if upnp.GatewayOutsideIP == "" {
		upnp.GatewayOutsideIP, err = upnp.getExternalIP()
		if err != nil {
			return err
		}
	}
	return nil
}

func (upnp *Upnp) getExternalIP() (ip string, err error) {
	if upnp.CtrlUrl == "" {
		if err := upnp.getDeviceDesc(); err != nil {
			return ip, err
		}
	}
	return SendRequestForExternalIP("http://" + upnp.Gateway.Host + upnp.CtrlUrl)
}

func (upnp *Upnp) searchGateway() (err error) {
	defer func(err error) {
		if errTemp := recover(); errTemp != nil {
			log.Println("upnp module error ", errTemp)
			err = errTemp.(error)
		}
	}(err)

	if upnp.Gateway.LocalHost == "" {
		conn, err := net.Dial("udp", "shangate.com:80")
		if err != nil {
			return fmt.Errorf("network error")
		}
		defer conn.Close()
		upnp.Gateway.LocalHost = strings.Split(conn.LocalAddr().String(), ":")[0]
	}
	ok := upnp.Gateway.Send()
	if ok {
		return nil
	}
	return errors.New("doesn't find any gateway device")
}

func (upnp *Upnp) getDeviceDesc() (err error) {
	if upnp.GatewayInsideIP == "" {
		if err := upnp.searchGateway(); err != nil {
			return err
		}
	}
	upnp.CtrlUrl, err = SendForDescription("http://"+upnp.Gateway.Host+upnp.Gateway.DeviceDescUrl, upnp.Gateway.Host, upnp.Gateway.ServiceType)
	if err != nil {
		return err
	}
	upnp.Gateway.Active = true
	return nil
}
