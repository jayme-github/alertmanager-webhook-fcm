package main

import (
	"flag"
	"log"
	"net/http"

	"firebase.google.com/go/messaging"
	"github.com/julienschmidt/httprouter"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	flagListen = flag.String("listen", ":9716", "[ip]:port to listen on for HTTP")
	fcmClient  *messaging.Client

	webhookRequests = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "alertmanager_webhook_requests_total",
			Help: "Total number of HTTP requests.",
		},
		[]string{"method", "handler", "status"},
	)
	templateErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "alertmanager_webhook_template_errors_total",
			Help: "Total number of errors executing message templates.",
		},
		[]string{"type"},
	)
	fcmErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "alertmanager_webhook_fcm_errors_total",
			Help: "Total number of errors talking to Firebase Cloud Messaging API.",
		},
		[]string{"handler", "topic"},
	)
)

func main() {
	var err error
	flag.Parse()

	fcmClient, err = NewMessaging()
	if err != nil {
		log.Fatalf("error getting Messaging client. Do you have GOOGLE_APPLICATION_CREDENTIALS set?: %v\n", err)
	}

	router := httprouter.New()
	router.SaveMatchedRoutePath = true
	router.GET("/", indexHandler)
	router.POST("/alert/:topic", alertHandler)
	router.POST("/generic/:topic", genericHandler)
	router.Handler("GET", "/metrics", promhttp.Handler())
	log.Printf("Listening on: %s\n", *flagListen)
	log.Fatal(http.ListenAndServe(*flagListen, router))
}
