package tests

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"sort"
	"testing"

	"github.com/asdine/storm"
	"github.com/paramite/lstvi/endpoints"
	"github.com/paramite/lstvi/models"
	"github.com/stretchr/testify/assert"
)

const (
	TEST_REQUEST_CONTENT_CORRECT  = `{"msg": "foobarbaz", "ts": 1566461840}`
	TEST_REQUEST_RESPONSE_CORRECT = `{"status": "ok"}`
	TEST_REQUEST_CONTENT_INVALID  = `{"msg123": "foobarbaz", "boo": 1566466640}`
	TEST_REQUEST_RESPONSE_INVALID = `{"status": "nok", "message": "Invalid request body: {"msg123": "foobarbaz", "boo": 1566466640}"}`

	TEST_LIST_COUNT = 15
)

type FakeResponseWriter struct {
	Status  *int
	Content *string
}

func (self FakeResponseWriter) Header() http.Header {
	return make(http.Header)
}

func (self FakeResponseWriter) Write(input []byte) (int, error) {
	*self.Content += string(input)
	return len(input), nil
}

func (self FakeResponseWriter) WriteHeader(statusCode int) {
	*self.Status = statusCode
}

type FakeBody struct {
	Content string
}

func (self FakeBody) Close() error {
	return nil
}

// Mimics Read of Request.Body
func (self FakeBody) Read(p []byte) (int, error) {
	actualLens := []int{len(self.Content), len(p)}
	sort.Ints(actualLens)
	for i := 0; i < actualLens[0]; i++ {
		p[i] = byte(self.Content[i])
	}
	err := io.EOF
	if len(self.Content) > actualLens[0] {
		err = nil
	}
	return actualLens[0], err
}

type MessageHandlingTestMatrix struct {
	Description string
	Body        string
	Response    string
	HttpCode    int
}

type MessageListResponse struct {
	Status string           `json:"status"`
	Result []models.Message `json:"result"`
}

func TestMessageEndpoints(t *testing.T) {
	// spawn temporary database
	tmpdir, err := ioutil.TempDir(".", "tmp_lstvi")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(tmpdir)

	t.Run("Test message save handling", func(t *testing.T) {
		db, err := storm.Open(path.Join(tmpdir, "message-save.db"))
		defer db.Close()
		if err != nil {
			t.Fatalf("Failed to open database: %s.", err)
		}
		// test response correctness on each request types
		matrix := []MessageHandlingTestMatrix{
			MessageHandlingTestMatrix{"Test case for valid request.",
				TEST_REQUEST_CONTENT_CORRECT, TEST_REQUEST_RESPONSE_CORRECT, 200},
			MessageHandlingTestMatrix{"Test case for invalid request.",
				TEST_REQUEST_CONTENT_INVALID, TEST_REQUEST_RESPONSE_INVALID, 400},
		}

		for _, testCase := range matrix {
			request, _ := http.NewRequest("POST", "http://message", FakeBody{testCase.Body})
			statusCache := 0
			responseCache := ""
			response := FakeResponseWriter{&statusCache, &responseCache}
			handler := endpoints.Message(db)
			handler(response, request)
			assert.Equalf(t, testCase.Response, *response.Content, testCase.Description)
			assert.Equal(t, testCase.HttpCode, *response.Status)
		}

		// test valid state of DB
		var msg models.Message
		err = db.One("Timestamp", 1566461840, &msg)
		if err != nil {
			t.Fatalf("Failed to fetch message from DB: %s", err)
		}
		assert.Equal(t, "foobarbaz", msg.Content)
	})

	t.Run("Test message listing", func(t *testing.T) {
		db, err := storm.Open(path.Join(tmpdir, "message-list.db"))
		defer db.Close()
		if err != nil {
			t.Fatalf("Failed to open database: %s.", err)
		}
		// create n message records
		for i := 0; i <= TEST_LIST_COUNT; i++ {
			request, _ := http.NewRequest("POST", "http://message", FakeBody{fmt.Sprintf("{\"msg\": \"xxx\", \"ts\": %d}", i)})
			statusCache := 0
			responseCache := ""
			response := FakeResponseWriter{&statusCache, &responseCache}
			handler := endpoints.Message(db)
			handler(response, request)
		}
		// request list of n-5 records
		request, _ := http.NewRequest("GET", fmt.Sprintf("http://messages?count=%d", TEST_LIST_COUNT-5), nil)
		statusCache := 0
		responseCache := ""
		response := FakeResponseWriter{&statusCache, &responseCache}
		handler := endpoints.MessageList(db)
		handler(response, request)

		list := MessageListResponse{}
		err = json.Unmarshal([]byte(responseCache), &list)
		if err != nil {
			t.Fatalf("Failed to unmarshal message list response: %s", err.Error())
		}
		assert.Equal(t, TEST_LIST_COUNT, list.Result[0].Timestamp)
		assert.Equal(t, 6, list.Result[len(list.Result)-1].Timestamp)
	})
}
