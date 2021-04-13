FROM golang:latest as appBuilder
WORKDIR /go/src/github.com/rgumi/whatsapp-mock
COPY . .
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -a -o app ./cmd/main.go


FROM busybox

RUN set -x && \
    addgroup -S app && adduser -S -G app app && \
    mkdir -p  /home/app/media /home/app/data && \
    chown -R app:app /home/app

USER app
WORKDIR /home/app

COPY --chown=app:app ./cmd/config.json /home/app
COPY --chown=app:app ./media /home/app/media
COPY --chown=app:app entrypoint.sh /home/app

RUN chmod +x ./entrypoint.sh

COPY --chown=app:app --from=appBuilder /go/src/github.com/rgumi/whatsapp-mock/app .

VOLUME [ "/home/app/data" ]

EXPOSE 8443/tcp
ENTRYPOINT ["./entrypoint.sh"]