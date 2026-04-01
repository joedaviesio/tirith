.PHONY: build clean test frontend

VERSION ?= 0.1.0-dev
BINARY := tirith

frontend:
	cd frontend && npm run build
	rm -rf internal/dashboard/frontend/*
	cp -r frontend/out/* internal/dashboard/frontend/

build: frontend
	go build -ldflags "-X main.version=$(VERSION)" -o $(BINARY) ./cmd/tirith

build-go:
	go build -ldflags "-X main.version=$(VERSION)" -o $(BINARY) ./cmd/tirith

test:
	go test ./...

clean:
	rm -f $(BINARY)

.PHONY: run-start run-report
run-start: build
	./$(BINARY) start

run-report: build-go
	./$(BINARY) report
