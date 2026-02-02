.PHONY: build run docker-build docker-run docker-stop clean setup

build:
	@echo "Building backend..."
	@go build -o bin/cli ./backend/cmd/cli
	@go build -o bin/api ./backend/cmd/api
	@echo "Building frontend..."
	@cd frontend && npm install && npm run build

run-api:
	@go run ./backend/cmd/api

run-cli:
	@go run ./backend/cmd/cli

docker-build:
	@if command -v docker >/dev/null 2>&1 && docker compose version >/dev/null 2>&1; then \
		echo "Using Docker Compose V2..."; \
		if docker ps >/dev/null 2>&1; then \
			docker compose build; \
		else \
			echo "Docker permission denied. Trying with sudo..."; \
			sudo docker compose build; \
		fi \
	elif command -v docker-compose >/dev/null 2>&1; then \
		echo "Using docker-compose..."; \
		if docker ps >/dev/null 2>&1; then \
			docker-compose build; \
		else \
			echo "Docker permission denied. Trying with sudo..."; \
			sudo docker-compose build; \
		fi \
	else \
		echo "Docker veya docker-compose bulunamadı. Lütfen Docker'ı yükleyin."; \
		echo "Veya yerel kurulum için: make setup && make run-api"; \
		exit 1; \
	fi

docker-run:
	@if command -v docker >/dev/null 2>&1 && docker compose version >/dev/null 2>&1; then \
		if docker ps >/dev/null 2>&1; then \
			docker compose up -d; \
		else \
			echo "Docker permission denied. Trying with sudo..."; \
			sudo docker compose up -d; \
		fi \
	elif command -v docker-compose >/dev/null 2>&1; then \
		if docker ps >/dev/null 2>&1; then \
			docker-compose up -d; \
		else \
			echo "Docker permission denied. Trying with sudo..."; \
			sudo docker-compose up -d; \
		fi \
	else \
		echo "Docker veya docker-compose bulunamadı. Lütfen Docker'ı yükleyin."; \
		echo "Veya yerel kurulum için: make setup && make run-api"; \
		exit 1; \
	fi

docker-stop:
	@if command -v docker >/dev/null 2>&1 && docker compose version >/dev/null 2>&1; then \
		if docker ps >/dev/null 2>&1; then \
			docker compose down; \
		else \
			sudo docker compose down; \
		fi \
	elif command -v docker-compose >/dev/null 2>&1; then \
		if docker ps >/dev/null 2>&1; then \
			docker-compose down; \
		else \
			sudo docker-compose down; \
		fi \
	fi

docker-logs:
	@if command -v docker >/dev/null 2>&1 && docker compose version >/dev/null 2>&1; then \
		if docker ps >/dev/null 2>&1; then \
			docker compose logs -f; \
		else \
			sudo docker compose logs -f; \
		fi \
	elif command -v docker-compose >/dev/null 2>&1; then \
		if docker ps >/dev/null 2>&1; then \
			docker-compose logs -f; \
		else \
			sudo docker-compose logs -f; \
		fi \
	fi

setup:
	@echo "Setting up for local development..."
	@cd frontend && npm install && npm run build
	@echo "Setup completed! Run 'make run-api' to start the server."

clean:
	@rm -rf bin/
	@rm -rf frontend/dist/
	@rm -rf frontend/node_modules/
	@echo "Clean completed"
