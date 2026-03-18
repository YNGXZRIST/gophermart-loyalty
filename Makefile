COVER_EXCLUDE ?=
TEST_EXCLUDE_PKGS ?= /cmd/accrual$$|/cmd/accrual/|/cmd/statictest$$|/cmd/statictest/|/internal/gopherman/repository/mock$$|/internal/gopherman/repository/mock/
ifeq ($(COVER_EXCLUDE),)
TEST_PKGS := $(shell go list ./... | grep -vE '$(TEST_EXCLUDE_PKGS)')
else
TEST_PKGS := $(shell go list ./... | grep -vE '$(TEST_EXCLUDE_PKGS)|$(COVER_EXCLUDE)')
endif

docker-up: ## Запустить docker-compose (БД + сервер), контейнеры с restart policy
	@echo "$(GREEN)Starting containers with docker-compose...$(NC)"
	docker compose up -d --build

docker-down: ## Остановить и удалить контейнеры docker-compose
	@echo "$(YELLOW)Stopping containers...$(NC)"
	docker compose down

docker-logs: ## Смотреть логи приложения
	docker compose logs -f app

test: ## Запустить тесты (без integration)
	@echo "$(GREEN)Running tests...$(NC)"
	@rm -f coverage.out
	go test -v -race -count=1 -coverprofile=coverage.out $(TEST_PKGS)
	@echo "$(GREEN)✅ Tests passed!$(NC)"
statictest: ## Запустить statictest
	@echo "$(GREEN)Running statictest...$(NC)"
	@if [ ! -f "./bin/statictest" ]; then \
		echo "$(YELLOW)Building statictest...$(NC)"; \
		go build -o ./bin/statictest ./cmd/statictest; \
	fi
	go vet -vettool=./bin/statictest ./...
	@echo "$(GREEN)✅ statictest passed!$(NC)"
coverage: test ## Показать покрытие
	@echo "$(GREEN)Generating coverage report...$(NC)"
	go tool cover  -html=coverage.out

coverage-percent: ## Показать общий процент покрытия
	@if [ ! -f coverage.out ]; then \
		echo "$(RED)❌ coverage.out не найден. Запустите: make test или make test-integration$(NC)"; \
		exit 1; \
	fi
	@echo "$(GREEN)Покрытие кода:$(NC)"
	@go tool cover -func=coverage.out | grep total | awk '{printf "  Всего: $(GREEN)%s$(NC)\n", $$3}'

test-coverpkg: ## Тесты с тегом integration и -coverpkg
	@echo "$(GREEN)Running tests (integration + coverpkg)...$(NC)"
	@rm -f coverage.out
	go test -v -race -count=1 -tags=integration -coverpkg="$$(printf '%s,' $(TEST_PKGS) | sed 's/,$$//')" -coverprofile=coverage.out $(TEST_PKGS)
	@echo "$(GREEN)✅ Tests passed!$(NC)"

coverage-packages: ## Показать процент покрытия по пакетам
	@if [ ! -f coverage.out ]; then \
		echo "$(RED)❌ coverage.out не найден. Запустите: make test или make test-integration$(NC)"; \
		exit 1; \
	fi
	@echo "$(GREEN)Покрытие по пакетам:$(NC)"
	@go tool cover -func=coverage.out | awk '$$0 !~ /^total/ { \
		path=$$1; sub(/:.*$$/, "", path); \
		match(path, /.*\//); pkg=(RLENGTH>0) ? substr(path, 1, RLENGTH-1) : "."; \
		gsub(/%/, "", $$3); sum[pkg]+=$$3; cnt[pkg]++ } \
		END { for (p in sum) printf "%6.1f%%  %s\n", sum[p]/cnt[p], p }' | sort -k1 -n
	@echo ""
	@echo "$(GREEN)Итого:$(NC)"
	@go tool cover -func=coverage.out | grep total | awk '{printf "  $(GREEN)%s$(NC)\n", $$3}'