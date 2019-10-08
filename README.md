### Dev `authr` app
1. goto the root directory
2. run `dep ensure`
3. create `authr-dev.config.yml` based on `config.yml` in the `configs` dir
4. execute `make run-local`

- Becnhmarks:
  go test -bench ./... -benchmem

- Coverage:
  go test -coverprofile=coverage.out ./...
  go tool cover -html=coverage.out