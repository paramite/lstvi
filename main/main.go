package main

import (
	"flag"
	"fmt"
	"net/http"

	goji "goji.io"
	"goji.io/pat"

	"github.com/asdine/storm"
	"github.com/paramite/lstvi/endpoints"
)

func main() {
	bindIface := flag.String("bind-iface", "127.0.0.1", "Interface on which to serve.")
	bindPort := flag.Int("bind-port", 16661, "Port on which to serve.")
	tsdbPath := flag.String("storage", "./lstvi_db.db", "Path to server database.")
	tlsCert := flag.String("tls-cert", "", "Path to tls certificate.")
	tlsKey := flag.String("tls-key", "", "Path to tls private key matching given tls-cert.")
	flag.Parse()

	// database init
	tsdb, err := storm.Open(*tsdbPath)
	if err != nil {
		fmt.Printf("Failed to open database %s: %s", *tsdbPath, err.Error())
	}
	defer tsdb.Close()

	// endpoints definition
	mux := goji.NewMux()
	mux.HandleFunc(pat.Post("/message"), endpoints.Message(tsdb))
	mux.HandleFunc(pat.Get("/messages"), endpoints.MessageList(tsdb))

	bindAddr := fmt.Sprintf("%s:%d", *bindIface, *bindPort)
	if len(*tlsCert) > 0 && len(*tlsKey) > 0 {
		fmt.Printf("%s\n", http.ListenAndServeTLS(bindAddr, *tlsCert, *tlsKey, mux))
	} else {
		fmt.Printf("%s\n", http.ListenAndServe(fmt.Sprintf("%s:%d", *bindIface, *bindPort), mux))
	}
}
