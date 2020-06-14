# alertmanager-webhook-fcm
Prometheus Alertmanager receiver that sends messages via Firebase Cloud Messaging

[![Build Status](https://travis-ci.org/jayme-github/alertmanager-webhook-fcm.svg?branch=master)](https://travis-ci.org/jayme-github/alertmanager-webhook-fcm) ![Docker Pulls](https://img.shields.io/docker/pulls/jaymedh/alertmanager-webhook-fcm)

# Usage
## Alertmanager API endpoint
Configure the `/alert/<topic>` API endpoint in `alertmanager.yml`.

You may configure this webhook in multiple receivers push messages to different Firebase Cloud Messaging topics:
```yaml
receivers:
- name: 'webhook.all'
  webhook_configs:
  - url: 'http://alertmanager-webhook-fcm:9716/alert/all'
- name: 'webhook.important'
  webhook_configs:
  - url: 'http://alertmanager-webhook-fcm:9716/alert/important'
```

## Generic API endpoint
There is a generic API endpoint that can be used to send simple notifications. It accepts a less comlex JSON format then alertmanager uses:

```bash
curl --header "Content-Type: application/json" \
  --request POST \
  --data '{"title":"The message title", "body":"The message body"}' \
  http://localhost:9716/generic/your-topic
```

## Permissions
This webhook needs `cloudmessaging.messages.create` permission for your Firebase project.

You may want to create a custom IAM role with only this permission and assign a new service account to it. Download the service account JSON file and set the environment variable `GOOGLE_APPLICATION_CREDENTIALS` to its absolute path.

## Docker (compose)
The container will need access to some root CA to verify Google's SSL certificates. You may want to mount your host certificates, llike:
```yaml
version: "2"
services:
  alertmanager-webhook-fcm:
    container_name: alertmanager-webhook-fcm
    image: jaymedh/alertmanager-webhook-fcm:v0.2
    ports:
      - 9716:9716
    environment:
      - GOOGLE_APPLICATION_CREDENTIALS=/etc/serviceAccountKey.json
    volumes:
      - /etc/ssl/certs:/etc/ssl/certs:ro
      - /serviceAccountKey.json:/etc/serviceAccountKey.json:ro
    restart: unless-stopped
```
