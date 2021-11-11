test:
	go test ./...

race:
	go test -race ./...

generate:
	rm log.go
	go generate main.go
