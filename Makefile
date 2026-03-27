APP := nocte
GO := go
BINDIR ?= $(HOME)/.local/bin

GOCACHE := $(CURDIR)/.gocache
GOMODCACHE := $(CURDIR)/.gomodcache
GOENV := GOCACHE=$(GOCACHE) GOMODCACHE=$(GOMODCACHE)

.PHONY: run build install test fmt tidy clean release

run:
	$(GOENV) $(GO) run ./cmd/nocte

build:
	$(GOENV) $(GO) build -o $(APP) ./cmd/nocte

install:
	mkdir -p $(BINDIR)
	$(GOENV) $(GO) build -o $(BINDIR)/$(APP) ./cmd/nocte

test:
	$(GOENV) $(GO) test ./...

fmt:
	$(GOENV) $(GO) fmt ./...

tidy:
	$(GOENV) $(GO) mod tidy

clean:
	rm -f $(APP)

release:
	@test -n "$(VERSION)" || (echo "VERSION is required, for example: make release VERSION=0.3.1"; exit 1)
	sh ./scripts/release.sh $(VERSION) $(if $(filter 1 true yes,$(PUSH)),--push,)
