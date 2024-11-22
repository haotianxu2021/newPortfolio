postgres:
	docker run --name postgres17 -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=xht232431 -d postgres:17-alpine

createdb:
	docker exec -it postgres17 createdb --username=root --owner=root portfolio

dropdb:
	docker exec -it postgres17 dropdb portfolio

migrateup:
	migrate -path db/migration -database "postgresql://root:xht232431@localhost:5432/portfolio?sslmode=disable" -verbose up

migratedown:
	migrate -path db/migration -database "postgresql://root:xht232431@localhost:5432/portfolio?sslmode=disable" -verbose down

.PHONY: postgres createdb dropdb migrateup migratedown