BINARY := dooing-tmux
GO     := go

.PHONY: build run test lint clean

build:
	$(GO) build -o $(BINARY) .

run:
	$(GO) run .

test:
	$(GO) test ./...

lint:
	golangci-lint run ./...

clean:
	rm -f $(BINARY)
