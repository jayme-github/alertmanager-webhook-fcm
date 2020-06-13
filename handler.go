package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/prometheus/alertmanager/template"
)

// JSONResponse is the standard response struct
type JSONResponse struct {
	Status  int
	Message string
}

func genericHandler(w http.ResponseWriter, req *http.Request) {
	// TODO: Allow to post title and message
	message := NewMessage("Test message", "A test message body")
	msg, err := fcmClient.Send(req.Context(), message)
	if err != nil {
		fcmErrors.Inc()
		sendJSONResponse(w, req, http.StatusInternalServerError, err.Error())
		return
	}
	sendJSONResponse(w, req, http.StatusOK, fmt.Sprintf("Message delivered: %s", msg))
}

func alertHandler(w http.ResponseWriter, req *http.Request) {
	templateData, err := readRequestBody(req)
	if err != nil {
		log.Printf("Error parsing request body: %v", err)
		sendJSONResponse(w, req, http.StatusBadRequest, err.Error())
		return
	}

	msg, err := processFcmMessage(req.Context(), templateData)

	if err != nil {
		log.Printf("Error sending fcm message: %v", err)
		sendJSONResponse(w, req, http.StatusInternalServerError, err.Error())
		return
	}

	sendJSONResponse(w, req, http.StatusOK, fmt.Sprintf("Message delivered: %s", msg))
}

func readRequestBody(req *http.Request) (*template.Data, error) {
	defer req.Body.Close()
	data := template.Data{}
	err := json.NewDecoder(req.Body).Decode(&data)
	return &data, err
}

func processFcmMessage(ctx context.Context, m *template.Data) (string, error) {
	message, err := NewMessageFromAlertmanagerData(m)
	if err != nil {
		te, _ := err.(*TemplateError)
		templateErrors.WithLabelValues(te.Type).Inc()
		return "", err
	}

	// Send a message to the devices subscribed to the topic.
	fcmResponse, err := fcmClient.Send(ctx, message)
	if err != nil {
		fcmErrors.Inc()
	}
	return fcmResponse, err
}

func sendJSONResponse(w http.ResponseWriter, req *http.Request, statusCode int, message string) {
	webhookRequests.WithLabelValues(req.Method, req.URL.Path, strconv.Itoa(statusCode)).Inc()

	response, _ := json.Marshal(JSONResponse{
		Status:  statusCode,
		Message: message,
	})
	w.WriteHeader(statusCode)
	_, err := w.Write(response)

	if err != nil {
		log.Printf("Error sending response to client: %v", err)
	}
}
