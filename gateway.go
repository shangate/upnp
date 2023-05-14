package upnp

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

type Gateway struct {
	Active        bool
	LocalAddr     string
	GatewayName   string
	Host          string
	DeviceDescUrl string
	Cache         string
	ST            string
	USN           string
	deviceType    string //"urn:schemas-upnp-org:service:WANIPConnection:1"
	ControlURL    string
	ServiceType   string
}

func (gateway *Gateway) Send() bool {
	message := "M-SEARCH * HTTP/1.1\r\n" +
		"HOST: 239.255.255.250:1900\r\n" +
		"ST: urn:schemas-upnp-org:service:WANIPConnection:1\r\n" +
		"MAN: \"ssdp:discover\"\r\n" + "MX: 3\r\n\r\n"
	result, err := gateway.send(message)
	if err != nil {
		gateway.Active = false
		return false
	}
	err = gateway.resolve(result)
	if err != nil {
		return false
	}

	gateway.ServiceType = "urn:schemas-upnp-org:service:WANIPConnection:1"
	gateway.Active = true
	return true
}

func (gateway *Gateway) send(message string) (result string, err error) {
	remoteAddr, err := net.ResolveUDPAddr("udp4", "239.255.255.250:1900")
	if err != nil {
		return result, fmt.Errorf("resolve udp addr error %v", err)
	}

	conn, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4zero, Port: 0})
	if err != nil {
		return result, fmt.Errorf("listen udp error %v", err)
	}
	defer conn.Close()

	_, err = conn.WriteToUDP([]byte(message), remoteAddr)
	if err != nil {
		fmt.Println("send message error ", err)
		return result, fmt.Errorf("send message error %v", err)
	}

	buf := make([]byte, 1024)
	err = conn.SetReadDeadline(time.Now().Add(time.Second * 3))
	if err != nil {
		fmt.Println("set read deadline error ", err)
		return result, fmt.Errorf("set read deadline error %v", err)
	}
	n, addr, err := conn.ReadFromUDP(buf)
	if err != nil {
		fmt.Println("recv message error ", err)
		return result, fmt.Errorf("recv message error %v", err)
	}

	gatewayAddr, _ := net.ResolveUDPAddr("udp4", addr.String())
	gateway.LocalAddr = gatewayAddr.IP.String()

	conn.SetReadDeadline(time.Time{})

	result = string(buf[:n])
	return result, nil
}

func (gateway *Gateway) resolve(result string) error {
	lines := strings.Split(result, "\r\n")
	for _, line := range lines {
		nameValues := strings.SplitAfterN(line, ":", 2)
		if len(nameValues) < 2 {
			continue
		}
		switch strings.ToUpper(strings.Trim(strings.Split(nameValues[0], ":")[0], " ")) {
		case "ST":
			gateway.ST = nameValues[1]
		case "CACHE-CONTROL":
			gateway.Cache = nameValues[1]
		case "LOCATION":
			urls := strings.Split(strings.Split(nameValues[1], "//")[1], "/")
			addr, err := net.ResolveUDPAddr("udp", urls[0])
			if err != nil {
				return err
			}
			gateway.Host = net.JoinHostPort(gateway.LocalAddr, strconv.Itoa(addr.Port))
			gateway.DeviceDescUrl = "/" + urls[1]
		case "SERVER":
			gateway.GatewayName = nameValues[1]
		default:
		}
	}
	return nil
}
