build:
	go build -tags protolegacy -o bin/protodump cmd/protodump/main.go

fmt:
	go fmt ./...

test:
	go test -v ./...
