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

# Documentation commands
docs:
	@go run cmd/docs/generate.go

godoc:
	@echo "📚 View Go documentation in terminal:"
	@go doc ./internal/modules/user

# Godoc serve dengan path lengkap untuk Windows
godoc-serve:
	@echo "🌐 Starting GoDoc server at http://localhost:6060"
	@echo "Open http://localhost:6060/pkg/study1/ in your browser"
	"$(shell go env GOPATH)/bin/godoc" -http=:6060

doc: docs

clean:
	@rm -rf bin/

# Development with air (fixed)
dev:
	@if command -v air > /dev/null; then \
		air; \
	else \
		echo "Air not installed. Installing correct version..."; \
		go install github.com/air-verse/air@latest; \
		$$(go env GOPATH)/bin/air; \
	fi

# Windows compatible dev
dev-win:
	@echo "🚀 Starting development server for Windows..."
	@while true; do \
		echo "Building..."; \
		go build -o bin/dev-api.exe cmd/api/main.go; \
		if [ $$? -eq 0 ]; then \
			echo "Starting server..."; \
			./bin/dev-api.exe & \
			PID=$$!; \
			echo "Waiting for file changes..."; \
			timeout /t 5 > nul; \
			taskkill /PID $$PID /F > nul 2>&1; \
			echo "Changes detected, restarting..."; \
		else \
			echo "Build failed, waiting for changes..."; \
			timeout /t 5 > nul; \
		fi; \
	done

deps:
	@go mod download
	@go mod tidy