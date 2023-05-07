package main

import (
	"fmt"
	"github.com/shangate/upnp"
)

func main() {
	u, e := upnp.NewUpnp()
	if e != nil {
		fmt.Println("new upnp error ", e)
		return
	}
	u.AddPortMapping(23322, 23322, "UDP")
	u.RemovePortMapping(23322, "UDP")

}
