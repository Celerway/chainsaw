test:
	go test ./...

race:
	go test -race ./...

check:
	golangci-lint run

generate:
	rm -f log.go
	go generate main.go
