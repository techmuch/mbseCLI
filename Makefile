.PHONY: build build-web build-go dev run tidy

# Full production build: frontend first (embedded via web/embed.go), then
# the Go binary that embeds it.
build: build-web build-go

build-web:
	cd web && npm install && npm run build

build-go:
	go build -o mbsecli ./cmd/mbsecli

# Run the Go server against Vite's dev server (hot-reloading UI). Run
# `cd web && npm run dev` in a second terminal alongside this.
dev:
	go run ./cmd/mbsecli serve --dev examples/drone.sysml

# Run the production binary against the bundled example model.
run: build
	./mbsecli serve examples/drone.sysml

tidy:
	go mod tidy
