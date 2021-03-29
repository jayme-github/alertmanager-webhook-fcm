package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"firebase.google.com/go/messaging"
	"github.com/julienschmidt/httprouter"
	"github.com/prometheus/alertmanager/template"
)

// JSONResponse is the standard response struct
type JSONResponse struct {
	Status  int
	Message string
}

type genericJSONRequest struct {
	Title        string
	Body         string
	Notification bool
}

func indexHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	w.Write([]byte(`<html>
	<head><title>alertmanager-webhook-fcm</title></head>
	<body>
	<h1>alertmanager-webhook-fcm</h1>
	<p><a href="/metrics">Metrics</a></p>
	<p><a href="/alert/<topic>">Alertmanager webhook</a></p>
	<p><a href="/generic/<topic>">Generic webhook</a></p>
	</body>
	</html>`))
}

func genericHandler(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
	var message *messaging.Message
	data, err := readGenericRequestBody(req)
	if err != nil {
		log.Printf("Error parsing request body: %v", err)
		sendJSONResponse(w, req, ps, http.StatusBadRequest, err.Error())
		return
	}

	topic := getParamTopic(ps)
	if data.Notification {
		message = NewNotificationMessage(topic, data.Title, data.Body)
	} else {
		message = NewDataMessage(topic, data.Title, data.Body)
	}
	msg, err := fcmClient.Send(req.Context(), message)
	fcmMessages.WithLabelValues(ps.MatchedRoutePath(), topic).Inc()

	if err != nil {
		fcmErrors.WithLabelValues(ps.MatchedRoutePath(), topic).Inc()
		sendJSONResponse(w, req, ps, http.StatusInternalServerError, err.Error())
		return
	}

	sendJSONResponse(w, req, ps,
		http.StatusOK,
		fmt.Sprintf("Message delivered to topic %s: %s", topic, msg),
	)
}

func alertHandler(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
	templateData, err := readAlertRequestBody(req)
	if err != nil {
		log.Printf("Error parsing request body: %v", err)
		sendJSONResponse(w, req, ps, http.StatusBadRequest, err.Error())
		return
	}

	topic := getParamTopic(ps)
	msg, err := processFcmMessage(req.Context(), ps.MatchedRoutePath(), topic, templateData)

	if err != nil {
		log.Printf("Error sending fcm message: %v", err)
		sendJSONResponse(w, req, ps, http.StatusInternalServerError, err.Error())
		return
	}

	sendJSONResponse(w, req, ps,
		http.StatusOK,
		fmt.Sprintf("Message delivered to topic %s: %s", topic, msg),
	)
}

// getParamTopic returns the topic from httprouter.Params or "default"
func getParamTopic(ps httprouter.Params) string {
	topic := ps.ByName("topic")
	if topic == "" {
		topic = "default"
	}
	return topic
}

func readGenericRequestBody(req *http.Request) (*genericJSONRequest, error) {
	defer req.Body.Close()
	data := genericJSONRequest{}
	err := json.NewDecoder(req.Body).Decode(&data)
	return &data, err
}

func readAlertRequestBody(req *http.Request) (*template.Data, error) {
	defer req.Body.Close()
	data := template.Data{}
	err := json.NewDecoder(req.Body).Decode(&data)
	return &data, err
}

func processFcmMessage(ctx context.Context, routePath, topic string, m *template.Data) (string, error) {
	message, err := NewMessageFromAlertmanagerData(topic, m)
	if err != nil {
		te, _ := err.(*TemplateError)
		templateErrors.WithLabelValues(te.Type).Inc()
		return "", err
	}

	// Send a message to the devices subscribed to the topic.
	fcmResponse, err := fcmClient.Send(ctx, message)
	fcmMessages.WithLabelValues(routePath, topic).Inc()
	if err != nil {
		fcmErrors.WithLabelValues(routePath, topic).Inc()
	}
	return fcmResponse, err
}

func sendJSONResponse(w http.ResponseWriter, req *http.Request, ps httprouter.Params, statusCode int, message string) {
	webhookRequests.WithLabelValues(req.Method, ps.MatchedRoutePath(), strconv.Itoa(statusCode)).Inc()

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
