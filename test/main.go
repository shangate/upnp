package main

import "upnp"

func main() {
	var u upnp.Upnp
	u.RemovePortMapping(23322, "UDP")
}
