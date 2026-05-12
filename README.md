# Web-Backgammon

Онлайн-игра в длинные нарды для двух игроков.

## Стек

- **Frontend:** Next.js 14 + TypeScript + Tailwind CSS + Zustand + Framer Motion
- **Backend:** Go 1.22+ + chi + nhooyr.io/websocket + pgx/v5
- **Database:** PostgreSQL 16
- **Deploy:** Docker Compose

## Структура

- `backend/` — Go-сервер (REST + WebSocket).
- `frontend/` — Next.js приложение.
- `docs/` — спецификация и планы разработки.
- `docker-compose.yml` — оркестрация контейнеров.

## Запуск (разработка)

```bash
cp .env.example .env
docker compose up -d postgres
cd backend && go test ./...
cd frontend && npm install && npm run dev
```

См. [docs/specs/backgammon-design.md](docs/specs/backgammon-design.md) для полной спецификации.
