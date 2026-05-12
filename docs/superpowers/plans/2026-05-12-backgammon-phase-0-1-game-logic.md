# Web-Backgammon — Plan 1: Scaffolding + Game Logic (длинные нарды)

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Создать структуру monorepo (`frontend/`, `backend/`, `docker-compose.yml`) и реализовать пакет `internal/game` с полной логикой длинных нарды по TDD — все правила, edge cases, покрытие >90%.

**Architecture:** Phase 0 — минимальный scaffolding всех трёх компонентов проекта (Next.js, Go, PostgreSQL в Docker). Phase 1 — pure-Go пакет `game` без внешних зависимостей. Никаких баз данных, сети или UI на этом этапе — только детерминированный игровой движок, который можно использовать из любого транспорта.

**Tech Stack:** Go 1.22+ (стандартная библиотека + `testify` для тестов), Next.js 14 (scaffolding), PostgreSQL 16 (только Docker-контейнер, схема — в Plan 2), Docker Compose.

**Ссылки на спецификацию:** [docs/specs/backgammon-design.md](../../specs/backgammon-design.md) — секции 1, 2, 13.

**Соглашения о координатах доски:**
- Пункты пронумерованы 1..24.
- Белые шашки стартуют на пункте 24, движутся 24 → 1, выкидываются с пунктов 1..6 (дом белых).
- Чёрные шашки стартуют на пункте 1, движутся 1 → 24, выкидываются с пунктов 19..24 (дом чёрных).
- Формула: для белых `to = from - die`; для чёрных `to = from + die`.
- `to == 0` (белые) или `to == 25` (чёрные) → bear off (выкидывание).

---

## File Structure

После выполнения этого плана:

```
Web-backgammon/
├── .gitignore                         # Корневой gitignore
├── README.md                          # Описание проекта + запуск
├── docker-compose.yml                 # PostgreSQL контейнер (только)
├── .env.example                       # Шаблон ENV для compose
├── backend/
│   ├── go.mod                         # github.com/yandex/web-backgammon (модуль)
│   ├── go.sum
│   ├── cmd/server/main.go             # Заглушка main (просто print "TODO Plan 2")
│   └── internal/game/
│       ├── color.go                   # Color enum, Direction
│       ├── board.go                   # Board struct, NewBoard, Point
│       ├── board_test.go
│       ├── move.go                    # Move struct, Apply method
│       ├── move_test.go
│       ├── dice.go                    # Dice interface + RandomDice + ExpandDice
│       ├── dice_test.go
│       ├── rules.go                   # Validation predicates (direction, occupation, zabor)
│       ├── rules_test.go
│       ├── moves.go                   # GenerateMoves, GenerateSequences, mandatory rules
│       ├── moves_test.go
│       ├── bearoff.go                 # AllInHome, BearOff validation
│       ├── bearoff_test.go
│       ├── game.go                    # Game state machine, public API
│       └── game_test.go
└── frontend/
    ├── package.json                   # Next.js 14 + Tailwind + TypeScript
    ├── tsconfig.json
    ├── next.config.mjs
    ├── tailwind.config.ts
    ├── postcss.config.mjs
    ├── .gitignore
    └── src/app/
        ├── layout.tsx                 # Минимальный layout
        ├── page.tsx                   # Главная страница с placeholder
        └── globals.css                # Tailwind directives
```

**Сборка проверяется командами:**
- `cd backend && go build ./... && go test ./...`
- `cd frontend && npm run build`
- `docker compose config` (валидация compose-файла)

---

## Phase 0: Scaffolding (Tasks 0.1 — 0.6)

### Task 0.1: Корневая структура проекта

**Files:**
- Create: `/mnt/ForAllOS/YD/GoProjects/Games/Web-backgammon/.gitignore`
- Create: `/mnt/ForAllOS/YD/GoProjects/Games/Web-backgammon/README.md`
- Create: `/mnt/ForAllOS/YD/GoProjects/Games/Web-backgammon/.env.example`

- [ ] **Step 1: Удалить устаревший `main.go` и `go.mod` из корня**

Корневой шаблон GoLand больше не нужен — Go будет жить в `backend/`.

```bash
cd /mnt/ForAllOS/YD/GoProjects/Games/Web-backgammon
rm -f main.go go.mod
```

- [ ] **Step 2: Создать корневой `.gitignore`**

```
# OS
.DS_Store
Thumbs.db

# IDE
.idea/
.vscode/
*.swp

# Env
.env
.env.local
.env.*.local
!.env.example

# Build artifacts
backend/bin/
backend/dist/
frontend/.next/
frontend/out/
frontend/node_modules/

# Test artifacts
coverage.out
coverage.html
*.test
```

- [ ] **Step 3: Создать `README.md`**

```markdown
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
```

- [ ] **Step 4: Создать `.env.example`**

```
# PostgreSQL
POSTGRES_DB=backgammon
POSTGRES_USER=bg_user
POSTGRES_PASSWORD=change_me_in_production

# Backend
PORT=8080
DATABASE_URL=postgres://bg_user:change_me_in_production@localhost:5432/backgammon?sslmode=disable
JWT_SECRET=change_me_to_random_64_char_string
ALLOWED_ORIGINS=http://localhost:3000

# Frontend
NEXT_PUBLIC_API_URL=http://localhost:8080
NEXT_PUBLIC_WS_URL=ws://localhost:8080
```

- [ ] **Step 5: Проверить структуру и зафиксировать**

```bash
ls -la /mnt/ForAllOS/YD/GoProjects/Games/Web-backgammon/
git init && git add . && git commit -m "chore: initial project skeleton with gitignore and env template"
```

Expected: видны `.gitignore`, `README.md`, `.env.example`, `docs/`; нет `main.go` и `go.mod` в корне.

---

### Task 0.2: Инициализация Go-модуля backend

**Files:**
- Create: `backend/go.mod`
- Create: `backend/cmd/server/main.go`

- [ ] **Step 1: Создать структуру и инициализировать модуль**

```bash
mkdir -p backend/cmd/server backend/internal/game
cd backend
go mod init github.com/denis/web-backgammon
```

Имя модуля можно изменить под фактический owner репозитория, но используйте одинаковое имя во всех `import` в этом плане.

- [ ] **Step 2: Создать заглушку `cmd/server/main.go`**

```go
package main

import "fmt"

func main() {
	fmt.Println("backgammon server: server implementation is in Plan 2 (REST + WS)")
}
```

- [ ] **Step 3: Добавить `testify` зависимость**

```bash
cd backend
go get github.com/stretchr/testify@v1.10.0
go mod tidy
```

- [ ] **Step 4: Проверить компиляцию**

Run: `cd backend && go build ./...`
Expected: успешная сборка, артефактов в репозитории не остаётся (gitignore).

Run: `cd backend && go run ./cmd/server`
Expected output: `backgammon server: server implementation is in Plan 2 (REST + WS)`

- [ ] **Step 5: Commit**

```bash
git add backend/
git commit -m "feat(backend): initialize Go module with testify and stub main"
```

---

### Task 0.3: Инициализация Next.js frontend

**Files:**
- Create: `frontend/package.json`
- Create: `frontend/tsconfig.json`
- Create: `frontend/next.config.mjs`
- Create: `frontend/tailwind.config.ts`
- Create: `frontend/postcss.config.mjs`
- Create: `frontend/.gitignore`
- Create: `frontend/src/app/layout.tsx`
- Create: `frontend/src/app/page.tsx`
- Create: `frontend/src/app/globals.css`

- [ ] **Step 1: Создать `frontend/package.json`**

```json
{
  "name": "web-backgammon-frontend",
  "version": "0.1.0",
  "private": true,
  "scripts": {
    "dev": "next dev",
    "build": "next build",
    "start": "next start",
    "lint": "next lint",
    "typecheck": "tsc --noEmit"
  },
  "dependencies": {
    "next": "^14.2.5",
    "react": "^18.3.1",
    "react-dom": "^18.3.1"
  },
  "devDependencies": {
    "@types/node": "^22.0.0",
    "@types/react": "^18.3.3",
    "@types/react-dom": "^18.3.0",
    "autoprefixer": "^10.4.19",
    "eslint": "^9.6.0",
    "eslint-config-next": "^14.2.5",
    "postcss": "^8.4.39",
    "tailwindcss": "^3.4.4",
    "typescript": "^5.5.3"
  }
}
```

- [ ] **Step 2: Создать конфиги TypeScript, Next, Tailwind, PostCSS**

`frontend/tsconfig.json`:
```json
{
  "compilerOptions": {
    "target": "ES2022",
    "lib": ["dom", "dom.iterable", "esnext"],
    "allowJs": false,
    "skipLibCheck": true,
    "strict": true,
    "noEmit": true,
    "esModuleInterop": true,
    "module": "esnext",
    "moduleResolution": "bundler",
    "resolveJsonModule": true,
    "isolatedModules": true,
    "jsx": "preserve",
    "incremental": true,
    "plugins": [{ "name": "next" }],
    "baseUrl": ".",
    "paths": { "@/*": ["./src/*"] }
  },
  "include": ["next-env.d.ts", "**/*.ts", "**/*.tsx", ".next/types/**/*.ts"],
  "exclude": ["node_modules"]
}
```

`frontend/next.config.mjs`:
```js
/** @type {import('next').NextConfig} */
const nextConfig = {
  reactStrictMode: true,
  output: "standalone"
};
export default nextConfig;
```

`frontend/tailwind.config.ts`:
```ts
import type { Config } from "tailwindcss";

const config: Config = {
  content: ["./src/**/*.{ts,tsx}"],
  theme: {
    extend: {
      colors: {
        neo: {
          bg: "#e0e5ec",
          light: "#ffffff",
          dark: "#a3b1c6",
          accent: "#6c63ff"
        }
      }
    }
  },
  plugins: []
};
export default config;
```

`frontend/postcss.config.mjs`:
```js
export default {
  plugins: { tailwindcss: {}, autoprefixer: {} }
};
```

`frontend/.gitignore`:
```
node_modules/
.next/
out/
*.tsbuildinfo
next-env.d.ts
```

- [ ] **Step 3: Создать минимальный App Router**

