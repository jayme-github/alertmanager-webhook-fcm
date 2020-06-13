package main

import (
	"bytes"
	"fmt"
	"strings"
	tmpltext "text/template"

	"github.com/prometheus/alertmanager/template"
	"golang.org/x/net/context"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/messaging"
)

const (
	titleTemplate = `{{- if eq .Status "resolved"}}[Resolved] {{ end -}}
{{ .CommonLabels.job }} {{ .CommonLabels.alertname }}`
	bodyTemplate = `{{- $lastsummary := "" -}}
{{- range $alert := .Alerts -}}
  {{- if not (eq $lastsummary $alert.Annotations.summary) }}
  - {{ $alert.Annotations.summary -}}
  {{- $lastsummary = $alert.Annotations.summary -}}
  {{- end -}}
{{- end }}`
)

var (
	tmplTitle *tmpltext.Template
	tmplBody  *tmpltext.Template
)

func init() {
	tmplTitle = tmpltext.Must(tmpltext.New("title").Option("missingkey=zero").Parse(titleTemplate))
	tmplBody = tmpltext.Must(tmpltext.New("body").Option("missingkey=zero").Parse(bodyTemplate))
}

type TemplateError struct {
	Type string
	Err  error
}

func (e *TemplateError) Error() string {
	return fmt.Sprintf("%s expansion failed: %v", strings.Title(e.Type), e.Err)
}

// NewMessaging returns a messaging client
func NewMessaging() (*messaging.Client, error) {
	ctx := context.Background()
	app, err := firebase.NewApp(context.Background(), nil)
	if err != nil {
		return nil, err
	}

	client, err := app.Messaging(ctx)
	if err != nil {
		return nil, err
	}
	return client, err
}

// NewMessage returns a new message struct
func NewMessage(title, body string) *messaging.Message {
	return &messaging.Message{
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Data: map[string]string{
			"title":        title,
			"body":         body,
			"click_action": "FLUTTER_NOTIFICATION_CLICK",
		},
		Topic: "all",
		// https://firebase.google.com/docs/cloud-messaging/concept-options#setting-the-priority-of-a-message
		Android: &messaging.AndroidConfig{
			Priority: "high",
		},
	}
}

func NewMessageFromAlertmanagerData(m *template.Data) (*messaging.Message, error) {
	title, err := tmpltextExecuteToString(tmplTitle, m)
	if err != nil {
		return nil, &TemplateError{Type: "title", Err: err}
	}

	body, err := tmpltextExecuteToString(tmplBody, m)
	if err != nil {
		return nil, &TemplateError{Type: "body", Err: err}
	}

	return NewMessage(title, body), nil
}

func tmpltextExecuteToString(tmpl *tmpltext.Template, data interface{}) (string, error) {
	var buff bytes.Buffer
	if err := tmpl.Execute(&buff, data); err != nil {
		return "", err
	}
	return buff.String(), nil
}
