build:
	go build -o protodump cmd/protodump/main.go 

fmt:
	go fmt ./...

test:
	go test -v ./...
