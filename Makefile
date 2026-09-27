build:
	go build -o bin/protodump ./cmd/protodump

fmt:
	go fmt ./...

test:
	go test -v ./...
