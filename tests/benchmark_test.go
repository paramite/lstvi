package tests

import (
	"bytes"
	"fmt"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/paramite/lstvi/endpoints"
)

const (
	BENCHMARK_LIST_COUNT = 100000
	BIND_IF              = "127.0.0.1"
	BIND_PORT            = 19991
	CLIENT_COUNT         = 4
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

func BenchmarkMessageListResponse(b *testing.B) {
	go endpoints.RunDispatcher(BENCHMARK_LIST_COUNT, BIND_IF, BIND_PORT, "", "")
	time.Sleep(time.Millisecond)

	b.Run("Benchmark message creation", func(p *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var wg sync.WaitGroup
			for c := 0; c < CLIENT_COUNT; c++ {
				wg.Add(1)
				go func() {
					for i := 0; i <= BENCHMARK_LIST_COUNT/CLIENT_COUNT; i++ {
						if err := sendPOSTRequest(fmt.Sprintf("http://%s:%d/message", BIND_IF, BIND_PORT), []byte(fmt.Sprintf(fmt.Sprintf("{\"msg\": \"xxx\", \"ts\": %d}", i+1)))); err != nil {
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
			resp, err := http.Get(fmt.Sprintf("http://%s:%d/messages?count=%d", BIND_IF, BIND_PORT, BENCHMARK_LIST_COUNT))
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
					resp, err := http.Get(fmt.Sprintf("http://%s:%d/messages?count=%d", BIND_IF, BIND_PORT, BENCHMARK_LIST_COUNT))
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
