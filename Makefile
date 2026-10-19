BINARY := bin/ko4fun-clientless

.PHONY: build run test vet fmt clean

build:
	go build -trimpath -ldflags "-s -w" -o $(BINARY) ./cmd/ko4fun-clientless

run:
	go run ./cmd/ko4fun-clientless

test:
	go test -race ./...

vet:
	go vet ./...

fmt:
	gofmt -w .

clean:
	rm -rf bin
