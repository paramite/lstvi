package main

import (
	"flag"
	"fmt"
	"net/http"

	goji "goji.io"
	"goji.io/pat"

	"github.com/paramite/lstvi/endpoints"
	"github.com/paramite/lstvi/memcache"
)

func main() {
	memInitCap := flag.Int("initial-capacity", 100, "Initial capacity for data storage.")
	bindIface := flag.String("bind-iface", "127.0.0.1", "Interface on which to serve.")
	bindPort := flag.Int("bind-port", 16661, "Port on which to serve.")
	tlsCert := flag.String("tls-cert", "", "Path to tls certificate.")
	tlsKey := flag.String("tls-key", "", "Path to tls private key matching given tls-cert.")
	flag.Parse()

	// database init
	cache := memcache.NewMessageCache(*memInitCap)

	// endpoints definition
	mux := goji.NewMux()
	mux.HandleFunc(pat.Post("/message"), endpoints.Message(cache))
	mux.HandleFunc(pat.Get("/messages"), endpoints.MessageList(cache))

	bindAddr := fmt.Sprintf("%s:%d", *bindIface, *bindPort)
	if len(*tlsCert) > 0 && len(*tlsKey) > 0 {
		fmt.Printf("%s\n", http.ListenAndServeTLS(bindAddr, *tlsCert, *tlsKey, mux))
	} else {
		fmt.Printf("%s\n", http.ListenAndServe(fmt.Sprintf("%s:%d", *bindIface, *bindPort), mux))
	}
}
