package upnp

import (
	"bytes"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Node struct {
	Name    string
	Content string
	Attr    map[string]string
	Child   []Node
}

func (n *Node) AddChild(node Node) {
	n.Child = append(n.Child, node)
}
func (n *Node) BuildXML() string {
	buf := bytes.NewBufferString("<")
	buf.WriteString(n.Name)
	for key, value := range n.Attr {
		buf.WriteString(" ")
		buf.WriteString(key + "=" + value)
	}
	buf.WriteString(">" + n.Content)

	for _, node := range n.Child {
		buf.WriteString(node.BuildXML())
	}
	buf.WriteString("</" + n.Name + ">")
	return buf.String()
}

func SendRequestForAddPortMapping(url string, localHost string, localPort, remotePort int, protocol string) bool {
	tr := &http.Transport{
		DisableKeepAlives: true,
	}
	client := http.Client{Transport: tr, Timeout: time.Second}

	request := buildRequestForAddPortMapping(url, localHost, localPort, remotePort, protocol)
	response, _ := client.Do(request)
	if response != nil {
		if response.StatusCode == 200 {
			return true
		}
	}
	return false
}

func buildRequestForAddPortMapping(url string, localHost string, localPort, remotePort int, protocol string) *http.Request {
	//请求头
	header := http.Header{}
	header.Set("Accept", "text/html, image/gif, image/jpeg, *; q=.2, */*; q=.2")
	header.Set("SOAPAction", `"urn:schemas-upnp-org:service:WANIPConnection:1#AddPortMapping"`)
	header.Set("Content-Type", "text/xml")
	header.Set("Connection", "Close")
	header.Set("Content-Length", "")
	//请求体
	body := Node{Name: "SOAP-ENV:Envelope",
		Attr: map[string]string{"xmlns:SOAP-ENV": `"http://schemas.xmlsoap.org/soap/envelope/"`,
			"SOAP-ENV:encodingStyle": `"http://schemas.xmlsoap.org/soap/encoding/"`}}
	childOne := Node{Name: `SOAP-ENV:Body`}
	childTwo := Node{Name: `m:AddPortMapping`,
		Attr: map[string]string{"xmlns:m": `"urn:schemas-upnp-org:service:WANIPConnection:1"`}}

	childList1 := Node{Name: "NewExternalPort", Content: strconv.Itoa(remotePort)}
	childList2 := Node{Name: "NewInternalPort", Content: strconv.Itoa(localPort)}
	childList3 := Node{Name: "NewProtocol", Content: protocol}
	childList4 := Node{Name: "NewEnabled", Content: "1"}
	childList5 := Node{Name: "NewInternalClient", Content: localHost}
	childList6 := Node{Name: "NewLeaseDuration", Content: "60"}
	childList7 := Node{Name: "NewPortMappingDescription", Content: "shangate"}
	childList8 := Node{Name: "NewRemoteHost"}
	childTwo.AddChild(childList1)
	childTwo.AddChild(childList2)
	childTwo.AddChild(childList3)
	childTwo.AddChild(childList4)
	childTwo.AddChild(childList5)
	childTwo.AddChild(childList6)
	childTwo.AddChild(childList7)
	childTwo.AddChild(childList8)

	childOne.AddChild(childTwo)
	body.AddChild(childOne)
	bodyStr := body.BuildXML()

	request, _ := http.NewRequest("POST", url, strings.NewReader(bodyStr))
	request.Header = header
	request.Header.Set("Content-Length", strconv.Itoa(len([]byte(bodyStr))))
	return request
}

func SendRequestForRemovePortMapping(url string, remotePort int, protocol string) error {
	tr := &http.Transport{
		DisableKeepAlives: true,
	}
	client := http.Client{Transport: tr, Timeout: time.Second}
	request := buildRequestForRemovePortMapping(url, remotePort, protocol)
	response, _ := client.Do(request)
	if response != nil {
		if response.StatusCode == 200 {
			return nil
		}
	}
	return fmt.Errorf("send request for remove port mapping error")
}

func buildRequestForRemovePortMapping(url string, remotePort int, protocol string) *http.Request {
	header := http.Header{}
	header.Set("Accept", "text/html, image/gif, image/jpeg, *; q=.2, */*; q=.2")
	header.Set("SOAPAction", `"urn:schemas-upnp-org:service:WANIPConnection:1#DeletePortMapping"`)
	header.Set("Content-Type", "text/xml")
	header.Set("Connection", "Close")
	header.Set("Content-Length", "")

	//请求体
	body := Node{Name: "SOAP-ENV:Envelope",
		Attr: map[string]string{"xmlns:SOAP-ENV": `"http://schemas.xmlsoap.org/soap/envelope/"`,
			"SOAP-ENV:encodingStyle": `"http://schemas.xmlsoap.org/soap/encoding/"`}}
	childOne := Node{Name: `SOAP-ENV:Body`}
	childTwo := Node{Name: `m:DeletePortMapping`,
		Attr: map[string]string{"xmlns:m": `"urn:schemas-upnp-org:service:WANIPConnection:1"`}}
	childList1 := Node{Name: "NewExternalPort", Content: strconv.Itoa(remotePort)}
	childList2 := Node{Name: "NewProtocol", Content: protocol}
	childList3 := Node{Name: "NewRemoteHost"}
	childTwo.AddChild(childList1)
	childTwo.AddChild(childList2)
	childTwo.AddChild(childList3)
	childOne.AddChild(childTwo)
	body.AddChild(childOne)
	bodyStr := body.BuildXML()

	request, _ := http.NewRequest("POST", url, strings.NewReader(bodyStr))
	request.Header = header
	request.Header.Set("Content-Length", strconv.Itoa(len([]byte(bodyStr))))
	return request
}
