dev-server:
        @echo "Starting development server"
        @cd server && make run

dev-web:
        @echo "Starting development web"
        @cd web && bun run dev

docker-run:
        @if docker compose up --build 2>/dev/null; then \
                : ; \
        else \
                echo "Falling back to Docker Compose V1"; \
                docker-compose up --build; \
        fi


docker-down:
        @if docker compose down 2>/dev/null; then \
                : ; \
        else \
                echo "Falling back to Docker Compose V1"; \
                docker-compose down; \
        fi

.PHONY: dev-server dev-web docker-run docker-down%