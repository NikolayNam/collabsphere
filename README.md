# collabsphere — README для разработчиков

## Цель
Этот документ описывает **локальный запуск**, **миграции**, **debug** и типовые проблемы разработки для `collabsphere`.

---

## Требования

### Общие
- Docker + Docker Compose (Docker Desktop на Windows)
- Go (версия как в `go.mod`)
- `codeweaver` (опционально, для выгрузки codebase)
- `gum` (опционально, если используется в Makefile/скриптах)

### Linux / WSL
- `make`
- **Makefile должен быть в формате LF** (не CRLF)

---

## Быстрый старт

### Linux / WSL (рекомендуемый режим)
```bash
# 1) сеть
make network

# 2) platform + postgres
make up-dev

# 3) миграции
make --trace migrate-up
```

Откат миграций:
```bash
make --trace migrate-down
```

---

### Windows (Docker Desktop, PowerShell/CMD)

Создание сети:
```powershell
docker network create web.network 2>$null
```

Запуск server API + local Postgres:
```powershell
docker compose -p collabsphere `
-f docker-compose.infrastructure.yaml `
-f docker-compose.platform.yaml `
--profile local up -d --build --force-recreate
```

Миграции:
```powershell
docker compose -p collabsphere-migrate `
-f docker-compose.migrate.yaml `
up --abort-on-container-exit --exit-code-from migrate migrate
```

---

## Миграции БД (goose)

Миграции лежат здесь:
- `platform/internal/runtime/infrastructure/db/migrations/`

### Важно: DO $$ ... $$;
Если в миграции используется PL/pgSQL блок `DO $$ ... $$;`, то его нужно **обязательно** оборачивать, иначе `goose` может разрезать SQL по `;` и упасть с ошибкой вида *unterminated dollar-quoted string*:

```sql
-- +goose StatementBegin
DO $$
BEGIN
-- ...
END
$$;
-- +goose StatementEnd
```

### Проверка существования таблицы (пример)
Если нужно “застраховать” порядок миграций:
```sql
-- +goose StatementBegin
DO $$
BEGIN
IF to_regclass('organizations') IS NULL THEN
RAISE EXCEPTION 'organizations table does not exist; run organizations migration first';
END IF;
END
$$;
-- +goose StatementEnd
```w

> Если у вас в контейнере выставлен `search_path=db,public`, то `to_regclass('organizations')` проверит таблицу по `search_path`.
> Если хотите строго — используйте `to_regclass('db.organizations')`.

---

## Команды Makefile (Linux/WSL)

### Сеть
```bash
make network
```

### Запуск окружения разработки
```bash
make up-dev
```

### Миграции
```bash
make --trace migrate-up
make --trace migrate-down
```

### Нормальный/подробный вывод make
Если нужно увидеть, какие команды реально выполняются:
```bash
make --trace migrate-up
```

Показать команды без выполнения:
```bash
make -n migrate-up
```

Максимально подробный debug make:
```bash
make --debug=v migrate-up
```

---

## Логи и диагностика Docker

Показать статус контейнеров:
```bash
docker compose -p collabsphere ps
```

Логи:
```bash
docker compose -p collabsphere logs -f
```

Логи мигратора (если запускали отдельным проектом):
```bash
docker compose -p collabsphere-migrate logs -f
```

---

## Генерация “дерева файлов + содержимого” (codeweaver)

### Windows
```powershell
del /f /q .\docs\codebase_actual.md 2>$null
codeweaver -input=. -output=./docs/codebase_actual.md -include="\.go$,\.md$,\.sql$,\.yaml$" -ignore="^\.git,^docs/"
```

### Linux / WSL
```bash
rm -f ./docs/codebase_actual.md
codeweaver -input=. -output=./docs/codebase_actual.md \
-include="\.go$,\.md$,\.sql$,\.yaml$" \
-ignore="^\.git,^docs/"
```

Полезное ужесточение

Если хочешь, можно ещё добавить проверку, что в migrations/ никто не редактирует файлы руками, например через CI:

сначала запускаешь build-migrations.sh

потом git diff --exit-code

Если diff есть — значит bundle не пересобран.

Пример для CI/local check:

./scripts/build-migrations.sh
git diff --exit-code -- platform/internal/runtime/infrastructure/db/migrations

---

## Утилиты

### gum
```bash
go install github.com/charmbracelet/gum@latest
```

---

### Makefile CRLF
Если `make` ведёт себя странно на Linux/WSL — проверь, что `Makefile` в LF.

---

## Архитектура (кратко)

Проект организован по модулям и слоям:
