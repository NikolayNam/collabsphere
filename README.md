# collabsphere-go

Запуск для windows:
# Создание сети для внешней работы
docker network create web.network || true
# Запуска server api + local бд postgres
docker compose -p collabsphere -f docker-compose.infrastructure.yaml -f docker-compose.platform.yaml --profile local up -d --build --force-recreate
# Миграция данных 
docker compose -p collabsphere-migrate -f docker-compose.migrate.yaml up --abort-on-container-exit --exit-code-from migrate migrate
# Получить дерево файлов + содержимое 
codeweaver -input=. -output=backend-context.md -include='\.go$,\.md$,\.sql$,\.yaml$' -clipboard