`frontend/src/app/globals.css`:
```css
@tailwind base;
@tailwind components;
@tailwind utilities;

body {
  background: #e0e5ec;
  color: #2d3748;
}
```

`frontend/src/app/layout.tsx`:
```tsx
import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "Длинные нарды",
  description: "Онлайн-игра в длинные нарды"
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="ru">
      <body>{children}</body>
    </html>
  );
}
```

`frontend/src/app/page.tsx`:
```tsx
export default function HomePage() {
  return (
    <main className="min-h-screen flex items-center justify-center">
      <h1 className="text-4xl font-bold text-neo-accent">Длинные нарды</h1>
    </main>
  );
}
```

- [ ] **Step 4: Установить зависимости и проверить сборку**

Run: `cd frontend && npm install`
Expected: установка завершается без ошибок.

Run: `cd frontend && npm run typecheck && npm run build`
Expected: TypeScript проходит, билд успешен, появилась папка `.next/`.

- [ ] **Step 5: Commit**

```bash
git add frontend/
git commit -m "feat(frontend): initialize Next.js 14 with TypeScript and Tailwind"
```

---

### Task 0.4: Docker Compose с PostgreSQL

**Files:**
- Create: `docker-compose.yml`

- [ ] **Step 1: Создать `docker-compose.yml`**

```yaml
services:
  postgres:
    image: postgres:16-alpine
    container_name: backgammon-postgres
    environment:
      POSTGRES_DB: ${POSTGRES_DB:-backgammon}
      POSTGRES_USER: ${POSTGRES_USER:-bg_user}
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD:-bg_pass}
    volumes:
      - postgres_data:/var/lib/postgresql/data
    ports:
      - "5432:5432"
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U ${POSTGRES_USER:-bg_user}"]
      interval: 5s
      timeout: 5s
      retries: 5

volumes:
  postgres_data:
```

> Сервисы `backend` и `frontend` будут добавлены в Plan 2 и Plan 5.

- [ ] **Step 2: Валидировать compose-файл**

Run: `docker compose config`
Expected: вывод нормализованной конфигурации без ошибок.

- [ ] **Step 3: Запустить и проверить PostgreSQL**

Run: `cp .env.example .env && docker compose up -d postgres`
Expected: контейнер `backgammon-postgres` запущен.

Run: `docker compose ps`
Expected: статус `healthy` (после 10-15 секунд).

Run: `docker exec backgammon-postgres psql -U bg_user -d backgammon -c "SELECT version();"`
Expected: вывод версии PostgreSQL 16.

- [ ] **Step 4: Остановить контейнер**

```bash
docker compose down
```

- [ ] **Step 5: Commit**

```bash
git add docker-compose.yml
git commit -m "feat: add docker-compose with PostgreSQL service"
```

---

### Task 0.5: Тег Phase 0

- [ ] **Step 1: Финальная проверка**

Run все три проверки последовательно:
```bash
cd backend && go build ./... && go test ./...
cd ../frontend && npm run typecheck && npm run build
cd .. && docker compose config > /dev/null && echo "compose ok"
```
Expected: все три команды успешны.

- [ ] **Step 2: Тег**

```bash
git tag phase-0-scaffolding
git log --oneline | head
```

---

## Phase 1: Game Logic — длинные нарды (Tasks 1.1 — 1.18)

### Контракты пакета `game`

Перед тем как писать тесты, зафиксируем API публичного пакета (то, что будет использовано в Plan 2/3):

```go
// Color: цвет игрока
type Color int
const (
    NoColor Color = iota
    White
    Black
)

// Phase: фаза игры
type Phase int
const (
    PhaseWaiting Phase = iota
    PhaseRollingFirst
    PhasePlaying
    PhaseBearingOff
    PhaseFinished
)

// Point: пункт на доске
type Point struct {
    Owner    Color
    Checkers int
}

// Board: вся доска
type Board struct {
    Points   [25]Point  // индексы 1..24, 0 не используется
    BorneOff [3]int     // BorneOff[White], BorneOff[Black]
}

// Move: одно перемещение
type Move struct {
    From int  // 1..24
    To   int  // 0 (bear-off для белых), 25 (bear-off для чёрных), 1..24 иначе
    Die  int  // 1..6
}

// Dice: интерфейс источника кубиков (для детерминированных тестов)
type Dice interface {
    Roll() (int, int)
}

// Game: публичная фасада состояния игры
type Game struct {
    Board         *Board
    CurrentTurn   Color
    Dice          []int
    RemainingDice []int
    Phase         Phase
    Winner        Color
    IsMars        bool
    MoveCount     int
}

// Публичные функции
func NewBoard() *Board
func NewGame() *Game
func (g *Game) Roll(d Dice) error
func (g *Game) ApplyMove(m Move) error
func (g *Game) EndTurn() error
func (g *Game) AvailableMoves() [][]Move  // список возможных последовательностей
func (b *Board) AllInHome(c Color) bool
```

---

### Task 1.1: Цвета и направление

**Files:**
- Create: `backend/internal/game/color.go`
- Create: `backend/internal/game/color_test.go`

- [ ] **Step 1: Написать падающий тест**

`backend/internal/game/color_test.go`:
```go
package game

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestColor_Opponent(t *testing.T) {
	assert.Equal(t, Black, White.Opponent())
	assert.Equal(t, White, Black.Opponent())
}

func TestColor_Direction(t *testing.T) {
	assert.Equal(t, -1, White.Direction(), "white moves from 24 toward 1")
	assert.Equal(t, +1, Black.Direction(), "black moves from 1 toward 24")
}

func TestColor_HomeRange(t *testing.T) {
	lo, hi := White.HomeRange()
	assert.Equal(t, 1, lo)
	assert.Equal(t, 6, hi)
	lo, hi = Black.HomeRange()
	assert.Equal(t, 19, lo)
	assert.Equal(t, 24, hi)
}

func TestColor_StartPoint(t *testing.T) {
	assert.Equal(t, 24, White.StartPoint())
	assert.Equal(t, 1, Black.StartPoint())
}

func TestColor_BearOffTarget(t *testing.T) {
	assert.Equal(t, 0, White.BearOffTarget())
	assert.Equal(t, 25, Black.BearOffTarget())
}
```

- [ ] **Step 2: Запустить тест — должен упасть с ошибкой компиляции**

Run: `cd backend && go test ./internal/game/ -run TestColor`
Expected: FAIL — `undefined: Color`, `undefined: White`, etc.

- [ ] **Step 3: Реализовать `color.go`**

`backend/internal/game/color.go`:
```go
package game

type Color int

const (
	NoColor Color = iota
	White
	Black
)

func (c Color) Opponent() Color {
	switch c {
	case White:
		return Black
	case Black:
		return White
	}
	return NoColor
}

func (c Color) Direction() int {
	switch c {
	case White:
		return -1
	case Black:
		return +1
	}
	return 0
}

func (c Color) HomeRange() (lo, hi int) {
	switch c {
	case White:
		return 1, 6
	case Black:
		return 19, 24
	}
	return 0, 0
}

func (c Color) StartPoint() int {
	switch c {
	case White:
		return 24
	case Black:
		return 1
	}
	return 0
}

func (c Color) BearOffTarget() int {
	switch c {
	case White:
		return 0
	case Black:
		return 25
	}
	return -1
}
```

- [ ] **Step 4: Запустить тест — должен пройти**

Run: `cd backend && go test ./internal/game/ -run TestColor -v`
Expected: PASS все 5 тестов.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/game/color.go backend/internal/game/color_test.go
git commit -m "feat(game): add Color type with direction, home range, and bear-off target"
```

---

### Task 1.2: Доска и начальная расстановка

**Files:**
- Create: `backend/internal/game/board.go`
- Create: `backend/internal/game/board_test.go`

- [ ] **Step 1: Написать падающий тест**

`backend/internal/game/board_test.go`:
```go
package game

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewBoard_InitialSetup(t *testing.T) {
	b := NewBoard()

	assert.Equal(t, 15, b.Points[24].Checkers, "white starts with 15 checkers on point 24")
	assert.Equal(t, White, b.Points[24].Owner)

	assert.Equal(t, 15, b.Points[1].Checkers, "black starts with 15 checkers on point 1")
	assert.Equal(t, Black, b.Points[1].Owner)

	for i := 2; i <= 23; i++ {
		assert.Equal(t, 0, b.Points[i].Checkers, "point %d must be empty", i)
		assert.Equal(t, NoColor, b.Points[i].Owner, "point %d must have no owner", i)
	}

	assert.Equal(t, 0, b.BorneOff[White])
	assert.Equal(t, 0, b.BorneOff[Black])
}

func TestBoard_CountCheckers(t *testing.T) {
	b := NewBoard()
	assert.Equal(t, 15, b.CountCheckers(White))
	assert.Equal(t, 15, b.CountCheckers(Black))
}
```

- [ ] **Step 2: Запустить тест — должен упасть**

Run: `cd backend && go test ./internal/game/ -run TestNewBoard -v`
Expected: FAIL — `undefined: NewBoard`, `undefined: Board`.

- [ ] **Step 3: Реализовать `board.go`**

`backend/internal/game/board.go`:
```go
package game

type Point struct {
	Owner    Color
	Checkers int
}

type Board struct {
	Points   [25]Point
	BorneOff [3]int
}

func NewBoard() *Board {
	b := &Board{}
	b.Points[White.StartPoint()] = Point{Owner: White, Checkers: 15}
	b.Points[Black.StartPoint()] = Point{Owner: Black, Checkers: 15}
	return b
}

func (b *Board) CountCheckers(c Color) int {
	total := b.BorneOff[c]
	for i := 1; i <= 24; i++ {
		if b.Points[i].Owner == c {
			total += b.Points[i].Checkers
		}
	}
	return total
}
```

- [ ] **Step 4: Запустить тесты**

Run: `cd backend && go test ./internal/game/ -v`
Expected: все тесты PASS.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/game/board.go backend/internal/game/board_test.go
git commit -m "feat(game): add Board with initial long-backgammon setup"
```

---

### Task 1.3: Move и Apply

**Files:**
- Create: `backend/internal/game/move.go`
- Create: `backend/internal/game/move_test.go`

