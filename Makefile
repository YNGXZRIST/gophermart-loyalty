statictest: ## Запустить statictest
	@echo "$(GREEN)Running statictest...$(NC)"
	@if [ ! -f "./bin/statictest" ]; then \
		echo "$(YELLOW)Building statictest...$(NC)"; \
		go build -o ./bin/statictest ./cmd/statictest; \
	fi
	go vet -vettool=./bin/statictest ./...
	@echo "$(GREEN)✅ statictest passed!$(NC)"