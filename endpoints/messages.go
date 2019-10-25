package endpoints

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/paramite/lstvi/memcache"
)

const DEFAULT_LIST_COUNT = 100

func buildJsonResponse(response http.ResponseWriter, content string, statusCode int) {
	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(statusCode)
	response.Write([]byte(content))
}

// MessageList writes last N (where N is given by the "request" parameter "count") records of messages to "response"
// in JSON format. Otherwise reports relevant response with appropriate HTTP code based on situation.
func MessageList(cache *memcache.MessageCache) func(response http.ResponseWriter, request *http.Request) {
	return func(response http.ResponseWriter, request *http.Request) {
		countStr := request.URL.Query().Get("count")
		count := DEFAULT_LIST_COUNT
		if len(countStr) > 0 {
			var err error
			count, err = strconv.Atoi(countStr)
			if err != nil || (err == nil && count <= 0) {
				buildJsonResponse(response, fmt.Sprintf("{\"status\": \"nok\", \"message\": \"Invalid count: %s. Count has to be positive integer\"}", countStr), http.StatusBadRequest)
				return
			}
		}
		output := cache.GetLast(count)
		if content, err := json.Marshal(output); err == nil {
			buildJsonResponse(response, fmt.Sprintf("{\"status\": \"ok\", \"result\": %s}", string(content)), http.StatusOK)
		} else {
			buildJsonResponse(response, fmt.Sprintf("{\"status\": \"nok\", \"message\": \"Failed to convert data to JSON format: %s\"}", err.Error()), http.StatusInternalServerError)
		}
	}
}

//Message saves given message to cache
func Message(queue chan memcache.Message) func(response http.ResponseWriter, request *http.Request) {
	return func(response http.ResponseWriter, request *http.Request) {
		var msg memcache.Message
		buffer := bytes.Buffer{}
		buffer.ReadFrom(request.Body)
		if err := json.Unmarshal(buffer.Bytes(), &msg); err == nil {
			if msg.Timestamp > 0 && len(msg.Content) > 0 {
				queue <- msg
				buildJsonResponse(response, "{\"status\": \"ok\"}", http.StatusOK)
			} else {
				buildJsonResponse(response, fmt.Sprintf("{\"status\": \"nok\", \"message\": \"Invalid request body: %s\"}", buffer.String()), http.StatusBadRequest)
			}
		} else {
			buildJsonResponse(response, fmt.Sprintf("{\"status\": \"nok\", \"message\": \"%s\"}", err.Error()), http.StatusBadRequest)
		}
	}
}
