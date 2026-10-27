.PHONY: help dev api web api-build web-build install clean test db-up db-down db-logs up down build logs

# Couleurs pour l'affichage
CYAN = \033[36m
GREEN = \033[32m
YELLOW = \033[33m
RESET = \033[0m

## help: Affiche la liste des commandes disponibles
help:
	@echo "$(CYAN)Commandes disponibles pour Agents DAO Platform :$(RESET)"
	@echo "  $(GREEN)make up$(RESET)         - Démarre TOUTE la stack Docker (Postgres, API Go, Web Next.js)"
	@echo "  $(GREEN)make down$(RESET)       - Arrête tous les conteneurs Docker"
	@echo "  $(GREEN)make build$(RESET)      - Reconstruit les images Docker de l'API et du Web"
	@echo "  $(GREEN)make logs$(RESET)       - Affiche les logs de tous les conteneurs"
	@echo "  $(GREEN)make db-up$(RESET)      - Démarre UNIQUEMENT la base PostgreSQL (pour dev local)"
	@echo "  $(GREEN)make db-down$(RESET)    - Arrête uniquement PostgreSQL"
	@echo "  $(GREEN)make db-logs$(RESET)    - Affiche les logs du conteneur PostgreSQL"
	@echo "  $(GREEN)make dev$(RESET)        - Démarre l'API Go et le Web Next.js en local (nécessite make db-up)"
	@echo "  $(GREEN)make install$(RESET)    - Installe les dépendances du frontend et du backend"
	@echo "  $(GREEN)make api$(RESET)        - Démarre uniquement le serveur backend Go en local"
	@echo "  $(GREEN)make web$(RESET)        - Démarre uniquement le serveur frontend Next.js en local"
	@echo "  $(GREEN)make clean$(RESET)      - Nettoie les fichiers de build temporaires"

## up: Démarre toute la stack conteneurisée (Postgres + API + Web)
up:
	@echo "$(GREEN)Démarrage de toute la stack avec Docker Compose...$(RESET)"
	docker compose up -d
	@echo "$(CYAN)Services actifs :$(RESET)"
	@echo "  • Frontend Web : http://localhost:3000"
	@echo "  • API Go       : http://localhost:8080"
	@echo "  • PostgreSQL   : localhost:5433"

## down: Stoppe tous les conteneurs
down:
	@echo "$(YELLOW)Arrêt de tous les conteneurs...$(RESET)"
	docker compose down

## build: Reconstruit les images Docker
build:
	@echo "$(YELLOW)Reconstruction des images Docker...$(RESET)"
	docker compose build

## logs: Logs de tous les conteneurs
logs:
	docker compose logs -f

## db-up: Démarre uniquement PostgreSQL dans Docker (idéal pour travailler avec make dev en local)
db-up:
	@echo "$(GREEN)Démarrage de PostgreSQL avec Docker Compose...$(RESET)"
	docker compose up -d postgres
	@echo "$(CYAN)PostgreSQL actif sur localhost:5433 (Base: agents_dao_db, User: postgres)$(RESET)"

## db-down: Stoppe PostgreSQL
db-down:
	@echo "$(YELLOW)Arrêt de PostgreSQL...$(RESET)"
	docker compose stop postgres

## db-logs: Logs de PostgreSQL
db-logs:
	docker compose logs -f postgres

## install: Installe les dépendances Go et Node.js
install:
	@echo "$(YELLOW)Installation des dépendances Go...$(RESET)"
	cd api && go mod download
	@echo "$(YELLOW)Installation des dépendances npm...$(RESET)"
	cd web && npm install

## dev: Lance à la fois l'API (Go) et le web (Next.js) en local
dev:
	@echo "$(GREEN)Démarrage simultané de l'API (port 8080) et du Web (port 3000)...$(RESET)"
	@(trap 'kill 0' SIGINT SIGTERM EXIT; \
		(cd api && go run main.go) & \
		(cd web && npm run dev) & \
		wait)

## api: Lance uniquement le serveur Go en local
api:
	@echo "$(GREEN)Démarrage du serveur Go...$(RESET)"
	cd api && go run main.go

## web: Lance uniquement le serveur Next.js en local
web:
	@echo "$(GREEN)Démarrage du serveur Next.js...$(RESET)"
	cd web && npm run dev

## api-build: Compile l'exécutable Go dans le dossier api/bin
api-build:
	@echo "$(YELLOW)Compilation de l'API Go...$(RESET)"
	mkdir -p api/bin
	cd api && go build -o bin/server main.go
	@echo "$(GREEN)Binaire généré : api/bin/server$(RESET)"

## web-build: Construit la version de production de Next.js
web-build:
	@echo "$(YELLOW)Build du frontend Next.js...$(RESET)"
	cd web && npm run build

## test: Exécute les tests Go
test:
	@echo "$(YELLOW)Exécution des tests Go...$(RESET)"
	cd api && go test -v ./...

## clean: Supprime les artefacts de build
clean:
	@echo "$(YELLOW)Nettoyage...$(RESET)"
	rm -rf api/bin
	rm -rf web/.next
