start: ## Start the app
	go run ./cmd/yume/main.go

build: ## Build the app
	go build --tags "fts5" ./cmd/yume/main.go