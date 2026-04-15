include .env
export
MIGRATIONS_PATH=./migrations

# создать новую миграцию
migrate-create:
	migrate create -ext sql -dir $(MIGRATIONS_PATH) -seq name=$(name)

# применить все миграции
migrate-up:
	migrate -path $(MIGRATIONS_PATH) -database "$(DB_URL)" up

# откатить последнюю миграцию
migrate-down:
	migrate -path $(MIGRATIONS_PATH) -database "$(DB_URL)" down 1

# откатить всё (осторожно)
migrate-down-all:
	migrate -path $(MIGRATIONS_PATH) -database "$(DB_URL)" down

# показать текущую версию
migrate-version:
	migrate -path $(MIGRATIONS_PATH) -database "$(DB_URL)" version

# форс-версия (если что-то сломалось)
migrate-force:
	migrate -path $(MIGRATIONS_PATH) -database "$(DB_URL)" force $(version)