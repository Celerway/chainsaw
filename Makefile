test:
	go test ./...

race:
	go test -race ./...

generate:
	rm -f log.go
	go generate main.go
