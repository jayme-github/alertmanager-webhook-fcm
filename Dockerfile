FROM scratch

COPY build/alertmanager-webhook-fcm.linux.amd64 /

EXPOSE 9716

ENTRYPOINT ["/alertmanager-webhook-fcm.linux.amd64"]
