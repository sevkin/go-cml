all:

test:
	go test -timeout 30s ./...

cmlcli:
	cd cmd/cmlcli && go run ./...
cmlcli-debug:
	cd cmd/cmlcli && go run ./... -resty-debug

cml2csv:
	@cd cmd/cml2csv && go run ./...