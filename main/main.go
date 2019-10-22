package main

import (
	"flag"

	"github.com/paramite/lstvi/endpoints"
)

func main() {
	memInitCap := flag.Int("initial-capacity", 100, "Initial capacity for data storage.")
	bindIface := flag.String("bind-iface", "127.0.0.1", "Interface on which to serve.")
	bindPort := flag.Int("bind-port", 16661, "Port on which to serve.")
	tlsCert := flag.String("tls-cert", "", "Path to tls certificate.")
	tlsKey := flag.String("tls-key", "", "Path to tls private key matching given tls-cert.")
	flag.Parse()

	endpoints.RunDispatcher(*memInitCap, *bindIface, *bindPort, *tlsCert, *tlsKey)
}