- [ ] **Step 1: Написать падающий тест**

`backend/internal/game/move_test.go`:
```go
package game

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMove_Apply_WhiteRegular(t *testing.T) {
	b := NewBoard()
	m := Move{From: 24, To: 18, Die: 6}

	err := m.Apply(b, White)

	assert.NoError(t, err)
	assert.Equal(t, 14, b.Points[24].Checkers)
	assert.Equal(t, White, b.Points[24].Owner)
	assert.Equal(t, 1, b.Points[18].Checkers)
	assert.Equal(t, White, b.Points[18].Owner)
}

func TestMove_Apply_BlackRegular(t *testing.T) {
	b := NewBoard()
	m := Move{From: 1, To: 5, Die: 4}

	err := m.Apply(b, Black)

	assert.NoError(t, err)
	assert.Equal(t, 14, b.Points[1].Checkers)
	assert.Equal(t, 1, b.Points[5].Checkers)
	assert.Equal(t, Black, b.Points[5].Owner)
}

func TestMove_Apply_EmptiesSourcePointOwner(t *testing.T) {
	b := &Board{}
	b.Points[24] = Point{Owner: White, Checkers: 1}
	m := Move{From: 24, To: 23, Die: 1}

	err := m.Apply(b, White)

	assert.NoError(t, err)
	assert.Equal(t, 0, b.Points[24].Checkers)
	assert.Equal(t, NoColor, b.Points[24].Owner, "empty point must reset its owner")
}

func TestMove_Apply_BearOffWhite(t *testing.T) {
	b := &Board{}
	b.Points[5] = Point{Owner: White, Checkers: 1}
	m := Move{From: 5, To: 0, Die: 5}

	err := m.Apply(b, White)

	assert.NoError(t, err)
	assert.Equal(t, 0, b.Points[5].Checkers)
	assert.Equal(t, 1, b.BorneOff[White])
}

func TestMove_Apply_BearOffBlack(t *testing.T) {
	b := &Board{}
	b.Points[22] = Point{Owner: Black, Checkers: 1}
	m := Move{From: 22, To: 25, Die: 3}

	err := m.Apply(b, Black)

	assert.NoError(t, err)
	assert.Equal(t, 0, b.Points[22].Checkers)
	assert.Equal(t, 1, b.BorneOff[Black])
}

func TestMove_Apply_NoCheckerAtSource(t *testing.T) {
	b := NewBoard()
	m := Move{From: 10, To: 4, Die: 6}

	err := m.Apply(b, White)

	assert.Error(t, err)
}
```

- [ ] **Step 2: Запустить тест — должен упасть**

Run: `cd backend && go test ./internal/game/ -run TestMove_Apply -v`
Expected: FAIL — `undefined: Move`.

- [ ] **Step 3: Реализовать `move.go`**

`backend/internal/game/move.go`:
```go
package game

import "fmt"

type Move struct {
	From int
	To   int
	Die  int
}

func (m Move) Apply(b *Board, c Color) error {
	src := &b.Points[m.From]
	if src.Owner != c || src.Checkers == 0 {
		return fmt.Errorf("no %v checker at point %d", c, m.From)
	}

	src.Checkers--
	if src.Checkers == 0 {
		src.Owner = NoColor
	}

	if m.To == c.BearOffTarget() {
		b.BorneOff[c]++
		return nil
	}

	dst := &b.Points[m.To]
	dst.Owner = c
	dst.Checkers++
	return nil
}
```

- [ ] **Step 4: Запустить тесты**

Run: `cd backend && go test ./internal/game/ -v`
Expected: все тесты PASS.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/game/move.go backend/internal/game/move_test.go
git commit -m "feat(game): add Move with Apply for regular and bear-off moves"
```

---

### Task 1.4: Кубики (обычный бросок + детерминированный источник)

**Files:**
- Create: `backend/internal/game/dice.go`
- Create: `backend/internal/game/dice_test.go`

- [ ] **Step 1: Написать падающий тест**

`backend/internal/game/dice_test.go`:
```go
package game

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExpandDice_Regular(t *testing.T) {
	result := ExpandDice(3, 5)
	assert.Equal(t, []int{3, 5}, result)
}

func TestExpandDice_Double(t *testing.T) {
	result := ExpandDice(4, 4)
	assert.Equal(t, []int{4, 4, 4, 4}, result, "double gives 4 uses")
}

func TestFixedDice_Roll(t *testing.T) {
	d := NewFixedDice([][2]int{{3, 5}, {6, 6}})

	a, b := d.Roll()
	assert.Equal(t, 3, a)
	assert.Equal(t, 5, b)

	a, b = d.Roll()
	assert.Equal(t, 6, a)
	assert.Equal(t, 6, b)
}

func TestRandomDice_RollInRange(t *testing.T) {
	d := NewRandomDice(42)
	for i := 0; i < 100; i++ {
		a, b := d.Roll()
		assert.GreaterOrEqual(t, a, 1)
		assert.LessOrEqual(t, a, 6)
		assert.GreaterOrEqual(t, b, 1)
		assert.LessOrEqual(t, b, 6)
	}
}
```

- [ ] **Step 2: Запустить тест — должен упасть**

Run: `cd backend && go test ./internal/game/ -run "TestExpandDice|TestFixedDice|TestRandomDice" -v`
Expected: FAIL — `undefined: ExpandDice`, `undefined: NewFixedDice`.

- [ ] **Step 3: Реализовать `dice.go`**

`backend/internal/game/dice.go`:
```go
package game

import "math/rand"

type Dice interface {
	Roll() (int, int)
}

type RandomDice struct {
	rng *rand.Rand
}

func NewRandomDice(seed int64) *RandomDice {
	return &RandomDice{rng: rand.New(rand.NewSource(seed))}
}

func (d *RandomDice) Roll() (int, int) {
	return d.rng.Intn(6) + 1, d.rng.Intn(6) + 1
}

type FixedDice struct {
	rolls []([2]int)
	idx   int
}

func NewFixedDice(rolls [][2]int) *FixedDice {
	return &FixedDice{rolls: rolls}
}

func (d *FixedDice) Roll() (int, int) {
	r := d.rolls[d.idx]
	d.idx++
	return r[0], r[1]
}

func ExpandDice(a, b int) []int {
	if a == b {
		return []int{a, a, a, a}
	}
	return []int{a, b}
}
```

- [ ] **Step 4: Запустить тесты**

Run: `cd backend && go test ./internal/game/ -v`
Expected: все тесты PASS.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/game/dice.go backend/internal/game/dice_test.go
git commit -m "feat(game): add Dice interface with Random and Fixed implementations"
```

---

### Task 1.5: Базовая валидация хода (направление, границы, заполненность)

**Files:**
- Create: `backend/internal/game/rules.go`
- Create: `backend/internal/game/rules_test.go`

- [ ] **Step 1: Написать падающий тест**

`backend/internal/game/rules_test.go`:
```go
package game

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsValidMove_WhiteCorrectDirection(t *testing.T) {
	b := NewBoard()
	ok, _ := IsValidMove(b, White, Move{From: 24, To: 18, Die: 6})
	assert.True(t, ok)
}

func TestIsValidMove_WhiteWrongDirection(t *testing.T) {
	b := &Board{}
	b.Points[10] = Point{Owner: White, Checkers: 1}
	ok, err := IsValidMove(b, White, Move{From: 10, To: 16, Die: 6})
	assert.False(t, ok)
	assert.NotNil(t, err)
}

func TestIsValidMove_BlackCorrectDirection(t *testing.T) {
	b := NewBoard()
	ok, _ := IsValidMove(b, Black, Move{From: 1, To: 7, Die: 6})
	assert.True(t, ok)
}

func TestIsValidMove_NoCheckerAtSource(t *testing.T) {
	b := NewBoard()
	ok, _ := IsValidMove(b, White, Move{From: 13, To: 7, Die: 6})
	assert.False(t, ok)
}

func TestIsValidMove_OpponentAtSource(t *testing.T) {
	b := NewBoard()
	ok, _ := IsValidMove(b, White, Move{From: 1, To: 0, Die: 1}) // 1 — чёрная стартовая
	assert.False(t, ok)
}

func TestIsValidMove_DestinationOccupiedByOpponent(t *testing.T) {
	b := &Board{}
	b.Points[24] = Point{Owner: White, Checkers: 15}
	b.Points[18] = Point{Owner: Black, Checkers: 1}
	ok, _ := IsValidMove(b, White, Move{From: 24, To: 18, Die: 6})
	assert.False(t, ok, "cannot land on opponent's point (no hitting in long backgammon)")
}

func TestIsValidMove_DestinationOwnColorAllowed(t *testing.T) {
	b := &Board{}
	b.Points[24] = Point{Owner: White, Checkers: 14}
	b.Points[18] = Point{Owner: White, Checkers: 1}
	ok, _ := IsValidMove(b, White, Move{From: 24, To: 18, Die: 6})
	assert.True(t, ok)
}

func TestIsValidMove_DieMismatch(t *testing.T) {
	b := NewBoard()
	ok, _ := IsValidMove(b, White, Move{From: 24, To: 18, Die: 5}) // 24-5=19, не 18
	assert.False(t, ok)
}

func TestIsValidMove_OutOfBoundsForRegularMove(t *testing.T) {
	b := &Board{}
	b.Points[3] = Point{Owner: White, Checkers: 1}
	// для регулярного хода (не bear-off) to должен быть 1..24
	ok, _ := IsValidMove(b, White, Move{From: 3, To: -2, Die: 5})
	assert.False(t, ok)
}
```

- [ ] **Step 2: Запустить тест — должен упасть**

Run: `cd backend && go test ./internal/game/ -run TestIsValidMove -v`
Expected: FAIL — `undefined: IsValidMove`.

- [ ] **Step 3: Реализовать `rules.go`**

