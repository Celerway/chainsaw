test:
	rm log.go
	go generate main.go
	go test ./...

race:
	go test -race ./...

