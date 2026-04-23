default: test

.PHONY: test
test:
	go test ./... -v $(TESTARGS) -timeout 120m

.PHONY: generate
generate:
	cd tools && go generate ./...
