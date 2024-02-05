EXE = deadman-listener

build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -ldflags '-extldflags "-s -w -static"' -o $(EXE) .

docker:
	docker build -t $(EXE) --platform=linux/amd64 .

clean:
	rm -f $(EXE)
