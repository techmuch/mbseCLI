.PHONY: build build-web build-go dev run tidy install uninstall

# Full production build: frontend first (embedded via web/embed.go), then
# the Go binary that embeds it.
build: build-web build-go

build-web:
	cd web && npm install && npm run build

build-go:
	go build -o mbsecli ./cmd/mbsecli

# Build frontend and install binary into $$GOBIN / $$GOPATH/bin (or $$(PREFIX)/bin if PREFIX is set).
install: build-web
ifdef PREFIX
	mkdir -p $(PREFIX)/bin
	go build -o $(PREFIX)/bin/mbsecli ./cmd/mbsecli
else
	go install ./cmd/mbsecli
endif

uninstall:
ifdef PREFIX
	rm -f $(PREFIX)/bin/mbsecli
else
	rm -f $$(go env GOPATH)/bin/mbsecli
	if [ -n "$$(go env GOBIN)" ]; then rm -f $$(go env GOBIN)/mbsecli; fi
endif

# Run the Go server against Vite's dev server (hot-reloading UI). Run
# `cd web && npm run dev` in a second terminal alongside this.
dev:
	go run ./cmd/mbsecli start --dev examples/drone.sysml

# Run the production binary against the bundled example model.
run: build
	./mbsecli start --open examples/drone.sysml

tidy:
	go mod tidy
