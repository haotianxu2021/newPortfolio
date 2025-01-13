postgres:
	docker run --name postgres17 -p 5432:5432 -e POSTGRES_USER=${DB_USER} -e POSTGRES_PASSWORD=${DB_PASSWORD} -d postgres:17-alpine

createdb:
	docker exec -it postgres17 createdb --username=${DB_USER} --owner=${DB_USER} portfolio

dropdb:
	docker exec -it postgres17 dropdb portfolio

migrateup:
	migrate -path db/migration -database "postgresql://${DB_USER}:${DB_PASSWORD}@localhost:5432/${DB_NAME}?sslmode=disable" -verbose up

migratedown:
	migrate -path db/migration -database "postgresql://${DB_USER}:${DB_PASSWORD}@localhost:5432/${DB_NAME}?sslmode=disable" -verbose down

sqlc:
	sqlc generate

.PHONY: postgres createdb dropdb migrateup migratedown sqlc
