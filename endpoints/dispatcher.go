package endpoints

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	goji "goji.io"
	"goji.io/pat"

	"github.com/paramite/lstvi/memcache"
)

type ConnectionData struct {
	BindInterface string `json:"bind_host"`
	BindPort      int    `json:"bind_port"`
	TlsCert       string `json:"tls_cert"`
	TlsKey        string `json:"tls_key"`
}

func (self *ConnectionData) BindAddr() string {
	return fmt.Sprintf("%s:%d", self.BindInterface, self.BindPort)
}

func (self *ConnectionData) Secured() bool {
	return len(self.TlsCert) > 0 && len(self.TlsKey) > 0
}

type Dispatcher struct {
	cache      *memcache.MessageCache
	mux        *goji.Mux
	server     *http.Server
	config     string
	Connection *ConnectionData
}

func NewDispatcher(memInitCap int, config string) (*Dispatcher, error) {
	var output Dispatcher
	output.config = config
	err := output.loadConfig()
	if err != nil {
		return &output, err
	}

	// database init
	output.cache = memcache.NewMessageCache(memInitCap)
	go output.cache.Process()

	// endpoints definition
	output.mux = goji.NewMux()
	output.mux.HandleFunc(pat.Post("/message"), Message(output.cache.Queue))
	output.mux.HandleFunc(pat.Get("/messages"), MessageList(output.cache))

	return &output, nil
}

func (self *Dispatcher) loadConfig() error {
	file, err := ioutil.ReadFile(self.config)
	if err != nil {
		return err
	}

	var cfg ConnectionData
	err = json.Unmarshal(file, &cfg)
	if err != nil {
		return err
	}

	self.Connection = &cfg
	return nil
}

func (self *Dispatcher) listen() {
	self.server = &http.Server{Handler: self.mux, Addr: self.Connection.BindAddr()}

	var err error
	if self.Connection.Secured() {
		err = self.server.ListenAndServeTLS(self.Connection.TlsCert, self.Connection.TlsKey)
	} else {
		err = self.server.ListenAndServe()
	}

	if err != nil && err != http.ErrServerClosed {
		fmt.Printf("Error starting server: %s\n", err)
	}
}

func (self *Dispatcher) Start() {
	intrChan := make(chan os.Signal, 1)
	signal.Notify(intrChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	doneChan := make(chan bool, 1)
	go func() {
		for sig := range intrChan {
			fmt.Printf("%v\n", sig)
			switch sig {
			case syscall.SIGINT, syscall.SIGTERM:
				fmt.Printf("Shutting down server...\n")
				self.server.SetKeepAlivesEnabled(false)
				self.server.Shutdown(context.Background())
				doneChan <- true
			case syscall.SIGHUP:
				self.loadConfig()
				self.server.SetKeepAlivesEnabled(false)
				self.server.Shutdown(context.Background())
				go self.listen()
				fmt.Printf("Restarted, listening on %s\n", self.Connection.BindAddr())
			}
		}
	}()

	go self.listen()
	fmt.Printf("Listening on %s\n", self.Connection.BindAddr())
	<-doneChan
}
