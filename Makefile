build:
	go build -o protodump cmd/main.go 

test:
	go test -v ./...
