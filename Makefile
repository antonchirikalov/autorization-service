cover:
	rm -f coverage.out
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

dist:
	if test -d vendor; then dep ensure -update; else dep ensure; fi
	rm -fr dist
	GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o dist/authorization-service cmd/authr/main.go

clean:
	rm -fr coverage.out dist

run-local:
	go run cmd/authr/main.go --config=configs/authr-dev.config.yml