`backend/internal/game/rules.go`:
```go
package game

import "fmt"

func IsValidMove(b *Board, c Color, m Move) (bool, error) {
	if m.Die < 1 || m.Die > 6 {
		return false, fmt.Errorf("die out of range: %d", m.Die)
	}
	if m.From < 1 || m.From > 24 {
		return false, fmt.Errorf("from out of range: %d", m.From)
	}

	src := b.Points[m.From]
	if src.Owner != c || src.Checkers == 0 {
		return false, fmt.Errorf("no %v checker at %d", c, m.From)
	}

	expectedTo := m.From + c.Direction()*m.Die
	if expectedTo != m.To {
		return false, fmt.Errorf("die %d from %d for %v leads to %d, not %d",
			m.Die, m.From, c, expectedTo, m.To)
	}

	if m.To == c.BearOffTarget() {
		return isValidBearOff(b, c, m)
	}

	if m.To < 1 || m.To > 24 {
		return false, fmt.Errorf("to out of range for regular move: %d", m.To)
	}

	dst := b.Points[m.To]
	if dst.Owner != NoColor && dst.Owner != c {
		return false, fmt.Errorf("point %d is occupied by opponent", m.To)
	}

	return true, nil
}

// isValidBearOff — заглушка, полная реализация в Task 1.10.
func isValidBearOff(b *Board, c Color, m Move) (bool, error) {
	return false, fmt.Errorf("bear-off not yet allowed in this task")
}
```

- [ ] **Step 4: Запустить тесты**

Run: `cd backend && go test ./internal/game/ -v`
Expected: все тесты PASS (тесты на bear-off ещё не добавлены).

- [ ] **Step 5: Commit**

```bash
git add backend/internal/game/rules.go backend/internal/game/rules_test.go
git commit -m "feat(game): add IsValidMove with direction, ownership, and occupation checks"
```

---

### Task 1.6: "Глухой забор" (запрет 6+ блоков, отрезающих все шашки соперника)

**Files:**
- Modify: `backend/internal/game/rules.go`
- Modify: `backend/internal/game/rules_test.go`

> **Правило:** в длинных нардах нельзя строить непрерывный ряд из 6 и более занятых пунктов, если хотя бы одна шашка соперника **впереди** этого ряда (по направлению её движения). Это правило проверяется *после* предполагаемого хода — будет ли в результате построен запрещённый забор.
>
> "Впереди" определяется направлением движения соперника:
> - Для белых движение от высоких к низким, "впереди ряда" = пункт с *более высоким* номером.
> - Для чёрных движение от низких к высоким, "впереди ряда" = пункт с *более низким* номером.

- [ ] **Step 1: Дописать падающие тесты**

Добавить в `backend/internal/game/rules_test.go`:
```go
func TestIsValidMove_GlukhoiZabor_ClosingSixthBlocked(t *testing.T) {
	// Белые держат подряд 17..21 (5 пунктов). Чёрная одиночка на 23 — впереди ряда.
	// Ход White from=24 die=2 → to=22 замкнул бы 17..22 (6 подряд), отрезая чёрную → запрещён.
	b := &Board{}
	b.Points[17] = Point{Owner: White, Checkers: 2}
	b.Points[18] = Point{Owner: White, Checkers: 2}
	b.Points[19] = Point{Owner: White, Checkers: 2}
	b.Points[20] = Point{Owner: White, Checkers: 2}
	b.Points[21] = Point{Owner: White, Checkers: 2}
	b.Points[24] = Point{Owner: White, Checkers: 1}
	b.Points[23] = Point{Owner: Black, Checkers: 1}

	ok, _ := IsValidMove(b, White, Move{From: 24, To: 22, Die: 2})
	assert.False(t, ok, "closing 6th consecutive point ahead of opponent must be blocked")
}

func TestIsValidMove_GlukhoiZabor_AllowedNoOpponentAhead(t *testing.T) {
	// Те же 5 подряд 17..21, но чёрная позади (на 15). Закрытие 22 разрешено.
	b := &Board{}
	b.Points[17] = Point{Owner: White, Checkers: 2}
	b.Points[18] = Point{Owner: White, Checkers: 2}
	b.Points[19] = Point{Owner: White, Checkers: 2}
	b.Points[20] = Point{Owner: White, Checkers: 2}
	b.Points[21] = Point{Owner: White, Checkers: 2}
	b.Points[24] = Point{Owner: White, Checkers: 1}
	b.Points[15] = Point{Owner: Black, Checkers: 1}

	ok, _ := IsValidMove(b, White, Move{From: 24, To: 22, Die: 2})
	assert.True(t, ok, "6 in a row is allowed if no opponent checker is ahead of the wall")
}
```

- [ ] **Step 2: Запустить — упадёт**

Run: `cd backend && go test ./internal/game/ -run GlukhoiZabor -v`
Expected: FAIL — забор не проверяется.

- [ ] **Step 3: Реализовать проверку забора в `rules.go`**

Добавить в `rules.go` перед `return true, nil`:
```go
	if wouldCreateGlukhoiZabor(b, c, m) {
		return false, fmt.Errorf("move would create a 6+ block ahead of opponent (glukhoi zabor)")
	}
```

Добавить функцию:
```go
// wouldCreateGlukhoiZabor моделирует ход и проверяет, образуется ли непрерывный
// ряд из 6+ пунктов цвета c с шашкой соперника впереди этого ряда.
func wouldCreateGlukhoiZabor(b *Board, c Color, m Move) bool {
	sim := *b
	sim.Points[m.From] = Point{Owner: c, Checkers: b.Points[m.From].Checkers - 1}
	if sim.Points[m.From].Checkers == 0 {
		sim.Points[m.From].Owner = NoColor
	}
	if m.To >= 1 && m.To <= 24 {
		sim.Points[m.To] = Point{Owner: c, Checkers: b.Points[m.To].Checkers + 1}
	}

	opponentAhead := func(start, end int) bool {
		// Соперник "впереди ряда" — там, куда он движется.
		// Белые ряд блокирует чёрных, которые идут к высоким номерам → впереди = пункты > end.
		// Чёрные ряд блокирует белых, которые идут к низким → впереди = пункты < start.
		opp := c.Opponent()
		if c == White {
			for p := end + 1; p <= 24; p++ {
				if sim.Points[p].Owner == opp {
					return true
				}
			}
		} else {
			for p := start - 1; p >= 1; p-- {
				if sim.Points[p].Owner == opp {
					return true
				}
			}
		}
		return false
	}

	run := 0
	runStart := 0
	for p := 1; p <= 24; p++ {
		if sim.Points[p].Owner == c {
			if run == 0 {
				runStart = p
			}
			run++
			if run >= 6 && opponentAhead(runStart, p) {
				return true
			}
		} else {
			run = 0
			runStart = 0
		}
	}
	return false
}
```

- [ ] **Step 4: Запустить все тесты**

Run: `cd backend && go test ./internal/game/ -v`
Expected: все PASS, включая оба теста на забор.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/game/rules.go backend/internal/game/rules_test.go
git commit -m "feat(game): enforce glukhoi zabor rule (no 6+ block ahead of opponent)"
```

---

### Task 1.7: Генерация одиночных ходов

**Files:**
- Create: `backend/internal/game/moves.go`
- Create: `backend/internal/game/moves_test.go`

- [ ] **Step 1: Падающий тест**

`backend/internal/game/moves_test.go`:
```go
package game

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateSingleMoves_InitialBoardWhite(t *testing.T) {
	b := NewBoard()
	// Только одна стартовая стопка у белых на 24. Возможный ход на любой свободный пункт впереди.
	moves := GenerateSingleMoves(b, White, 3)
	assert.Contains(t, moves, Move{From: 24, To: 21, Die: 3})
}

func TestGenerateSingleMoves_NoMoveForOccupiedByOpponent(t *testing.T) {
	b := &Board{}
	b.Points[10] = Point{Owner: White, Checkers: 1}
	b.Points[4] = Point{Owner: Black, Checkers: 1}
	moves := GenerateSingleMoves(b, White, 6)
	assert.Empty(t, moves, "destination 4 is occupied by black")
}

func TestGenerateSingleMoves_MultipleSources(t *testing.T) {
	b := &Board{}
	b.Points[24] = Point{Owner: White, Checkers: 2}
	b.Points[20] = Point{Owner: White, Checkers: 1}
	moves := GenerateSingleMoves(b, White, 4)
	assert.Contains(t, moves, Move{From: 24, To: 20, Die: 4})
	assert.Contains(t, moves, Move{From: 20, To: 16, Die: 4})
}
```

- [ ] **Step 2: Run — fail**

Run: `cd backend && go test ./internal/game/ -run TestGenerateSingleMoves -v`
Expected: FAIL — `undefined: GenerateSingleMoves`.

- [ ] **Step 3: Реализовать `moves.go`**

`backend/internal/game/moves.go`:
```go
package game

// GenerateSingleMoves возвращает все валидные ходы цветом c для кубика die.
func GenerateSingleMoves(b *Board, c Color, die int) []Move {
	var out []Move
	for from := 1; from <= 24; from++ {
		if b.Points[from].Owner != c || b.Points[from].Checkers == 0 {
			continue
		}
		to := from + c.Direction()*die
		m := Move{From: from, To: to, Die: die}
		if ok, _ := IsValidMove(b, c, m); ok {
			out = append(out, m)
		}
	}
	return out
}
```

- [ ] **Step 4: Запустить тесты**

Run: `cd backend && go test ./internal/game/ -v`
Expected: все PASS.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/game/moves.go backend/internal/game/moves_test.go
git commit -m "feat(game): add GenerateSingleMoves for a single die value"
```

---

### Task 1.8: Последовательности ходов для броска (обе кости + дубль)

**Files:**
- Modify: `backend/internal/game/moves.go`
- Modify: `backend/internal/game/moves_test.go`

- [ ] **Step 1: Дописать падающие тесты**

