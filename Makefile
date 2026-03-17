# Исключить пакеты из тестов (например, integration): COVER_EXCLUDE='integration|/cmd/'
COVER_EXCLUDE ?=
# Список пакетов для тестов: все ./... или без COVER_EXCLUDE
TEST_PKGS := $(shell go list ./...)
ifneq ($(COVER_EXCLUDE),)
TEST_PKGS := $(shell go list ./... | grep -vE '$(COVER_EXCLUDE)')
endif

test: ## Запустить тесты (без integration)
	@echo "$(GREEN)Running tests...$(NC)"
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
