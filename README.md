# collabsphere

Запуск для windows:
# Создание сети для внешней работы
docker network create web.network || true
# Запуска server api + local бд postgres
docker compose -p collabsphere -f docker-compose.infrastructure.yaml -f docker-compose.platform.yaml --profile local up -d --build --force-recreate
# Миграция данных 
docker compose -p collabsphere-migrate -f docker-compose.migrate.yaml up --abort-on-container-exit --exit-code-from migrate migrate
# Получить дерево файлов + содержимое
del /f /q .\docs\codebase_actual.md 2>nul
codeweaver -input=. -output=./docs/codebase_actual.md -include="\.go$,\.md$,\.sql$,\.yaml$" -ignore="^\.git,^docs/"
# Загрузить пакет для управления make
go install github.com/charmbracelet/gum@latest