В `moves_test.go`:
```go
func TestGenerateSequences_TwoDistinctDice(t *testing.T) {
	b := &Board{}
	b.Points[24] = Point{Owner: White, Checkers: 2}
	dice := []int{3, 5}

	sequences := GenerateSequences(b, White, dice)

	// Должно быть хотя бы две последовательности: 3 потом 5, и 5 потом 3.
	assert.NotEmpty(t, sequences)
	hasOrder35 := false
	hasOrder53 := false
	for _, seq := range sequences {
		if len(seq) == 2 && seq[0].Die == 3 && seq[1].Die == 5 {
			hasOrder35 = true
		}
		if len(seq) == 2 && seq[0].Die == 5 && seq[1].Die == 3 {
			hasOrder53 = true
		}
	}
	assert.True(t, hasOrder35, "expect a sequence using die 3 then 5")
	assert.True(t, hasOrder53, "expect a sequence using die 5 then 3")
}

func TestGenerateSequences_Double(t *testing.T) {
	b := &Board{}
	b.Points[24] = Point{Owner: White, Checkers: 4}
	dice := ExpandDice(2, 2) // 4 двойки

	sequences := GenerateSequences(b, White, dice)

	// Должна существовать последовательность из 4 ходов.
	found4 := false
	for _, seq := range sequences {
		if len(seq) == 4 {
			found4 = true
		}
	}
	assert.True(t, found4, "double 2 must allow a 4-move sequence")
}

func TestGenerateSequences_NoMovesAvailable(t *testing.T) {
	b := &Board{}
	b.Points[10] = Point{Owner: White, Checkers: 1}
	// Перекрыть оба возможных хода
	b.Points[10-3] = Point{Owner: Black, Checkers: 1}
	b.Points[10-5] = Point{Owner: Black, Checkers: 1}

	sequences := GenerateSequences(b, White, []int{3, 5})

	// Допустимо: пустой список (нет ходов) или последовательности нулевой длины.
	for _, seq := range sequences {
		assert.Empty(t, seq, "no moves should be possible")
	}
}
```

- [ ] **Step 2: Run — fail**

Run: `cd backend && go test ./internal/game/ -run TestGenerateSequences -v`
Expected: FAIL — `undefined: GenerateSequences`.

- [ ] **Step 3: Дополнить `moves.go`**

Добавить в `moves.go`:
```go
// GenerateSequences возвращает все максимально длинные последовательности
// ходов, которые игрок может сделать с данным набором кубиков.
// Сначала перебираются все перестановки порядка применения кубиков, затем
// фильтруется список так, что остаются только последовательности максимальной
// длины (правило "обязан использовать максимум возможных кубиков").
func GenerateSequences(b *Board, c Color, dice []int) [][]Move {
	var all [][]Move
	collect(b, c, dice, []Move{}, &all)

	if len(all) == 0 {
		return nil
	}

	maxLen := 0
	for _, seq := range all {
		if len(seq) > maxLen {
			maxLen = len(seq)
		}
	}

	var filtered [][]Move
	for _, seq := range all {
		if len(seq) == maxLen {
			filtered = append(filtered, seq)
		}
	}
	return filtered
}

func collect(b *Board, c Color, dice []int, prefix []Move, out *[][]Move) {
	if len(dice) == 0 {
		cp := append([]Move(nil), prefix...)
		*out = append(*out, cp)
		return
	}

	anyMove := false
	for i, die := range dice {
		used := make(map[int]bool)
		if used[die] {
			continue
		}
		used[die] = true

		moves := GenerateSingleMoves(b, c, die)
		for _, m := range moves {
			anyMove = true
			sim := *b
			_ = m.Apply(&sim, c)
			rest := append([]int{}, dice[:i]...)
			rest = append(rest, dice[i+1:]...)
			collect(&sim, c, rest, append(prefix, m), out)
		}
	}

	if !anyMove {
		cp := append([]Move(nil), prefix...)
		*out = append(*out, cp)
	}
}
```

- [ ] **Step 4: Run all**

Run: `cd backend && go test ./internal/game/ -v`
Expected: все PASS.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/game/moves.go backend/internal/game/moves_test.go
git commit -m "feat(game): generate maximal move sequences for a dice roll"
```

---

### Task 1.9: Обязательное использование большего кубика

**Files:**
- Modify: `backend/internal/game/moves.go`
- Modify: `backend/internal/game/moves_test.go`

> **Правило:** если игрок может использовать только один кубик из двух, он обязан использовать *больший*.
>
> Это правило естественно следует из фильтрации по максимальной длине — но *с дополнением*: если все максимальные последовательности имеют длину 1, оставляем только те, что используют больший кубик.

- [ ] **Step 1: Падающий тест**

В `moves_test.go`:
```go
func TestGenerateSequences_MustUseLargerDie(t *testing.T) {
	// Белая одиночка на пункте 8 + 14 шашек на 24 (не все в доме → bear-off запрещён).
	// Кубики {3, 5}:
	//   8→3 (die=5): после хода с 3 ходов нет (нет шашки на 3 для следующей кости).
	//   8→5 (die=3): после хода с 5 ходов нет (нет шашки на 5).
	// Максимальная длина = 1. По правилу обязательного большего остаётся только ход с die=5.
	b := &Board{}
	b.Points[8] = Point{Owner: White, Checkers: 1}
	b.Points[24] = Point{Owner: White, Checkers: 14}

	sequences := GenerateSequences(b, White, []int{3, 5})

	assert.NotEmpty(t, sequences)
	for _, seq := range sequences {
		assert.Equal(t, 1, len(seq))
		assert.Equal(t, 5, seq[0].Die, "must use the larger die when only one is usable")
	}
}
```

- [ ] **Step 2: Run — fail**

Run: `cd backend && go test ./internal/game/ -run TestGenerateSequences_MustUseLargerDie -v`
Expected: FAIL — текущая фильтрация по длине не выбирает больший кубик.

- [ ] **Step 3: Дополнить `GenerateSequences`**

В `moves.go`, заменить блок фильтрации:
```go
	maxLen := 0
	for _, seq := range all {
		if len(seq) > maxLen {
			maxLen = len(seq)
		}
	}

	var filtered [][]Move
	for _, seq := range all {
		if len(seq) == maxLen {
			filtered = append(filtered, seq)
		}
	}
```

на:
```go
	maxLen := 0
	for _, seq := range all {
		if len(seq) > maxLen {
			maxLen = len(seq)
		}
	}

	var filtered [][]Move
	for _, seq := range all {
		if len(seq) == maxLen {
			filtered = append(filtered, seq)
		}
	}

	// Если максимальная длина = 1 и в исходных кубиках было два разных значения,
	// то по правилу обязательного большего оставляем только использующие больший.
	if maxLen == 1 && len(dice) == 2 && dice[0] != dice[1] {
		larger := dice[0]
		if dice[1] > larger {
			larger = dice[1]
		}
		var onlyLarger [][]Move
		for _, seq := range filtered {
			if seq[0].Die == larger {
				onlyLarger = append(onlyLarger, seq)
			}
		}
		if len(onlyLarger) > 0 {
			filtered = onlyLarger
		}
	}
	return filtered
```

(И убрать дублирующий `return filtered` в конце функции.)

- [ ] **Step 4: Run all**

Run: `cd backend && go test ./internal/game/ -v`
Expected: все PASS.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/game/moves.go backend/internal/game/moves_test.go
git commit -m "feat(game): enforce mandatory larger die when only one is usable"
```

---

### Task 1.10: Выкидывание (bear-off): база + правило старшего пункта

**Files:**
- Create: `backend/internal/game/bearoff.go`
- Create: `backend/internal/game/bearoff_test.go`
- Modify: `backend/internal/game/rules.go` (заменить заглушку `isValidBearOff`)

> **Правила выкидывания (длинные нарды):**
> 1. Стартовать выкидывание можно только когда все 15 шашек цвета `c` находятся в доме (1..6 для белых, 19..24 для чёрных) или уже выкинуты.
> 2. Если кубик `die` соответствует точно занятому пункту — шашка снимается.
> 3. Если на пункте `die` нет шашки, можно снять с *более старшего* пункта (для белых — больше die, для чёрных — меньше die-смещение).
> 4. Если и более старшего нет — игрок обязан сделать ход *внутри дома* (если есть валидный).

- [ ] **Step 1: Падающие тесты**

`backend/internal/game/bearoff_test.go`:
```go
package game

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAllInHome_White_True(t *testing.T) {
	b := &Board{}
	b.Points[6] = Point{Owner: White, Checkers: 15}
	assert.True(t, b.AllInHome(White))
}

func TestAllInHome_White_FalseOneOutside(t *testing.T) {
	b := &Board{}
	b.Points[6] = Point{Owner: White, Checkers: 14}
	b.Points[7] = Point{Owner: White, Checkers: 1}
	assert.False(t, b.AllInHome(White))
}

func TestAllInHome_White_TrueIncludingBorneOff(t *testing.T) {
	b := &Board{}
	b.Points[6] = Point{Owner: White, Checkers: 10}
	b.BorneOff[White] = 5
	assert.True(t, b.AllInHome(White))
}

func TestAllInHome_Black(t *testing.T) {
	b := &Board{}
	b.Points[24] = Point{Owner: Black, Checkers: 10}
	b.Points[19] = Point{Owner: Black, Checkers: 5}
	assert.True(t, b.AllInHome(Black))
}

func TestBearOff_ExactDie_White(t *testing.T) {
	b := &Board{}
	b.Points[5] = Point{Owner: White, Checkers: 1}
	b.Points[6] = Point{Owner: White, Checkers: 14}
	// die=5, with all in home, from=5, to=0
	ok, _ := IsValidMove(b, White, Move{From: 5, To: 0, Die: 5})
	assert.True(t, ok)
}

func TestBearOff_HigherPointFallback_White(t *testing.T) {
	// Все в доме, die=6, но на 6 нет шашек — снимаем с самой старшей (на 5).
	b := &Board{}
	b.Points[5] = Point{Owner: White, Checkers: 15}
	ok, _ := IsValidMove(b, White, Move{From: 5, To: 0, Die: 6})
	assert.True(t, ok, "if no checker at exact-die point, may bear off from a higher point — wait, for white higher = smaller? No: 5 < 6, so 5 is closer to bear-off. Rule: snimat' from samogo starshego when no exact match.")
}

func TestBearOff_NotAllInHome_Rejected(t *testing.T) {
	b := &Board{}
	b.Points[6] = Point{Owner: White, Checkers: 14}
	b.Points[7] = Point{Owner: White, Checkers: 1} // не в доме
	ok, _ := IsValidMove(b, White, Move{From: 6, To: 0, Die: 6})
	assert.False(t, ok)
}

func TestBearOff_HigherStillExists_FromExactRequired(t *testing.T) {
	// Все в доме. die=4, шашки на 4 и на 5.
	// from=5, to=0 (die=4 → 5→1) — это обычный ход внутри дома, не bear-off.
	// from=4, to=0, die=4 — обычный bear-off, ok.
	// Если игрок хочет from=5, to=0, die=4 — это НЕ bear-off (5-4=1, не 0).
	// Если хочет from=6, to=0, die=4 — но на 6 нет шашек.
	// Правило fallback: с пункта меньше die можно bear-off только если выше die нет шашек.
	b := &Board{}
	b.Points[4] = Point{Owner: White, Checkers: 1}
	b.Points[5] = Point{Owner: White, Checkers: 14}

	// Попытка bear-off с пункта 3 die=4 — на пункте 3 нет шашек.
	ok, _ := IsValidMove(b, White, Move{From: 3, To: 0, Die: 4})
	assert.False(t, ok, "no checker on point 3")

	// Попытка bear-off с пункта 5 die=4 — пункт 5 > die=4, но есть пункт 5, и нет точного 4? Нет, есть 4.
	// Тогда с 5 bear-off НЕ разрешён (так как точный 4 существует).
	ok, _ = IsValidMove(b, White, Move{From: 5, To: 0, Die: 4})
	assert.False(t, ok, "exact die=4 is occupied, cannot bear off from higher point")

	// Точный 4: ok.
	ok, _ = IsValidMove(b, White, Move{From: 4, To: 0, Die: 4})
	assert.True(t, ok)
}

func TestBearOff_Black_ExactDie(t *testing.T) {
	b := &Board{}
	b.Points[22] = Point{Owner: Black, Checkers: 1}
	b.Points[19] = Point{Owner: Black, Checkers: 14}
	// die=3, from=22, to=25 (22+3=25)
	ok, _ := IsValidMove(b, Black, Move{From: 22, To: 25, Die: 3})
	assert.True(t, ok)
}
```

