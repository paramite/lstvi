package endpoints

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/asdine/storm"
	"github.com/paramite/lstvi/models"
)

const DEFAULT_LIST_COUNT = 100

func buildJsonResponse(response http.ResponseWriter, content string, statusCode int) {
	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(statusCode)
	response.Write([]byte(content))
}

// MessageList writes last N (where N is given by the "request" parameter "count") records of messages to "response"
// in JSON format. Otherwise reports relevant response with appropriate HTTP code based on situation.
func MessageList(db *storm.DB) func(response http.ResponseWriter, request *http.Request) {
	return func(response http.ResponseWriter, request *http.Request) {
		countStr := request.URL.Query().Get("count")
		count := DEFAULT_LIST_COUNT
		if len(countStr) > 0 {
			var err error
			count, err = strconv.Atoi(countStr)
			if err != nil || (err == nil && count <= 0) {
				http.Error(response, fmt.Sprintf("Invalid count: %s. Count has to be positive integer", countStr), http.StatusBadRequest)
				return
			}
		}
		output := make([]models.Message, 0, count)
		if err := db.AllByIndex("Timestamp", &output, storm.Reverse(), storm.Limit(count)); err == nil {
			if content, err := json.Marshal(output); err == nil {
				buildJsonResponse(response, fmt.Sprintf("{\"status\": \"ok\", \"result\": %s}", string(content)), http.StatusOK)
			} else {
				buildJsonResponse(response, fmt.Sprintf("{\"status\": \"nok\", \"message\": \"Failed to convert data to JSON format: %s\"}", err.Error()), http.StatusInternalServerError)
			}
		} else {
			buildJsonResponse(response, fmt.Sprintf("{\"status\": \"nok\", \"message\": \"Failed to fetch data: %s\"}", err.Error()), http.StatusInternalServerError)
		}
	}
}

//Message saves given message to database
func Message(db *storm.DB) func(response http.ResponseWriter, request *http.Request) {
	return func(response http.ResponseWriter, request *http.Request) {
		var msg models.Message
		buffer := bytes.Buffer{}
		buffer.ReadFrom(request.Body)
		if err := json.Unmarshal(buffer.Bytes(), &msg); err == nil {
			if msg.Timestamp > 0 && len(msg.Content) > 0 {
				if err := db.Save(&msg); err == nil {
					buildJsonResponse(response, "{\"status\": \"ok\"}", http.StatusOK)
				} else {
					buildJsonResponse(response, fmt.Sprintf("{\"status\": \"nok\", \"message\": \"%s\"}", err.Error()), http.StatusInternalServerError)
				}
			} else {
				buildJsonResponse(response, fmt.Sprintf("{\"status\": \"nok\", \"message\": \"Invalid request body: %s\"}", buffer.String()), http.StatusBadRequest)
			}
		} else {
			buildJsonResponse(response, fmt.Sprintf("{\"status\": \"nok\", \"message\": \"%s\"}", err.Error()), http.StatusBadRequest)
		}
	}
}
