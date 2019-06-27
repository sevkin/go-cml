all:

test:
	go test -timeout 30s ./...

cmlcli:
	cd cmd/cmlcli && go run main.go pinger.go
cmlcli-debug:
	cd cmd/cmlcli && go run main.go pinger.go -resty-debug
