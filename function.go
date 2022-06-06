// Package cxwh contains an example Dialogflow CX webhook
package cxwh

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

type fulfillmentInfo struct {
	Tag string `json:"tag"`
}

type sessionInfo struct {
	Session    string            `json:"session"`
	Parameters map[string]string `json:"parameters"`
}

type text struct {
	Text []string `json:"text"`
}

type responseMessage struct {
	Text text `json:"text"`
}

type fulfillmentResponse struct {
	Messages []responseMessage `json:"messages"`
}

// webhookRequest is used to unmarshal a WebhookRequest JSON object. Note that
// not all members need to be defined--just those that you need to process.
// As an alternative, you could use the types provided by the Dialogflow protocol buffers:
// https://pkg.go.dev/google.golang.org/genproto/googleapis/cloud/dialogflow/cx/v3#WebhookRequest
type webhookRequest struct {
	FulfillmentInfo fulfillmentInfo `json:"fulfillmentInfo"`
	SessionInfo     sessionInfo     `json:"sessionInfo"`
}

// webhookResponse is used to marshal a WebhookResponse JSON object. Note that
// not all members need to be defined--just those that you need to process.
// As an alternative, you could use the types provided by the Dialogflow protocol buffers:
// https://pkg.go.dev/google.golang.org/genproto/googleapis/cloud/dialogflow/cx/v3#WebhookResponse
type webhookResponse struct {
	FulfillmentResponse fulfillmentResponse `json:"fulfillmentResponse"`
	SessionInfo         sessionInfo         `json:"sessionInfo"`
}

// confirm handles webhook calls using the "joke" tag.
func joke(request webhookRequest) (webhookResponse, error) {

	t, _ := http.Get("https://v2.jokeapi.dev/joke/Any?blacklistFlags=nsfw,religious,political,racist,sexist,explicit&format=txt&type=single")
	responseData, _ := ioutil.ReadAll(t.Body)
	tt := string(responseData)
	p := map[string]string{}

	response := webhookResponse{
		FulfillmentResponse: fulfillmentResponse{
			Messages: []responseMessage{
				{
					Text: text{
						Text: []string{tt},
					},
				},
			},
		},
		SessionInfo: sessionInfo{
			Parameters: p,
		},
	}
	return response, nil
}

// confirm handles webhook calls using the "validateAccountId" tag.
func validateAccountNumber(request webhookRequest) (webhookResponse, error) {
	accountIds := [10]int{1923, 0515, 1812, 9678, 1732, 1624, 0176, 1659, 8464, 9810}
	result := false
	for _, x := range accountIds {
		if strconv.Itoa(x) == request.SessionInfo.Parameters["account-number"] {
			result = true
			break
		}
	}
	p := map[string]string{"valid-account-number": strconv.FormatBool(result)}

	response := webhookResponse{
		FulfillmentResponse: fulfillmentResponse{
			Messages: []responseMessage{
				{
					Text: text{
						Text: []string{},
					},
				},
			},
		},
		SessionInfo: sessionInfo{
			Parameters: p,
		},
	}
	return response, nil

}

// confirm handles webhook calls using the "generateAccountNumber" tag.
func generateAccountNumber(request webhookRequest) (webhookResponse, error) {
	accountIds := [10]int{1923, 0515, 1812, 9678, 1732, 1624, 0176, 1659, 8464, 9810}

	t := strconv.Itoa(accountIds[rand.Intn(9)])
	p := map[string]string{"account-number": t}

	response := webhookResponse{
		FulfillmentResponse: fulfillmentResponse{
			Messages: []responseMessage{
				{
					Text: text{
						Text: []string{},
					},
				},
			},
		},
		SessionInfo: sessionInfo{
			Parameters: p,
		},
	}
	return response, nil

}

// confirm handles webhook calls using the "greeting" tag.
func greeting(request webhookRequest) (webhookResponse, error) {

	hours, _, _ := time.Now().Clock()
	fulfillmentArray := [4]string{"How are you doing?", "How can I help you? ", "What can I do for you today?", "How can I assist? "}

	t := fmt.Sprintf("Good Morning")
	if hours > 12 {
		t = fmt.Sprintf("Good Afternoon")
	}
	if hours > 20 {
		t = fmt.Sprintf("Good Night")
	}
	t = t + "! " + fulfillmentArray[rand.Intn(3)]
	p := map[string]string{"hours": strconv.Itoa(hours)}

	response := webhookResponse{
		FulfillmentResponse: fulfillmentResponse{
			Messages: []responseMessage{
				{
					Text: text{
						Text: []string{t},
					},
				},
			},
		},
		SessionInfo: sessionInfo{
			Parameters: p,
		},
	}
	return response, nil
}

// handleError handles internal errors.
func handleError(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	fmt.Fprintf(w, "ERROR: %v", err)
}

// HandleWebhookRequest handles WebhookRequest and sends the WebhookResponse.
func HandleWebhookRequest(w http.ResponseWriter, r *http.Request) {
	var request webhookRequest
	var response webhookResponse
	var err error

	// Read input JSON
	if err = json.NewDecoder(r.Body).Decode(&request); err != nil {
		handleError(w, err)
		return
	}
	log.Printf("Request: %+v", request)

	// Get the tag from the request, and call the corresponding
	// function that handles that tag.
	// This example only has one possible tag,
	// but most agents would have many.
	switch tag := request.FulfillmentInfo.Tag; tag {
	case "generateAccountNumber":
		response, err = generateAccountNumber(request)
	case "validateAccountNumber":
		response, err = validateAccountNumber(request)
	case "greeting":
		response, err = greeting(request)
	case "joke":
		response, err = joke(request)
	default:
		err = fmt.Errorf("Unknown tag: %s", tag)
	}
	if err != nil {
		handleError(w, err)
		return
	}
	log.Printf("Response: %+v", response)

	// Send response
	if err = json.NewEncoder(w).Encode(&response); err != nil {
		handleError(w, err)
		return
	}
}
