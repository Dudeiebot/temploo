include .env
DB_URL=postgresql://${DB_USERNAME}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=disable
create_migration:
	migrate create -ext=sql -dir=database/migrations -seq $(name)

migrate_up:
	migrate -path=database/migrations -database "$(DB_URL)" -verbose up

migrate_down:
	migrate -path=database/migrations -database "$(DB_URL)" -verbose down

migrate_force:
	migrate -path=database/migrations -database "$(DB_URL)" force 1

dev:
	go run ./cmd/server

.PHONY: create_migration migrate_up migrate_down migrate_force dev
