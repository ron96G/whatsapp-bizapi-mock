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

COPY ./cmd/config.json /home/app
COPY ./media /home/app/media
COPY entrypoint.sh /home/app

COPY --from=appBuilder /go/src/github.com/rgumi/whatsapp-mock/app .

VOLUME [ "/home/app/data" ]

EXPOSE 8443/tcp
ENTRYPOINT ["./entrypoint.sh"]