package upnp

import (
	"fmt"
	"net"
	"strings"
	"time"
)

type Gateway struct {
	Active        bool
	LocalHost     string
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
	c := make(chan string)
	go gateway.send(message, c)
	result := <-c
	if result == "" {
		gateway.Active = false
		return false
	}
	gateway.resolve(result)

	gateway.ServiceType = "urn:schemas-upnp-org:service:WANIPConnection:1"
	gateway.Active = true
	return true
}

func (gateway *Gateway) send(message string, c chan string) error {
	var conn *net.UDPConn
	defer func() {
		if r := recover(); r != nil {
			//timeout
		}
	}()
	go func(conn *net.UDPConn) {
		defer func() {
			if r := recover(); r != nil {
				//doesn't timeout
			}
		}()
		time.Sleep(time.Second * 3)
		c <- ""
		conn.Close()
	}(conn)
	remoteAddr, err := net.ResolveUDPAddr("udp", "239.255.255.250:1900")
	if err != nil {
		return fmt.Errorf("resolve udp addr error %v", err)
	}
	locaAddr, err := net.ResolveUDPAddr("udp", gateway.LocalHost+":")
	if err != nil {
		return fmt.Errorf("resolve udp addr error %v", err)
	}
	conn, err = net.ListenUDP("udp", locaAddr)
	if err != nil {
		return fmt.Errorf("listen udp error %v", err)
	}
	defer conn.Close()
	_, err = conn.WriteToUDP([]byte(message), remoteAddr)
	if err != nil {
		return fmt.Errorf("send message error %v", err)
	}
	buf := make([]byte, 1024)
	n, _, err := conn.ReadFromUDP(buf)
	if err != nil {
		return fmt.Errorf("recv message error %v", err)
	}

	result := string(buf[:n])
	c <- result
	return nil
}

func (gateway *Gateway) resolve(result string) {
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
			gateway.Host = urls[0]
			gateway.DeviceDescUrl = "/" + urls[1]
		case "SERVER":
			gateway.GatewayName = nameValues[1]
		default:
		}
	}
}