- [ ] **Step 2: Run — fail**

Run: `cd backend && go test ./internal/game/ -run "TestAllInHome|TestBearOff" -v`
Expected: FAIL — `AllInHome` не реализована, bear-off возвращает ошибку.

- [ ] **Step 3: Реализовать `bearoff.go`**

`backend/internal/game/bearoff.go`:
```go
package game

func (b *Board) AllInHome(c Color) bool {
	lo, hi := c.HomeRange()
	for p := 1; p <= 24; p++ {
		if p >= lo && p <= hi {
			continue
		}
		if b.Points[p].Owner == c && b.Points[p].Checkers > 0 {
			return false
		}
	}
	return true
}

// pointForBearOffDie возвращает пункт, который соответствует кубику die
// для цвета c при выкидывании.
// Для белых: die=6 → пункт 6, die=1 → пункт 1.
// Для чёрных: die=6 → пункт 19, die=1 → пункт 24.
func pointForBearOffDie(c Color, die int) int {
	switch c {
	case White:
		return die
	case Black:
		return 25 - die
	}
	return 0
}

// highestOccupiedInHome возвращает самый "старший" (наиболее далёкий от bear-off)
// занятый пункт цвета c в его доме. Если нет — 0.
func highestOccupiedInHome(b *Board, c Color) int {
	lo, hi := c.HomeRange()
	switch c {
	case White:
		for p := hi; p >= lo; p-- {
			if b.Points[p].Owner == c && b.Points[p].Checkers > 0 {
				return p
			}
		}
	case Black:
		for p := lo; p <= hi; p++ {
			if b.Points[p].Owner == c && b.Points[p].Checkers > 0 {
				return p
			}
		}
	}
	return 0
}

// distanceToBearOff: для белых — это значение пункта (1..6), для чёрных — 25-p (1..6).
func distanceToBearOff(c Color, p int) int {
	switch c {
	case White:
		return p
	case Black:
		return 25 - p
	}
	return 0
}
```

- [ ] **Step 4: Реализовать `isValidBearOff` в `rules.go`**

В `rules.go` заменить заглушку:
```go
func isValidBearOff(b *Board, c Color, m Move) (bool, error) {
	if !b.AllInHome(c) {
		return false, fmt.Errorf("cannot bear off: not all checkers are in home")
	}

	srcDist := distanceToBearOff(c, m.From)

	// Точное соответствие кубику.
	if srcDist == m.Die {
		return true, nil
	}

	// Кубик больше расстояния — допустимо только если ВЫШЕ нет шашек.
	if m.Die > srcDist {
		highest := highestOccupiedInHome(b, c)
		highestDist := distanceToBearOff(c, highest)
		if highestDist == srcDist {
			// Это и есть самая старшая.
			return true, nil
		}
		return false, fmt.Errorf("bear-off from %d with die %d not allowed: higher checker on %d", m.From, m.Die, highest)
	}

	// Кубик меньше расстояния — это не bear-off, это обычный ход внутри дома (но to=BearOffTarget?).
	// Сюда мы не должны попадать при m.To == BearOffTarget — это означает, что die*direction != BearOffTarget - from.
	return false, fmt.Errorf("invalid bear-off attempt from %d with die %d", m.From, m.Die)
}
```

- [ ] **Step 5: Run all & commit**

Run: `cd backend && go test ./internal/game/ -v`
Expected: все PASS.

```bash
git add backend/internal/game/bearoff.go backend/internal/game/bearoff_test.go backend/internal/game/rules.go
git commit -m "feat(game): implement bear-off with exact-die and higher-point fallback rules"
```

---

### Task 1.11: Bear-off в `GenerateSingleMoves`

**Files:**
- Modify: `backend/internal/game/moves.go`
- Modify: `backend/internal/game/moves_test.go`

> Текущая `GenerateSingleMoves` пробует только `to = from + dir*die`. Для bear-off нужны два сценария:
> 1. Точный bear-off (`to = c.BearOffTarget()` если `srcDist == die`).
> 2. Fallback bear-off с самой старшей шашки (если `die > srcDist` и это самая старшая занятая).

- [ ] **Step 1: Падающий тест**

В `moves_test.go`:
```go
func TestGenerateSingleMoves_BearOff_Exact(t *testing.T) {
	b := &Board{}
	b.Points[6] = Point{Owner: White, Checkers: 15}
	moves := GenerateSingleMoves(b, White, 6)
	hasBearOff := false
	for _, m := range moves {
		if m.From == 6 && m.To == 0 {
			hasBearOff = true
		}
	}
	assert.True(t, hasBearOff)
}

func TestGenerateSingleMoves_BearOff_FallbackHigher(t *testing.T) {
	b := &Board{}
	b.Points[4] = Point{Owner: White, Checkers: 15} // все в доме, самая старшая на 4
	moves := GenerateSingleMoves(b, White, 6)
	hasBearOff := false
	for _, m := range moves {
		if m.From == 4 && m.To == 0 {
			hasBearOff = true
		}
	}
	assert.True(t, hasBearOff, "die=6 > top=4, fallback bear-off from 4 must be available")
}

func TestGenerateSingleMoves_BearOff_FallbackBlocked(t *testing.T) {
	b := &Board{}
	b.Points[4] = Point{Owner: White, Checkers: 1}
	b.Points[5] = Point{Owner: White, Checkers: 14}
	// die=6, самый старший на 5; bear-off с 4 НЕ разрешён.
	moves := GenerateSingleMoves(b, White, 6)
	for _, m := range moves {
		assert.False(t, m.From == 4 && m.To == 0, "must not bear off from 4 when 5 has checkers")
	}
}
```

- [ ] **Step 2: Run — fail**

Run: `cd backend && go test ./internal/game/ -run TestGenerateSingleMoves_BearOff -v`
Expected: FAIL — bear-off-ходы не генерируются.

- [ ] **Step 3: Дополнить `GenerateSingleMoves`**

Заменить тело функции в `moves.go`:
```go
func GenerateSingleMoves(b *Board, c Color, die int) []Move {
	var out []Move
	for from := 1; from <= 24; from++ {
		if b.Points[from].Owner != c || b.Points[from].Checkers == 0 {
			continue
		}
		// Обычный ход внутри доски.
		to := from + c.Direction()*die
		if to >= 1 && to <= 24 {
			m := Move{From: from, To: to, Die: die}
			if ok, _ := IsValidMove(b, c, m); ok {
				out = append(out, m)
			}
		}
	}

	// Bear-off ходы, если все в доме.
	if b.AllInHome(c) {
		target := c.BearOffTarget()
		// Точный bear-off.
		exact := pointForBearOffDie(c, die)
		if exact >= 1 && exact <= 24 && b.Points[exact].Owner == c && b.Points[exact].Checkers > 0 {
			out = append(out, Move{From: exact, To: target, Die: die})
		}
		// Fallback: самая старшая шашка ниже die.
		highest := highestOccupiedInHome(b, c)
		if highest > 0 && distanceToBearOff(c, highest) < die {
			out = append(out, Move{From: highest, To: target, Die: die})
		}
	}
	return out
}
```

- [ ] **Step 4: Run all**

Run: `cd backend && go test ./internal/game/ -v`
Expected: все PASS.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/game/moves.go backend/internal/game/moves_test.go
git commit -m "feat(game): include bear-off moves in single-move generator"
```

---

### Task 1.12: Game state machine + Roll

**Files:**
- Create: `backend/internal/game/game.go`
- Create: `backend/internal/game/game_test.go`

- [ ] **Step 1: Падающий тест**

`backend/internal/game/game_test.go`:
```go
package game

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewGame_InitialState(t *testing.T) {
	g := NewGame()
	assert.Equal(t, PhaseWaiting, g.Phase)
	assert.NotNil(t, g.Board)
	assert.Equal(t, 15, g.Board.Points[24].Checkers)
	assert.Equal(t, NoColor, g.Winner)
}

func TestGame_Roll_TransitionsFromRollingFirst(t *testing.T) {
	g := NewGame()
	g.Phase = PhaseRollingFirst
	g.CurrentTurn = White

	err := g.Roll(NewFixedDice([][2]int{{3, 5}}))

	assert.NoError(t, err)
	assert.Equal(t, []int{3, 5}, g.Dice)
	assert.Equal(t, []int{3, 5}, g.RemainingDice)
	assert.Equal(t, PhasePlaying, g.Phase)
}

