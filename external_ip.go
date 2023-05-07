package upnp

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func SendRequestForExternalIP(url string) (ip string, err error) {
	tr := &http.Transport{
		DisableKeepAlives: true,
	}
	client := http.Client{Transport: tr, Timeout: time.Second}

	request := buildRequestForExternalIP(url)
	response, _ := client.Do(request)
	if response != nil {
		resultBody, _ := ioutil.ReadAll(response.Body)
		if response.StatusCode == 200 {
			return resolveForExternalIP(string(resultBody))
		}
	}
	return ip, fmt.Errorf("send request for external ip error")
}

func buildRequestForExternalIP(url string) *http.Request {
	//请求头
	header := http.Header{}
	header.Set("Accept", "text/html, image/gif, image/jpeg, *; q=.2, */*; q=.2")
	header.Set("SOAPAction", `"urn:schemas-upnp-org:service:WANIPConnection:1#GetExternalIPAddress"`)
	header.Set("Content-Type", "text/xml")
	header.Set("Connection", "Close")
	header.Set("Content-Length", "")
	//请求体
	body := Node{Name: "SOAP-ENV:Envelope",
		Attr: map[string]string{"xmlns:SOAP-ENV": `"http://schemas.xmlsoap.org/soap/envelope/"`,
			"SOAP-ENV:encodingStyle": `"http://schemas.xmlsoap.org/soap/encoding/"`}}
	childOne := Node{Name: `SOAP-ENV:Body`}
	childTwo := Node{Name: `m:GetExternalIPAddress`,
		Attr: map[string]string{"xmlns:m": `"urn:schemas-upnp-org:service:WANIPConnection:1"`}}
	childOne.AddChild(childTwo)
	body.AddChild(childOne)

	bodyStr := body.BuildXML()
	//请求
	request, _ := http.NewRequest("POST", url, strings.NewReader(bodyStr))
	request.Header = header
	request.Header.Set("Content-Length", strconv.Itoa(len([]byte(body.BuildXML()))))
	return request
}

// NewExternalIPAddress
func resolveForExternalIP(resultStr string) (ip string, err error) {
	inputReader := strings.NewReader(resultStr)
	decoder := xml.NewDecoder(inputReader)
	ISexternalIP := false
	for t, err := decoder.Token(); err == nil; t, err = decoder.Token() {
		switch token := t.(type) {
		// 处理元素开始（标签）
		case xml.StartElement:
			name := token.Name.Local
			if name == "NewExternalIPAddress" {
				ISexternalIP = true
			}
		// 处理元素结束（标签）
		case xml.EndElement:
		// 处理字符数据（这里就是元素的文本）
		case xml.CharData:
			if ISexternalIP == true {
				return string([]byte(token)), nil
			}
		default:
			// ...
		}
	}
	return ip, fmt.Errorf("resolve data for external ip error")
}
