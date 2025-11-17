.PHONY: run build test migrate migrate-up migrate-down migrate-refresh migrate-fresh clean

run:
	@go run cmd/api/main.go

build:
	@go build -o bin/api cmd/api/main.go

test:
	@go test ./...

# Migration commands
migrate:
	@go run cmd/migrate/main.go up

migrate-generate:
	@go run cmd/migrate/main.go generate

migrate-up:
	@go run cmd/migrate/main.go up

migrate-down:
	@go run cmd/migrate/main.go down

migrate-refresh:
	@go run cmd/migrate/main.go refresh

migrate-fresh:
	@go run cmd/migrate/main.go fresh

clean:
	@rm -rf bin/

dev:
	@if command -v air > /dev/null; then \
		air; \
	else \
		echo "Air not installed. Installing..."; \
		go install github.com/cosmtrek/air@latest; \
		air; \
	fi

deps:
	@go mod download
	@go mod tidy

# Documentation commands
docs:
	@go run cmd/docs/generate.go

docs-serve:
	@echo "📚 Serving documentation at http://localhost:8081"
	@python3 -m http.server 8081 -d docs/ || echo "Python not available, install python to serve docs"

# Full setup including docs
setup: deps migrate-generate migrate-up docs
	@echo "✅ Setup completed: Dependencies, migrations, and documentation generated"

	
godoc:
	@echo "📚 Generating Go documentation..."
	@go doc -all ./...

godoc-serve:
	@echo "🌐 Starting GoDoc server at http://localhost:6060"
	@echo "Open http://localhost:6060/pkg/study1/ in your browser"
	@godoc -http=:6060

doc: docs godoc