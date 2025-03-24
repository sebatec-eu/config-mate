GO ?= go

.PHONY: test vet
test vet:
	go $@ ./...
