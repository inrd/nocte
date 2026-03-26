APP := nocte
GO := go

GOCACHE := $(CURDIR)/.gocache
GOMODCACHE := $(CURDIR)/.gomodcache
GOENV := GOCACHE=$(GOCACHE) GOMODCACHE=$(GOMODCACHE)

.PHONY: run build fmt tidy clean

run:
	$(GOENV) $(GO) run ./cmd/nocte

build:
	$(GOENV) $(GO) build -o $(APP) ./cmd/nocte

fmt:
	$(GOENV) $(GO) fmt ./...

tidy:
	$(GOENV) $(GO) mod tidy

clean:
	rm -f $(APP)
