include .env
export


export PROJECT_ROOT=$(shell pwd)

env-up:
	@docker compose up -d todoapp-postgres

env-down:
	@docker compose down todoapp-postgres

env-cleanup:
	@read -p "Очистить все volume файлы окружения? [y/N]: " ans; \
        if [ "$$ans" = "y" ]; then \
          docker compose down todoapp-postgres && \
          sudo rm -rf out/pgdata && \
          echo "Файлы окружение очищены"; \
        else \
          echo "Очистка отменена"; \
        fi

env-port-forward:
	@docker compose up -d port-forwarder

env-port-close:
	@docker compose down port-forwarder

migrate-create:
	@if [ -z "$(seq)" ]; then \
  		echo "Отсутствует нужный парраметр seq."; \
  		exit 1; \
  	fi; \
	docker compose run --rm todoapp-postgres-migrate \
		create \
		-ext sql \
		-dir /migrations \
		-seq "$(seq)"

migrate-up:
	@make migrate-action action=up

migrate-down:
	@make migrate-action action=down

migrate-action:
	@if [ -z "$(action)" ]; then \
  		echo "Отсутствует нужный парраметр action."; \
  		exit 1; \
  	fi; \
	docker compose run --rm todoapp-postgres-migrate \
    	-path /migrations \
		-database postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@todoapp-postgres:5432/${POSTGRES_DB}?sslmode=disable \
    	"$(action)"

test-target:
	@echo "value: $(var)"
