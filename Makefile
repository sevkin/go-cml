all:

test:
	go test ./...

cmlcli:
	cd cmd/cmlcli && go run main.go pinger.go
