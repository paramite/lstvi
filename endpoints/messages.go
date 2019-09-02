package endpoints

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/asdine/storm"
	"github.com/paramite/lstvi/models"
	"goji.io/pat"
)

const DEFAULT_LIST_COUNT = 100

func report(response http.ResponseWriter, message string, code int) {
	if code == http.StatusOK {
		response.Write([]byte("{\"status\": \"ok\"}"))
	} else {
		http.Error(response, "{\"status\": \"nok\"}", code)
	}
}

// MessageList writes last N (where N is given by the "request" parameter "count") records of messages to "response"
// in JSON format. Otherwise reports relevant response with appropriate HTTP code based on situation.
func MessageList(db *storm.DB) func(response http.ResponseWriter, request *http.Request) {
	return func(response http.ResponseWriter, request *http.Request) {
		countStr := pat.Param(request, "count")
		count := DEFAULT_LIST_COUNT
		if len(countStr) > 0 {
			count, err := strconv.Atoi(countStr)
			if err != nil || (err == nil && count <= 0) {
				http.Error(response, fmt.Sprintf("Invalid count: %s. Count has to be positive integer", countStr), http.StatusBadRequest)
				return
			}
		}

		output := make([]models.Message, 0, count)
		if err := db.AllByIndex("Pk", &output, storm.Reverse(), storm.Limit(count), storm.Reverse()); err == nil {
			if content, err := json.Marshal(output); err == nil {
				response.Header().Set("Content-Type", "application/json")
				response.Write(content)
			} else {
				http.Error(response, fmt.Sprintf("Failed to convert data to JSON format: %s", err.Error()), http.StatusInternalServerError)
			}
		} else {
			http.Error(response, fmt.Sprintf("Failed to fetch data: %s", err.Error()), http.StatusInternalServerError)
		}
	}
}

//Message saves given message to database
func Message(db *storm.DB) func(response http.ResponseWriter, request *http.Request) {
	return func(response http.ResponseWriter, request *http.Request) {
		var msg models.Message
		decoder := json.NewDecoder(request.Body)
		if err := decoder.Decode(&msg); err == nil {
			if err := db.Save(&msg); err == nil {
				report(response, "", http.StatusOK)
			} else {
				report(response, fmt.Sprintf("Failed to write to DB: %s", err.Error()), http.StatusInternalServerError)
			}
		} else {
			report(response, fmt.Sprintf("Invalid message format. Failed to unmarshal: %s", err.Error()), http.StatusBadRequest)
		}
	}
}
