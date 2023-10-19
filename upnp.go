package upnp

import (
	"errors"
	"fmt"
	"log"
	"net"
	"strings"
	"time"
)

type Upnp struct {
	localAddr        string
	gateway          Gateway
	gatewayInsideIP  string
	gatewayOutsideIP string
	ctrlUrl          string
}

func NewUpnp() (upnp *Upnp, err error) {
	upnp = &Upnp{}
	err = upnp.checkExternalIP()
	if err != nil {
		return nil, err
	}
	return upnp, nil
}

func (upnp *Upnp) AddPortMapping(localPort, remotePort int, protocol string) (err error) {
	if upnp.gatewayOutsideIP == "" {
		if err := upnp.checkExternalIP(); err != nil {
			return err
		}
	}
	if ok := SendRequestForAddPortMapping("http://"+upnp.gateway.Host+upnp.ctrlUrl, upnp.localAddr, localPort, remotePort, protocol); ok {
		log.Println("add port mapping successfully：protocol:", protocol, "local:", localPort, "remote:", remotePort)
	} else {
		upnp.gateway.Active = false
		return fmt.Errorf("add port mapping failed")
	}
	return nil
}

func (upnp *Upnp) RemovePortMapping(remotePort int, protocol string) (err error) {
	if upnp.gatewayOutsideIP == "" {
		if err := upnp.checkExternalIP(); err != nil {
			return err
		}
	}
	return SendRequestForRemovePortMapping("http://"+upnp.gateway.Host+upnp.ctrlUrl, remotePort, protocol)
}

func (upnp *Upnp) checkExternalIP() (err error) {
	if upnp.gatewayOutsideIP == "" {
		upnp.gatewayOutsideIP, err = upnp.getExternalIP()
		if err != nil {
			return err
		}
	}
	return nil
}

func (upnp *Upnp) getExternalIP() (ip string, err error) {
	if upnp.ctrlUrl == "" {
		if err := upnp.getDeviceDesc(); err != nil {
			return ip, err
		}
	}
	return SendRequestForExternalIP("http://" + upnp.gateway.Host + upnp.ctrlUrl)
}

func (upnp *Upnp) searchGateway() (err error) {
	defer func(err error) {
		if errTemp := recover(); errTemp != nil {
			log.Println("upnp module error ", errTemp)
			err = errTemp.(error)
		}
	}(err)

	if upnp.localAddr == "" {
		conn, err := net.DialTimeout("udp4", "google.com:80", time.Second*3)
		if err != nil {
			return fmt.Errorf("network error")
		}
		defer conn.Close()
		upnp.localAddr = strings.Split(conn.LocalAddr().String(), ":")[0]
	}
	ok := upnp.gateway.Send()
	if ok {
		return nil
	}
	return errors.New("doesn't find any gateway device")
}

func (upnp *Upnp) getDeviceDesc() (err error) {
	if upnp.gateway.DeviceDescUrl == "" {
		if err := upnp.searchGateway(); err != nil {
			return err
		}
	}
	upnp.ctrlUrl, err = SendForDescription("http://"+upnp.gateway.Host+upnp.gateway.DeviceDescUrl, upnp.gateway.Host, upnp.gateway.ServiceType)
	if err != nil {
		return err
	}
	upnp.gateway.Active = true
	return nil
}
