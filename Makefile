APP := not
GO := go

GOCACHE := $(CURDIR)/.gocache
GOMODCACHE := $(CURDIR)/.gomodcache
GOENV := GOCACHE=$(GOCACHE) GOMODCACHE=$(GOMODCACHE)

.PHONY: run build fmt tidy clean

run:
	$(GOENV) $(GO) run ./cmd/not

build:
	$(GOENV) $(GO) build -o $(APP) ./cmd/not

fmt:
	$(GOENV) $(GO) fmt ./...

tidy:
	$(GOENV) $(GO) mod tidy

clean:
	rm -f $(APP)
