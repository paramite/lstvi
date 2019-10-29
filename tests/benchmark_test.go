package tests

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/paramite/lstvi/endpoints"
)

const (
	BENCHMARK_LIST_COUNT = 100000
	CLIENT_COUNT         = 4
	CONF_CONTENT         = `{"bind_host": "127.0.0.1", "bind_port": 19991}`
)

func sendPOSTRequest(url string, data []byte) error {
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	defer client.CloseIdleConnections()

	if resp.StatusCode != 200 {
		body := make([]byte, 1024)
		resp.Body.Read(body)
		return fmt.Errorf("Failed to process message creation request:\n%s", string(body))
	}
	return nil
}

func GenerateTestConfig(content string) (string, error) {
	file, err := ioutil.TempFile(".", "lstvi_test_config")
	if err != nil {
		return "", err
	}
	defer file.Close()
	file.WriteString(content)
	if err != nil {
		return "", err
	}
	return file.Name(), nil
}

func BenchmarkMessageListResponse(b *testing.B) {
	config, err := GenerateTestConfig(CONF_CONTENT)
	if err != nil {
		b.Fatalf("Failed to create temporary config file: %s", err.Error())
	}
	dispatcher, err := endpoints.NewDispatcher(BENCHMARK_LIST_COUNT, config)
	if err != nil {
		b.Fatalf("Failed to load config file: %s", err.Error())
	}
	go dispatcher.Start()
	time.Sleep(time.Millisecond)

	b.Run("Benchmark message creation", func(p *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var wg sync.WaitGroup
			for c := 0; c < CLIENT_COUNT; c++ {
				wg.Add(1)
				go func() {
					endpointUrl := fmt.Sprintf("http://%s:%d/message", dispatcher.Connection.BindInterface, dispatcher.Connection.BindPort)
					for i := 0; i <= BENCHMARK_LIST_COUNT/CLIENT_COUNT; i++ {
						if err := sendPOSTRequest(endpointUrl, []byte(fmt.Sprintf(fmt.Sprintf("{\"msg\": \"xxx\", \"ts\": %d}", i+1)))); err != nil {
							b.Fatalf("Failed to create message #%d: %s", i, err)
						}
					}
					wg.Done()
				}()
			}
		}
	})

	time.Sleep(20 * time.Second)

	b.Run("Benchmark single client - message list", func(p *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			endpointUrl := fmt.Sprintf("http://%s:%d/messages?count=%d", dispatcher.Connection.BindInterface, dispatcher.Connection.BindPort, BENCHMARK_LIST_COUNT)
			resp, err := http.Get(endpointUrl)
			if err != nil {
				b.Fatal("Failed to fetch messages")
			}
			defer resp.Body.Close()
			if resp.StatusCode != 200 {
				body := make([]byte, 1024)
				resp.Body.Read(body)
				b.Fatalf("Failed to process message creation request:\n%s", string(body))
			}
		}
	})

	b.Run("Benchmark multiple clients - message list", func(p *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var wg sync.WaitGroup
			for c := 0; c < CLIENT_COUNT; c++ {
				wg.Add(1)
				go func() {
					endpointUrl := fmt.Sprintf("http://%s:%d/messages?count=%d", dispatcher.Connection.BindInterface, dispatcher.Connection.BindPort, BENCHMARK_LIST_COUNT)
					resp, err := http.Get(endpointUrl)
					if err != nil {
						b.Fatal("Failed to fetch messages")
					}
					defer resp.Body.Close()
					if resp.StatusCode != 200 {
						body := make([]byte, 1024)
						resp.Body.Read(body)
						b.Fatalf("Failed to process message creation request:\n%s", string(body))
					}
					wg.Done()
				}()
				wg.Wait()
			}
		}
	})
}