func TestGame_Roll_DoubleExpansion(t *testing.T) {
	g := NewGame()
	g.Phase = PhasePlaying
	g.CurrentTurn = White

	err := g.Roll(NewFixedDice([][2]int{{4, 4}}))

	assert.NoError(t, err)
	assert.Equal(t, []int{4, 4}, g.Dice)
	assert.Equal(t, []int{4, 4, 4, 4}, g.RemainingDice)
}

func TestGame_Roll_RejectedInWrongPhase(t *testing.T) {
	g := NewGame()
	g.Phase = PhaseFinished

	err := g.Roll(NewFixedDice([][2]int{{1, 2}}))

	assert.Error(t, err)
}
```

- [ ] **Step 2: Run — fail**

Run: `cd backend && go test ./internal/game/ -run TestGame_Roll -v`
Expected: FAIL — `undefined: Game`, `undefined: NewGame`.

- [ ] **Step 3: Реализовать `game.go` (минимум для Roll)**

`backend/internal/game/game.go`:
```go
package game

import "fmt"

type Phase int

const (
	PhaseWaiting Phase = iota
	PhaseRollingFirst
	PhasePlaying
	PhaseBearingOff
	PhaseFinished
)

type Game struct {
	Board         *Board
	CurrentTurn   Color
	Dice          []int
	RemainingDice []int
	Phase         Phase
	Winner        Color
	IsMars        bool
	MoveCount     int
}

func NewGame() *Game {
	return &Game{
		Board:       NewBoard(),
		CurrentTurn: NoColor,
		Phase:       PhaseWaiting,
		Winner:      NoColor,
	}
}

func (g *Game) Roll(d Dice) error {
	if g.Phase != PhasePlaying && g.Phase != PhaseRollingFirst && g.Phase != PhaseBearingOff {
		return fmt.Errorf("cannot roll in phase %d", g.Phase)
	}
	a, b := d.Roll()
	g.Dice = []int{a, b}
	g.RemainingDice = ExpandDice(a, b)
	if g.Phase == PhaseRollingFirst {
		g.Phase = PhasePlaying
	}
	return nil
}
```

- [ ] **Step 4: Run tests**

Run: `cd backend && go test ./internal/game/ -v`
Expected: все PASS.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/game/game.go backend/internal/game/game_test.go
git commit -m "feat(game): add Game state machine with Roll and phase transitions"
```

---

### Task 1.13: ApplyMove с валидацией и расходом кубиков

**Files:**
- Modify: `backend/internal/game/game.go`
- Modify: `backend/internal/game/game_test.go`

- [ ] **Step 1: Падающие тесты**

В `game_test.go`:
```go
func TestGame_ApplyMove_HappyPath(t *testing.T) {
	g := NewGame()
	g.Phase = PhasePlaying
	g.CurrentTurn = White
	g.Dice = []int{3, 5}
	g.RemainingDice = []int{3, 5}

	err := g.ApplyMove(Move{From: 24, To: 19, Die: 5})

	assert.NoError(t, err)
	assert.Equal(t, []int{3}, g.RemainingDice)
	assert.Equal(t, 14, g.Board.Points[24].Checkers)
	assert.Equal(t, 1, g.Board.Points[19].Checkers)
	assert.Equal(t, 1, g.MoveCount)
}

func TestGame_ApplyMove_InvalidMoveRejected(t *testing.T) {
	g := NewGame()
	g.Phase = PhasePlaying
	g.CurrentTurn = White
	g.RemainingDice = []int{3}

	err := g.ApplyMove(Move{From: 10, To: 7, Die: 3}) // нет шашки на 10

	assert.Error(t, err)
	assert.Equal(t, []int{3}, g.RemainingDice, "remaining dice must not change on rejection")
}

func TestGame_ApplyMove_DieNotInRemaining(t *testing.T) {
	g := NewGame()
	g.Phase = PhasePlaying
	g.CurrentTurn = White
	g.RemainingDice = []int{5}

	err := g.ApplyMove(Move{From: 24, To: 21, Die: 3})

	assert.Error(t, err, "die 3 not in remaining dice")
}

func TestGame_ApplyMove_WrongPhase(t *testing.T) {
	g := NewGame()
	g.Phase = PhaseWaiting

	err := g.ApplyMove(Move{From: 24, To: 21, Die: 3})

	assert.Error(t, err)
}

func TestGame_ApplyMove_NotPlayerTurn(t *testing.T) {
	g := NewGame()
	g.Phase = PhasePlaying
	g.CurrentTurn = Black
	g.RemainingDice = []int{6}

	err := g.ApplyMove(Move{From: 24, To: 18, Die: 6}) // ход белыми

	assert.Error(t, err)
}
```

- [ ] **Step 2: Run — fail**

Run: `cd backend && go test ./internal/game/ -run TestGame_ApplyMove -v`
Expected: FAIL — `Game.ApplyMove undefined`.

- [ ] **Step 3: Реализовать `ApplyMove`**

В `game.go` добавить:
```go
func (g *Game) ApplyMove(m Move) error {
	if g.Phase != PhasePlaying && g.Phase != PhaseBearingOff {
		return fmt.Errorf("cannot move in phase %d", g.Phase)
	}

	// Кубик должен быть в RemainingDice.
	dieIdx := -1
	for i, d := range g.RemainingDice {
		if d == m.Die {
			dieIdx = i
			break
		}
	}
	if dieIdx < 0 {
		return fmt.Errorf("die %d not available in remaining dice %v", m.Die, g.RemainingDice)
	}

	// Проверка владельца ходящего.
	if g.Board.Points[m.From].Owner != g.CurrentTurn {
		return fmt.Errorf("not your checker at %d", m.From)
	}

	if ok, err := IsValidMove(g.Board, g.CurrentTurn, m); !ok {
		return err
	}

	if err := m.Apply(g.Board, g.CurrentTurn); err != nil {
		return err
	}

	g.RemainingDice = append(g.RemainingDice[:dieIdx], g.RemainingDice[dieIdx+1:]...)
	g.MoveCount++

	if g.Board.AllInHome(g.CurrentTurn) && g.Phase == PhasePlaying {
		g.Phase = PhaseBearingOff
	}
	return nil
}
```

- [ ] **Step 4: Run all**

Run: `cd backend && go test ./internal/game/ -v`
Expected: все PASS.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/game/game.go backend/internal/game/game_test.go
git commit -m "feat(game): add Game.ApplyMove with full validation and die consumption"
```

---

### Task 1.14: EndTurn + проверка обязательного использования кубиков

**Files:**
- Modify: `backend/internal/game/game.go`
- Modify: `backend/internal/game/game_test.go`

> **Правило:** игрок не может завершить ход, если возможно использовать ещё хотя бы один кубик. EndTurn проверяет, что либо все кубики использованы, либо для оставшихся нет ни одного валидного хода.

- [ ] **Step 1: Падающие тесты**

В `game_test.go`:
```go
func TestGame_EndTurn_AfterAllDiceUsed(t *testing.T) {
	g := NewGame()
	g.Phase = PhasePlaying
	g.CurrentTurn = White
	g.RemainingDice = []int{}

	err := g.EndTurn()

	assert.NoError(t, err)
	assert.Equal(t, Black, g.CurrentTurn)
	assert.Empty(t, g.Dice)
	assert.Empty(t, g.RemainingDice)
}

func TestGame_EndTurn_RefusedWhenDiceUsable(t *testing.T) {
	g := NewGame()
	g.Phase = PhasePlaying
	g.CurrentTurn = White
	g.RemainingDice = []int{3} // ход с 24 на 21 возможен

	err := g.EndTurn()

	assert.Error(t, err)
	assert.Equal(t, White, g.CurrentTurn, "turn must stay with white")
}

func TestGame_EndTurn_AllowedWhenNoUsableMoves(t *testing.T) {
	// Сценарий: белая шашка на 6, кубик 6 (6→0 = bear-off), но не все в доме.
	// → bear-off не разрешён, обычный ход 6→0 невозможен (out of range).
	// → нет ходов → EndTurn должен пройти.
	g := NewGame()
	g.Phase = PhasePlaying
	g.CurrentTurn = White
	g.Board = &Board{}
	g.Board.Points[6] = Point{Owner: White, Checkers: 1}
	g.Board.Points[24] = Point{Owner: White, Checkers: 14}
	g.RemainingDice = []int{6}

	err := g.EndTurn()

	assert.NoError(t, err)
	assert.Equal(t, Black, g.CurrentTurn)
}
```

- [ ] **Step 2: Run — fail**

Run: `cd backend && go test ./internal/game/ -run TestGame_EndTurn -v`
Expected: FAIL — `Game.EndTurn undefined`.

- [ ] **Step 3: Реализовать `EndTurn`**

В `game.go`:
```go
func (g *Game) EndTurn() error {
	if g.Phase != PhasePlaying && g.Phase != PhaseBearingOff {
		return fmt.Errorf("cannot end turn in phase %d", g.Phase)
	}

	if len(g.RemainingDice) > 0 {
		// Если есть хотя бы один валидный ход с оставшимися кубиками — нельзя завершать.
		sequences := GenerateSequences(g.Board, g.CurrentTurn, g.RemainingDice)
		for _, seq := range sequences {
			if len(seq) > 0 {
				return fmt.Errorf("cannot end turn: %d usable moves remain", len(seq))
			}
		}
	}

	g.CurrentTurn = g.CurrentTurn.Opponent()
	g.Dice = nil
	g.RemainingDice = nil
	return nil
}
```

- [ ] **Step 4: Run all**

Run: `cd backend && go test ./internal/game/ -v`
Expected: все PASS.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/game/game.go backend/internal/game/game_test.go
git commit -m "feat(game): add EndTurn with mandatory-dice-usage check"
```

---

### Task 1.15: Определение победителя + Марс

**Files:**
- Modify: `backend/internal/game/game.go`
- Modify: `backend/internal/game/game_test.go`

- [ ] **Step 1: Падающие тесты**

