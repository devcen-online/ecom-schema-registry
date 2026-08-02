# ecom-schema-registry

Registry схем событий (Protobuf/JSON Schema) с проверкой обратной
совместимости — блокирует breaking changes ещё до код-ревью (подзадача #32, FR-004).

## Схемы

`schemas/*.json` — JSON Schema событий:
- `catalog.product.updated` — создание/изменение товара PIM (→ Search projection);
- `catalog.offer.updated` — обновление оффера (Offers → Offers/Search projection);
- `catalog.offer.price_changed` — изменение цены оффера;
- `catalog.unit.updated` — обновление unit/тенанта (статусы pending/active/suspended/closed, продавцы).

## Проверка обратной совместимости

Критерии breaking (internal/breaking):
- удалено обязательное поле;
- добавлено обязательное поле;
- изменён тип поля;
- удалено значение enum;
- сужен диапазон (minimum увеличен, maximum уменьшен).

Удаление необязательного поля — warning, не блокирует.

```bash
go run ./cmd/breaking-check schemas/                 # валидация всех схем
go run ./cmd/breaking-check old.json new.json        # сравнение двух схем
```

Exit code: 0 — ок, 1 — найдены breaking-изменения.

## CI

`ci/check-schemas.yml` — GitHub Actions (проверка на каждый PR по schemas/):
build → unit-тесты → валидация → breaking-check против origin/main.

Примечание: файл лежит в `ci/`, а не `.github/workflows/`, потому что
текущий токен не имеет workflow-скоупа. Перенесите при наличии прав.

## Тесты

```bash
go test ./...
```

Покрытие: идентичные схемы без замечаний; удаление/добавление required;
изменение типа; удаление enum-значения; увеличение minimum; warning для
необязательного поля.

## Ссылки
- PRD: docs/prd/PRD-032-test-infrastructure.md, FR-004; BDD S-6..S-7.
