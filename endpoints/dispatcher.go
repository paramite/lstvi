package endpoints

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	goji "goji.io"
	"goji.io/pat"

	"github.com/paramite/lstvi/memcache"
)

func RunDispatcher(memInitCap int, bindIface string, bindPort int, tlsCert, tlsKey string) {
	// database init
	cache := memcache.NewMessageCache(memInitCap)

	// endpoints definition
	mux := goji.NewMux()
	mux.HandleFunc(pat.Post("/message"), Message(cache))
	mux.HandleFunc(pat.Get("/messages"), MessageList(cache))

	bindAddr := fmt.Sprintf("%s:%d", bindIface, bindPort)
	server := &http.Server{Handler: mux, Addr: bindAddr}

	intrChan := make(chan os.Signal, 1)
	doneChan := make(chan bool, 1)
	signal.Notify(intrChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-intrChan
		server.Shutdown(context.Background())
		doneChan <- true
	}()

	fmt.Printf("Starting server...\n")
	go func() {
		if len(tlsCert) > 0 && len(tlsKey) > 0 {
			fmt.Printf("%s\n", server.ListenAndServeTLS(tlsCert, tlsKey))
		} else {
			fmt.Printf("%s\n", server.ListenAndServe())
		}
	}()

	<-doneChan
	fmt.Printf("Shutting down server...\n")
}
