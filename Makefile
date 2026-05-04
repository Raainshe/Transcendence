COMPOSE := docker compose

.PHONY: up build down restart restart-build logs ps clean help

up:
	$(COMPOSE) up -d

build:
	$(COMPOSE) up -d --build

down:
	$(COMPOSE) down

restart:
	$(COMPOSE) down
	$(COMPOSE) up -d

restart-build:
	$(COMPOSE) down
	$(COMPOSE) up -d --build

logs:
	$(COMPOSE) logs -f

ps:
	$(COMPOSE) ps

clean:
	$(COMPOSE) down -v

help:
	@echo "make up            - start backend + frontend (no rebuild)"
	@echo "make build         - start backend + frontend with rebuild"
	@echo "make restart       - stop and start (no rebuild)"
	@echo "make restart-build - stop and start with rebuild"
	@echo "make down          - stop and remove containers"
	@echo "make logs          - tail logs from all services"
	@echo "make clean         - down + remove volumes (wipes db)"
