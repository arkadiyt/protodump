build:
	go build -o bin/protodump cmd/protodump/main.go

fmt:
	go fmt ./...

test:
	go test -v ./...
