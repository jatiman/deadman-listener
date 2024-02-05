FROM golang:alpine as builder

WORKDIR /tmp/app
COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -ldflags '-extldflags "-s -w -static"' -o deadman-listener .

FROM scratch
COPY --from=builder /tmp/app/deadman-listener /usr/local/bin/deadman-listener

EXPOSE 9095

ENTRYPOINT ["/usr/local/bin/deadman-listener"]