В `game_test.go`:
```go
func TestGame_Victory_WhiteWins(t *testing.T) {
	g := NewGame()
	g.Phase = PhaseBearingOff
	g.CurrentTurn = White
	g.Board = &Board{}
	g.Board.Points[1] = Point{Owner: White, Checkers: 1}
	g.Board.BorneOff[White] = 14
	g.Board.BorneOff[Black] = 5 // черные что-то выкинули, не марс
	g.RemainingDice = []int{1}

	err := g.ApplyMove(Move{From: 1, To: 0, Die: 1})

	assert.NoError(t, err)
	assert.Equal(t, PhaseFinished, g.Phase)
	assert.Equal(t, White, g.Winner)
	assert.False(t, g.IsMars)
}

func TestGame_Victory_Mars(t *testing.T) {
	g := NewGame()
	g.Phase = PhaseBearingOff
	g.CurrentTurn = White
	g.Board = &Board{}
	g.Board.Points[1] = Point{Owner: White, Checkers: 1}
	g.Board.BorneOff[White] = 14
	g.Board.BorneOff[Black] = 0 // марс
	g.Board.Points[1] = Point{Owner: White, Checkers: 1}
	// чёрные где-то на доске, всё ещё 15
	g.Board.Points[20] = Point{Owner: Black, Checkers: 15}
	g.RemainingDice = []int{1}

	err := g.ApplyMove(Move{From: 1, To: 0, Die: 1})

	assert.NoError(t, err)
	assert.Equal(t, PhaseFinished, g.Phase)
	assert.Equal(t, White, g.Winner)
	assert.True(t, g.IsMars)
}
```

- [ ] **Step 2: Run — fail**

Run: `cd backend && go test ./internal/game/ -run TestGame_Victory -v`
Expected: FAIL — победа не определяется.

- [ ] **Step 3: Дополнить `ApplyMove` детекцией победы**

В `game.go`, в конце `ApplyMove` (после установки `PhaseBearingOff`):
```go
	if g.Board.BorneOff[g.CurrentTurn] == 15 {
		g.Phase = PhaseFinished
		g.Winner = g.CurrentTurn
		g.IsMars = g.Board.BorneOff[g.CurrentTurn.Opponent()] == 0
	}
	return nil
}
```

(Удалить предыдущий `return nil` если он остался.)

- [ ] **Step 4: Run all**

Run: `cd backend && go test ./internal/game/ -v`
Expected: все PASS.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/game/game.go backend/internal/game/game_test.go
git commit -m "feat(game): detect winner and Mars (double victory) condition"
```

---

### Task 1.16: AvailableMoves публичный API

**Files:**
- Modify: `backend/internal/game/game.go`
- Modify: `backend/internal/game/game_test.go`

- [ ] **Step 1: Падающий тест**

В `game_test.go`:
```go
func TestGame_AvailableMoves_ReturnsSequences(t *testing.T) {
	g := NewGame()
	g.Phase = PhasePlaying
	g.CurrentTurn = White
	g.RemainingDice = []int{3, 5}

	sequences := g.AvailableMoves()

	assert.NotEmpty(t, sequences)
	for _, seq := range sequences {
		assert.NotEmpty(t, seq)
		for _, m := range seq {
			assert.Contains(t, []int{3, 5}, m.Die)
		}
	}
}

func TestGame_AvailableMoves_WrongPhase(t *testing.T) {
	g := NewGame()
	g.Phase = PhaseWaiting

	sequences := g.AvailableMoves()

	assert.Nil(t, sequences)
}
```

- [ ] **Step 2: Run — fail**

Run: `cd backend && go test ./internal/game/ -run TestGame_AvailableMoves -v`
Expected: FAIL — `Game.AvailableMoves undefined`.

- [ ] **Step 3: Реализовать**

В `game.go`:
```go
func (g *Game) AvailableMoves() [][]Move {
	if g.Phase != PhasePlaying && g.Phase != PhaseBearingOff {
		return nil
	}
	if len(g.RemainingDice) == 0 {
		return nil
	}
	return GenerateSequences(g.Board, g.CurrentTurn, g.RemainingDice)
}
```

- [ ] **Step 4: Run all**

Run: `cd backend && go test ./internal/game/ -v`
Expected: все PASS.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/game/game.go backend/internal/game/game_test.go
git commit -m "feat(game): expose AvailableMoves as public Game API"
```

---

### Task 1.17: Бросок на первый ход (RollFirst)

**Files:**
- Modify: `backend/internal/game/game.go`
- Modify: `backend/internal/game/game_test.go`

> **Правило:** в начале партии оба игрока бросают по одной кости; у кого больше — ходит первым. Если равенство — переброс.

- [ ] **Step 1: Падающие тесты**

В `game_test.go`:
```go
func TestGame_RollFirst_WhiteWinsAndStarts(t *testing.T) {
	g := NewGame()
	g.Phase = PhaseWaiting

	err := g.RollFirst(NewFixedDice([][2]int{{5, 3}}))

	assert.NoError(t, err)
	assert.Equal(t, PhasePlaying, g.Phase)
	assert.Equal(t, White, g.CurrentTurn, "white rolled 5, black rolled 3 → white starts")
	assert.Equal(t, []int{5, 3}, g.Dice)
	assert.Equal(t, []int{5, 3}, g.RemainingDice)
}

func TestGame_RollFirst_BlackWins(t *testing.T) {
	g := NewGame()
	g.Phase = PhaseWaiting

	err := g.RollFirst(NewFixedDice([][2]int{{2, 6}}))

	assert.NoError(t, err)
	assert.Equal(t, Black, g.CurrentTurn)
	assert.Equal(t, []int{2, 6}, g.RemainingDice)
}

func TestGame_RollFirst_Tie_Rerolls(t *testing.T) {
	g := NewGame()
	g.Phase = PhaseWaiting
	d := NewFixedDice([][2]int{{4, 4}, {2, 5}})

	err := g.RollFirst(d)

	assert.NoError(t, err)
	assert.Equal(t, Black, g.CurrentTurn, "tie 4-4, then 2-5 → black wins")
	assert.Equal(t, []int{2, 5}, g.Dice)
}

func TestGame_RollFirst_WrongPhase(t *testing.T) {
	g := NewGame()
	g.Phase = PhasePlaying

	err := g.RollFirst(NewFixedDice([][2]int{{1, 2}}))

	assert.Error(t, err)
}
```

- [ ] **Step 2: Run — fail**

Run: `cd backend && go test ./internal/game/ -run TestGame_RollFirst -v`
Expected: FAIL — `Game.RollFirst undefined`.

- [ ] **Step 3: Реализовать**

В `game.go`:
```go
func (g *Game) RollFirst(d Dice) error {
	if g.Phase != PhaseWaiting {
		return fmt.Errorf("RollFirst allowed only in PhaseWaiting, got %d", g.Phase)
	}
	for {
		a, b := d.Roll()
		if a == b {
			continue
		}
		if a > b {
			g.CurrentTurn = White
		} else {
			g.CurrentTurn = Black
		}
		g.Dice = []int{a, b}
		g.RemainingDice = ExpandDice(a, b)
		g.Phase = PhasePlaying
		return nil
	}
}
```

- [ ] **Step 4: Run all**

Run: `cd backend && go test ./internal/game/ -v`
Expected: все PASS.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/game/game.go backend/internal/game/game_test.go
git commit -m "feat(game): add RollFirst to determine starting player with tie-breaks"
```

---

### Task 1.18: Финальная проверка покрытия и тег Phase 1

**Files:** нет (только проверки).

- [ ] **Step 1: Запустить все тесты с подсчётом покрытия**

Run:
```bash
cd backend
go test ./internal/game/... -cover -coverprofile=coverage.out
go tool cover -func=coverage.out | tail -5
```

Expected: total coverage ≥ 90%. Если меньше — добавить тесты на непокрытые ветви.

- [ ] **Step 2: Запустить с race detector**

Run: `cd backend && go test ./internal/game/... -race`
Expected: PASS без data races.

- [ ] **Step 3: Запустить линтер (опционально)**

Если установлен `golangci-lint`:
```bash
cd backend && golangci-lint run ./internal/game/...
```
Expected: 0 issues. Если линтер не установлен — пропустить, добавить установку в Plan 5.

- [ ] **Step 4: Сводка тестов**

Run: `cd backend && go test ./internal/game/... -v 2>&1 | grep -E "^(=== RUN|--- (PASS|FAIL))" | wc -l`
Expected: количество тестов ≥ 30 (примерная оценка по всем задачам).

- [ ] **Step 5: Тег Phase 1**

```bash
git tag phase-1-game-logic
git log --oneline phase-0-scaffolding..phase-1-game-logic
```

Expected: видны все коммиты Phase 1.

---

## Verification & Handoff

После выполнения всех задач:

| Проверка | Команда | Ожидаемый результат |
|---|---|---|
| Сборка Go | `cd backend && go build ./...` | Без ошибок |
| Тесты Go с покрытием | `cd backend && go test ./internal/game/... -cover` | ≥ 90% |
| Тесты Go race | `cd backend && go test ./internal/game/... -race` | PASS |
| Билд Next.js | `cd frontend && npm run build` | Успешен |
| Docker compose | `docker compose config` | Без ошибок |
| Git теги | `git tag` | `phase-0-scaffolding`, `phase-1-game-logic` |

После этого:
1. Запустить `advisor()` для финального ревью движка перед началом Plan 2.
2. Перейти к Plan 2 (DB + REST API), который зависит от публичного API пакета `game`.

---

## Open questions / known gaps

Эти моменты вынесены за рамки Plan 1 — реализация в последующих планах:

| Вопрос | Решение | План |
|---|---|---|
| Сериализация Board в JSONB | `json.Marshal`/`json.Unmarshal` пакета `game` — добавить в Plan 2 | Plan 2 |
| HTTP API над Game | REST + WS handlers | Plan 2 + Plan 3 |
| Таймеры ходов (60 сек) | Логика на стороне WS-хаба, не в `game` | Plan 3 |
| Отображение шашек, кубиков | Компоненты React | Plan 4 |
| Чат и лобби | Отдельные модули | Plan 2 (бэк), Plan 4 (фронт) |
