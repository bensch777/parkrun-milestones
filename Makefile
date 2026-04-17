.PHONY: build
build:
	mkdir -p .bin
	go build -o .bin/parkrun-milestones ./cmd/milestones/

.PHONY: vet
vet:
	go vet ./...
