# Makefile

# 変数定義
DB_USER ?= root
DB_PASS ?= password
DB_HOST ?= localhost
DB_PORT ?= 3306
DB_NAME ?= todolist

# カラー定義
GREEN  := \033[0;32m
YELLOW := \033[0;33m
RED    := \033[0;31m
NC     := \033[0m # No Color

.PHONY: help
help: ## ヘルプを表示
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: setup
setup: ## 初期セットアップ
	@echo "$(GREEN)Setting up project...$(NC)"
	@go mod tidy
	@cp .env.example .env
	@echo "$(YELLOW)Please edit .env file with your configuration$(NC)"

.PHONY: db-create
db-create: ## データベースを作成
	@echo "$(GREEN)Creating database...$(NC)"
	@mysql -u$(DB_USER) -p$(DB_PASS) -h$(DB_HOST) -P$(DB_PORT) -e "CREATE DATABASE IF NOT EXISTS $(DB_NAME) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"

.PHONY: db-drop
db-drop: ## データベースを削除
	@echo "$(RED)Dropping database...$(NC)"
	@read -p "Are you sure? [y/N] " -n 1 -r; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		mysql -u$(DB_USER) -p$(DB_PASS) -h$(DB_HOST) -P$(DB_PORT) -e "DROP DATABASE IF EXISTS $(DB_NAME);"; \
		echo "\n$(RED)Database dropped$(NC)"; \
	fi

.PHONY: migrate-init
migrate-init: ## 初期マイグレーション実行
	@echo "$(GREEN)Running initial migration...$(NC)"
	@mysql -u$(DB_USER) -p$(DB_PASS) -h$(DB_HOST) -P$(DB_PORT) $(DB_NAME) < database/migrations/init.sql
	@echo "$(GREEN)Migration completed$(NC)"

.PHONY: migrate
migrate: ## マイグレーションを実行
	@echo "$(GREEN)Running migrations...$(NC)"
	@go run cmd/migrate/main.go -command up

.PHONY: migrate-status
migrate-status: ## マイグレーションステータスを表示
	@echo "$(GREEN)Migration status:$(NC)"
	@go run cmd/migrate/main.go -command status

.PHONY: migrate-rollback
migrate-rollback: ## マイグレーションをロールバック
	@echo "$(YELLOW)Rolling back migration...$(NC)"
	@go run cmd/migrate/main.go -command down -steps 1

.PHONY: migrate-create
migrate-create: ## 新しいマイグレーションを作成
	@read -p "Enter migration name: " name; \
	go run cmd/migrate/main.go -command create -name $$name

.PHONY: seed
seed: ## シードデータを投入
	@echo "$(GREEN)Seeding database...$(NC)"
	@mysql -u$(DB_USER) -p$(DB_PASS) -h$(DB_HOST) -P$(DB_PORT) $(DB_NAME) < database/migrations/seed_data.sql

.PHONY: db-cleanup
db-cleanup: ## 期限切れデータをクリーンアップ
	@echo "$(GREEN)Cleaning up database...$(NC)"
	@mysql -u$(DB_USER) -p$(DB_PASS) -h$(DB_HOST) -P$(DB_PORT) $(DB_NAME) < database/migrations/cleanup.sql

.PHONY: db-reset
db-reset: db-drop db-create migrate-init seed ## データベースをリセット（drop → create → migrate → seed）
	@echo "$(GREEN)Database reset completed$(NC)"

.PHONY: run
run: ## サーバーを起動
	@echo "$(GREEN)Starting server...$(NC)"
	@go run cmd/server/main.go

.PHONY: dev
dev: ## 開発サーバーを起動（ホットリロード）
	@echo "$(GREEN)Starting development server...$(NC)"
	@air

.PHONY: build
build: ## アプリケーションをビルド
	@echo "$(GREEN)Building application...$(NC)"
	@go build -o bin/server cmd/server/main.go
	@go build -o bin/migrate cmd/migrate/main.go

.PHONY: test
test: ## テストを実行
	@echo "$(GREEN)Running tests...$(NC)"
	@go test -v ./...

.PHONY: test-coverage
test-coverage: ## カバレッジ付きでテストを実行
	@echo "$(GREEN)Running tests with coverage...$(NC)"
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)Coverage report generated: coverage.html$(NC)"

.PHONY: lint
lint: ## リンターを実行
	@echo "$(GREEN)Running linter...$(NC)"
	@golangci-lint run

.PHONY: fmt
fmt: ## コードフォーマット
	@echo "$(GREEN)Formatting code...$(NC)"
	@go fmt ./...

.PHONY: clean
clean: ## ビルド成果物をクリーン
	@echo "$(GREEN)Cleaning build artifacts...$(NC)"
	@rm -rf bin/
	@rm -f coverage.out coverage.html

.PHONY: docker-up
docker-up: ## Docker Composeで起動
	@echo "$(GREEN)Starting Docker containers...$(NC)"
	@docker-compose up -d

.PHONY: docker-down
docker-down: ## Docker Composeを停止
	@echo "$(YELLOW)Stopping Docker containers...$(NC)"
	@docker-compose down

.PHONY: docker-logs
docker-logs: ## Docker ログを表示
	@docker-compose logs -f

# デフォルトターゲット
.DEFAULT_GOAL := help