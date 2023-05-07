package upnp

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

func SendForDescription(url string, gatewayHost string, serviceType string) (ctrlUrl string, err error) {
	tr := &http.Transport{
		DisableKeepAlives: true,
	}
	client := http.Client{Transport: tr, Timeout: time.Second}

	request := buildRequestForDescription(url, gatewayHost)
	response, _ := client.Do(request)
	if response != nil {
		resultBody, _ := ioutil.ReadAll(response.Body)
		if response.StatusCode == 200 {
			return resolve(serviceType, string(resultBody)), nil
		}
	}
	return ctrlUrl, fmt.Errorf("request description error")
}

func buildRequestForDescription(url string, gatewayHost string) *http.Request {
	//请求头
	header := http.Header{}
	header.Set("Accept", "text/html, image/gif, image/jpeg, *; q=.2, */*; q=.2")
	header.Set("User-Agent", "preston")
	header.Set("Host", gatewayHost)
	header.Set("Connection", "keep-alive")

	//请求
	request, _ := http.NewRequest("GET", url, nil)
	request.Header = header
	// request := http.Request{Method: "GET", Proto: "HTTP/1.1",
	// 	Host: this.upnp.Gateway.Host, Url: this.upnp.Gateway.DeviceDescUrl, Header: header}
	return request
}

func resolve(serviceType string, data string) (ctrlUrl string) {
	inputReader := strings.NewReader(data)

	// 从文件读取，如可以如下：
	// content, err := ioutil.ReadFile("studygolang.xml")
	// decoder := xml.NewDecoder(bytes.NewBuffer(content))

	lastLabel := ""

	ISUpnpServer := false

	IScontrolURL := false
	var controlURL string //`controlURL`
	// var eventSubURL string //`eventSubURL`
	// var SCPDURL string     //`SCPDURL`

	decoder := xml.NewDecoder(inputReader)
	for t, err := decoder.Token(); err == nil && !IScontrolURL; t, err = decoder.Token() {
		switch token := t.(type) {
		// 处理元素开始（标签）
		case xml.StartElement:
			if ISUpnpServer {
				name := token.Name.Local
				lastLabel = name
			}

		// 处理元素结束（标签）
		case xml.EndElement:
			// log.Println("结束标记：", token.Name.Local)
		// 处理字符数据（这里就是元素的文本）
		case xml.CharData:
			//得到url后其他标记就不处理了
			content := string([]byte(token))

			//找到提供端口映射的服务
			if content == serviceType {
				ISUpnpServer = true
				continue
			}
			//urn:upnp-org:serviceId:WANIPConnection
			if ISUpnpServer {
				switch lastLabel {
				case "controlURL":

					controlURL = content
					IScontrolURL = true
				case "eventSubURL":
					// eventSubURL = content
				case "SCPDURL":
					// SCPDURL = content
				}
			}
		default:
			// ...
		}
	}
	return controlURL
}
