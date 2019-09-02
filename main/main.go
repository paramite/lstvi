package main

import (
	"flag"
	"fmt"
	"net/http"

	goji "goji.io"
	"goji.io/pat"

	"github.com/paramite/lstvi/endpoints"
)

func main() {
	bindIface := flag.String("bind-iface", "127.0.0.1", "Interface on which to serve.")
	bindPort := flag.Int("bind-port", 666, "Port on which to serve.")
	tsdbPath := flag.String("storage", "~/lstvi_db.db", "Path to server database.")
	flag.Parse()

	mux := goji.NewMux()
	mux.HandleFunc(pat.Post("/message"), endpoints.Message(tsdb))
	mux.HandleFunc(pat.Get("/messages"), endpoints.MessageList(tsdb))

	http.ListenAndServe(fmt.Sprintf("%s:%d", *bindIface, *bindPort), mux)
}
