# Web-Backgammon — Design Document

> Онлайн-игра в **длинные нарды** для двух игроков с разных устройств.
> Целевая платформа: современные смартфоны (приоритет), планшеты, десктоп.
> Стек: **Next.js + Go + PostgreSQL**, реалтайм через **WebSocket**, развёртывание в **Docker Compose**.

---

## Сводка решений

| Параметр | Решение |
|---|---|
| Вариант нарды | Длинные нарды |
| Фронтенд | React + Next.js 14+ (App Router) |
| Бэкенд | Go (chi + nhooyr.io/websocket) |
| Реалтайм | WebSocket |
| Матчмейкинг | По ссылке/коду (8 символов base32) |
| Авторизация | Имя при входе по ссылке (session_token в httpOnly cookie) |
| Рендер доски | HTML/CSS + SVG (адаптивный viewBox) |
| Хранилище | PostgreSQL 16 |
| Деплой | Docker Compose |
| Визуальный стиль | Неоморфный |
| Анимации | Полные (Framer Motion) |
| Общение | Текстовый чат |
| Язык UI | Русский |
| История | Только результаты партий |

---

## Архитектура — Подход A: Монолитный Next.js + Go-бэкенд

```
[Next.js App] ←--REST + WS--→ [Go Server] ←→ [PostgreSQL]
   (frontend)                   (backend)        (db)
```

**Плюсы:**
- Чёткое разделение: UI на Next.js, вся логика на Go.
- Go идеален для конкурентного WS-сервера и игровой логики.
- Next.js даёт SSR для лендинга + SPA для игры.
- Простой Docker Compose деплой.

**Минусы:**
- Два процесса для деплоя.
- Нужно настраивать CORS и проксирование WS.

---

## Секция 1: Структура проекта

```
Web-backgammon/
├── frontend/                # Next.js приложение
│   ├── src/
│   │   ├── app/             # App Router (pages, layouts)
│   │   ├── components/      # React-компоненты
│   │   │   ├── board/       # Доска, шашки, кубики
│   │   │   ├── chat/        # Текстовый чат
│   │   │   ├── lobby/       # Лобби, создание комнаты
│   │   │   └── ui/          # Общие UI-компоненты (кнопки, инпуты)
│   │   ├── hooks/           # Custom hooks (useWebSocket, useGame)
│   │   ├── lib/             # Утилиты, типы, константы
│   │   ├── stores/          # Zustand-сторы (game, chat, ui)
│   │   └── styles/          # Глобальные стили, тема
│   ├── public/              # Статика (SVG доски, звуки)
│   └── package.json
├── backend/                 # Go-сервер
│   ├── cmd/server/          # Точка входа (main.go)
│   ├── internal/
│   │   ├── game/            # Игровая логика (правила, ходы, валидация)
│   │   ├── room/            # Управление комнатами
│   │   ├── ws/              # WebSocket хаб, клиенты, сообщения
│   │   ├── db/              # PostgreSQL репозитории
│   │   └── api/             # REST API
│   ├── migrations/          # SQL-миграции
│   └── go.mod
├── docker-compose.yml       # Frontend + Backend + PostgreSQL
└── docs/                    # Документация
```

**Ключевые библиотеки:**
- Frontend: Zustand (state), Framer Motion (анимации), Tailwind CSS (стили), TypeScript.
- Backend: `chi` (роутинг), `pgx/v5` (PostgreSQL), `nhooyr.io/websocket` (WS), `golang-migrate` (миграции).

---

## Секция 2: Игровая логика (длинные нарды)

### Правила

**Начальная расстановка:**
- 15 белых шашек на позиции 24.
- 15 чёрных шашек на позиции 1.
- Остальные 22 пункта свободны.

**Направление движения:**
- Белые: 24 → 1 (убыль номеров).
- Чёрные: 1 → 24 (рост номеров).

**Бросок кубиков:**
- Два кубика одновременно.
- Дубль (одинаковые) = 4 хода вместо 2.
- Значения кубиков используются по отдельности — складывать нельзя.

**Движение шашек:**
- Шашка перемещается на число пунктов, выпавшее на одном кубике.
- Нельзя ставить шашку на пункт, занятый хотя бы одной шашкой соперника (удара нет).
- Нельзя строить "глухой забор" — 6 и более занятых пунктов подряд, блокирующих все шашки соперника.
- Если ходов нет — пропуск хода.
- Если можно использовать только один кубик — обязан больший.
- Если можно оба — обязан использовать оба.

**Выкидывание (дом):**
- Дом для белых: пункты 1–6. Дом для чёрных: пункты 19–24.
- Выкидывание начинается после захода всех 15 шашек в дом.
- Если на пункте, равном кубику, нет шашки — снять с более старшего пункта.
- Если на более старшем тоже нет — сделать ход внутри дома.

**Победа:**
- Первый выкинувший все 15 шашек выигрывает.
- **Марс** — соперник не выкинул ни одной шашки (двойная победа).

### Пограничные случаи

| Случай | Обработка |
|---|---|
| Нет доступных ходов | Автоматический пропуск с уведомлением |
| Дубль | 4 хода, каждый по значению кубика |
| "Глухой забор" (6+ подряд) | Ход отклоняется, показывается причина |
| Частичное использование кубиков | Обязан использовать максимум возможных |
| Обязательный больший кубик | Если из двух можно один — берём больший |
| Выкидывание: нет шашки на пункте кубика | Снять с ближайшего старшего, либо ход внутри |
| Оба игрока в доме | Стандартное выкидывание, первый выкинувший — победил |
| Отключение игрока | 60 сек таймер на ход, затем автопропуск; 5 мин grace на возврат, затем поражение |

### Реализация на Go

```
internal/game/
├── board.go       # Структура доски, начальная расстановка
├── rules.go       # Валидация ходов, проверка блоков, обязательные ходы
├── moves.go       # Генерация доступных ходов, алгоритм перебора
├── dice.go        # Бросок кубиков, обработка дублей
├── bearoff.go     # Логика выкидывания из дома
├── validator.go   # Полная валидация хода + пограничные случаи
└── game.go        # Игровой цикл, смена ходов, определение победителя
```

**Ключевой алгоритм:** при каждом ходе генерируем полное дерево доступных ходов с учётом обязательности использования обоих кубиков и приоритета большего. Пустое дерево → автоматический пропуск.

---

## Секция 3: Протокол WebSocket

### От клиента к серверу

```jsonc
{ "type": "move",     "payload": { "from": 24, "to": 18, "die": 6 } }
{ "type": "end_turn", "payload": {} }
{ "type": "pass",     "payload": {} }
{ "type": "chat",     "payload": { "text": "Привет!" } }
{ "type": "ping",     "payload": {} }
```

### От сервера к клиенту

```jsonc
{ "type": "game_state",            "payload": { /* полный snapshot */ } }
{ "type": "opponent_moved",        "payload": { "from": 24, "to": 18, "die": 6, "remainingDice": [3] } }
{ "type": "dice_rolled",           "payload": { "dice": [6, 3], "isDouble": false } }
{ "type": "turn_changed",          "payload": { "player": "black", "timeLeft": 60 } }
{ "type": "move_error",            "payload": { "reason": "glukhoi_zabor" } }
{ "type": "chat_message",          "payload": { "from": "Алексей", "text": "Привет!", "time": "20:45" } }
{ "type": "opponent_disconnected", "payload": { "gracePeriod": 300 } }
{ "type": "opponent_reconnected",  "payload": {} }
{ "type": "game_over",             "payload": { "winner": "white", "isMars": false } }
{ "type": "pong",                  "payload": {} }
```

### Жизненный цикл игры

```
Создание комнаты → Ожидание игрока → Оба подключились →
Бросок на первый ход → Игровой цикл → Победа / Отключение
```

**Фазы:**
1. `waiting` — ждём второго игрока.
2. `rolling_first` — бросок на первый ход.
3. `playing` — основной игровой цикл.
4. `bearing_off` — один из игроков в фазе выкидывания.
5. `finished` — игра завершена.

### Надёжность

| Ситуация | Поведение |
|---|---|
| Ping-pong каждые 15 сек | Обнаружение разрыва за 30 сек |
| Отключение во время хода | Таймер 60 сек продолжает тикать |
| Отключение соперника | Уведомление + 5 мин grace period |
| Возвращение за 5 мин | Полная ресинхронизация через `game_state` |
| Не вернулся за 5 мин | Автопобеда оставшегося |
| Потеря WS-сообщения | Клиент запрашивает полное состояние |
| Переподключение | Привязка по `roomID + session_token` |

### Реализация WS-хаба

```
internal/ws/
├── hub.go        # Регистрация/удаление клиентов, broadcast по комнатам
├── client.go     # Один WS-клиент: read/write pumps, ping/pong
├── message.go    # Типы сообщений, парсинг JSON
└── handler.go    # HTTP → WS upgrade, привязка к комнате
```

---

## Секция 4: Фронтенд-архитектура и UI

### Стек

| Технология | Назначение |
|---|---|
| Next.js 14+ (App Router) | Фреймворк, SSR для лендинга, SPA для игры |
| TypeScript | Типобезопасность |
| Zustand | Стейт-менеджмент (game, chat, ui) |
| Framer Motion | Анимации шашек, кубиков, переходов |
| Tailwind CSS | Стилизация + неоморфный дизайн |
| SVG | Рендер доски и шашек |

### Роутинг

```
/                 → Лендинг: "Создать игру" / "Войти по коду"
/room/[code]      → Ожидание соперника (лоадер + код комнаты)
/game/[code]      → Игровая доска + чат
/game/[code]/result → Результат партии
```

### Zustand-сторы

```ts
// stores/gameStore.ts
interface GameStore {
  board: Point[];
  dice: number[];
  remainingDice: number[];
  turn: 'white' | 'black';
  phase: GamePhase;
  myColor: 'white' | 'black';
  selectedChecker: number | null;
  validMoves: Move[];
  timeLeft: number;
  // actions
  moveChecker: (from: number, to: number, die: number) => void;
  endTurn: () => void;
  selectChecker: (point: number) => void;
}

// stores/chatStore.ts
interface ChatStore {
  messages: ChatMessage[];
  addMessage: (msg: ChatMessage) => void;
}

// stores/uiStore.ts
interface UIStore {
  showSettings: boolean;
  animationsEnabled: boolean;
  soundEnabled: boolean;
}
```

### Компоненты доски

```
components/board/
├── Board.tsx              # SVG-доска целиком, адаптивная сетка
├── Point.tsx              # Один пункт (треугольник) — кликабельный
├── Checker.tsx            # Шашка — неоморфный стиль, drag-анимация
├── CheckerStack.tsx       # Стек шашек на пункте (до 15)
├── Dice.tsx               # Кубики с анимацией броска
├── Bar.tsx                # Центральная полоса (кубики + бар)
├── BearOffZone.tsx        # Зона выкидывания
├── ValidMoveIndicator.tsx # Подсветка доступных ходов
└── PlayerInfo.tsx         # Имя, цвет, таймер, статус
```

### Неоморфный дизайн

**Палитра:**
```
Фон:          #e0e5ec  (мягкий серо-голубой)
Светлая тень: #ffffff
Тёмная тень:  #a3b1c6
Акцент:       #6c63ff  (фиолетовый)
Белые шашки:  #f0f0f0 с неоморфной тенью
Чёрные шашки: #3a3a3a с неоморфной тенью
Доска:        #2d5016 (тёмно-зелёный) + #8B4513 (дерево)
```

**Классы:**
```css
.neo-inset {  /* вдавленный элемент */
  background: #e0e5ec;
  box-shadow: inset 6px 6px 12px #a3b1c6,
              inset -6px -6px 12px #ffffff;
}

.neo-raised { /* выпуклый элемент */
  background: #e0e5ec;
  box-shadow: 6px 6px 12px #a3b1c6,
              -6px -6px 12px #ffffff;
}
```

### Анимации

| Элемент | Анимация |
|---|---|
| Перемещение шашки | `layout` Framer Motion — плавный slide (300ms, ease-out) |
| Бросок кубиков | 3D-вращение + отскок (600ms) с рандомной задержкой |
| Подсветка ходов | Пульсация (opacity 0.4 → 0.8, 1s loop) |
| Появление чат-сообщения | Slide-up + fade-in (200ms) |
| Смена хода | Подсветка имени игрока (scale 1.0 → 1.05 → 1.0) |
| Победа | Конфетти-эффект + масштабирование текста |

### Мобильная адаптация

**Целевой диапазон CSS-viewport:**

| Категория | Ширина | Примеры |
|---|---|---|
| Компактные | 375–390px | iPhone SE, iPhone mini |
| Стандартные флагманы | 393–430px | iPhone 15/16, Samsung S24 |
| Крупные флагманы | 412–460px | OnePlus 9 Pro (412×919), iPhone 17 Pro Max (440×956) |

**Ключевые правила:**
- Доска занимает 90–95% ширины viewport, max-width 480px.
- SVG-доска через `viewBox` + `preserveAspectRatio="xMidYMeet"` — масштабируется без потери качества при DPR 3x и 3.5x.
- Тач-области минимум 44×44px CSS.
- Учёт безопасных зон: `env(safe-area-inset-*)` для iPhone с Dynamic Island.
- Брейкпоинты:
  - `max-width: 480px` — мобильная (вертикальная доска, чат-шторка снизу).
  - `481–1024px` — планшет (горизонтальная доска, чат сбоку).
  - `1025px+` — десктоп (горизонтальная доска, чат справа).

Для соотношения **20:9** (OnePlus 9 Pro): вертикальная доска ~412×700px, под ней ~200px для кубиков и инфо-панели, чат — выдвижная шторка поверх.

---

## Секция 5: Схема базы данных PostgreSQL

```sql
CREATE TABLE rooms (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  code            VARCHAR(8) UNIQUE NOT NULL,
  status          VARCHAR(20) NOT NULL,                -- waiting | playing | finished | abandoned
  created_at      TIMESTAMPTZ DEFAULT NOW(),
  expires_at      TIMESTAMPTZ NOT NULL
);
CREATE INDEX idx_rooms_code ON rooms(code);
CREATE INDEX idx_rooms_expires ON rooms(expires_at) WHERE status != 'finished';

CREATE TABLE players (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  room_id         UUID NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
  name            VARCHAR(40) NOT NULL,
  color           VARCHAR(5),                          -- white | black, NULL до распределения
  session_token   VARCHAR(64) UNIQUE NOT NULL,
  joined_at       TIMESTAMPTZ DEFAULT NOW(),
  last_seen_at    TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_players_room ON players(room_id);
CREATE INDEX idx_players_session ON players(session_token);

CREATE TABLE games (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  room_id         UUID UNIQUE NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
  board_state     JSONB NOT NULL,
  current_turn    VARCHAR(5) NOT NULL,
  dice            INTEGER[] NOT NULL DEFAULT '{}',
  remaining_dice  INTEGER[] NOT NULL DEFAULT '{}',
  phase           VARCHAR(20) NOT NULL,
  winner          VARCHAR(5),
  is_mars         BOOLEAN DEFAULT FALSE,
  turn_started_at TIMESTAMPTZ DEFAULT NOW(),
  move_count      INTEGER DEFAULT 0,
  created_at      TIMESTAMPTZ DEFAULT NOW(),
  updated_at      TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_games_room ON games(room_id);

CREATE TABLE moves (
  id              BIGSERIAL PRIMARY KEY,
  game_id         UUID NOT NULL REFERENCES games(id) ON DELETE CASCADE,
  move_number     INTEGER NOT NULL,
  player_color    VARCHAR(5) NOT NULL,
  dice_rolled     INTEGER[] NOT NULL,
  moves_data      JSONB NOT NULL,
  created_at      TIMESTAMPTZ DEFAULT NOW(),
  UNIQUE(game_id, move_number)
);
CREATE INDEX idx_moves_game ON moves(game_id);

CREATE TABLE game_results (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  room_id         UUID NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
  winner_name     VARCHAR(40) NOT NULL,
  loser_name      VARCHAR(40) NOT NULL,
  is_mars         BOOLEAN DEFAULT FALSE,
  total_moves     INTEGER NOT NULL,
  duration_sec    INTEGER NOT NULL,
  finished_at     TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_results_finished ON game_results(finished_at DESC);

CREATE TABLE chat_messages (
  id              BIGSERIAL PRIMARY KEY,
  room_id         UUID NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
  player_id       UUID NOT NULL REFERENCES players(id) ON DELETE CASCADE,
  text            VARCHAR(500) NOT NULL,
  created_at      TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_chat_room ON chat_messages(room_id, created_at);
```

**Решения:**
| Решение | Обоснование |
|---|---|
| `board_state` как JSONB | Быстрая запись/чтение всего состояния, индексируемые поля |
| Хранение всех ходов | Восстановление партии + детект читерства |
| `session_token` в cookie | Ресинхронизация после потери WS без логина |
| `expires_at` на rooms | Авточистка брошенных комнат (фоновый job каждый час) |
| `move_count` в games | Защита от устаревших WS-команд (идемпотентность) |
| Отдельная `game_results` | Не зависит от удаления комнаты, история сохраняется |

**Миграции:** `golang-migrate/migrate`, версионируемые SQL-файлы в `backend/migrations/`.

---

## Секция 6: REST API

| Метод | Путь | Описание |
|---|---|---|
| `POST` | `/api/rooms` | Создать комнату → `{id, code, url, sessionToken}` |
| `GET` | `/api/rooms/:code` | Инфо о комнате (статус, кол-во игроков) |
| `POST` | `/api/rooms/:code/join` | Войти в комнату (`{name}`) → `session_token` в cookie |
| `GET` | `/api/games/:roomId/state` | Состояние игры (для ресинхронизации) |
| `GET` | `/api/games/:roomId/history` | История ходов |
| `GET` | `/api/health` | Health check |
| `WS` | `/ws/:roomCode` | WebSocket-подключение (требует session_token) |

### Примеры

**POST /api/rooms**
```jsonc
// Запрос
{ "creatorName": "Алексей" }

// Ответ 201
{
  "id": "550e8400-...",
  "code": "X7K2QM",
  "url": "/game/X7K2QM",
  "sessionToken": "..."
}
```

**POST /api/rooms/:code/join**
```jsonc
// Запрос
{ "name": "Мария" }

// Ответ 200
{ "playerId": "...", "color": null, "sessionToken": "..." }

// Ошибки
404: комната не найдена
410: комната заполнена / завершена
422: имя невалидно (>40 символов, пустое, запрещённые символы)
```

### Rate limiting

| Endpoint | Лимит |
|---|---|
| `POST /api/rooms` | 5/мин с IP |
| `POST /api/rooms/:code/join` | 10/мин с IP |
| WS chat | 5/сек с клиента |
| WS move | 30/сек с клиента |

Реализация: `golang.org/x/time/rate`.

---

## Секция 7: Безопасность и анти-чит

**Принцип:** сервер — единственный источник истины.

| Угроза | Защита |
|---|---|
| Подмена результата кубиков | Кубики бросаются только на сервере |
| Невалидный ход | Полная валидация на сервере, клиент-валидация только для UX |
| Игра за оппонента | Привязка хода к `session_token` + проверка `current_turn == player.color` |
| Replay-атака на WS | `move_count` в каждом ходе |
| XSS через имя/чат | Санитизация на бэкенде (`bluemonday`) + React автоматический escape |
| CSRF на REST | `SameSite=Strict` cookie + CORS allowlist |
| Брутфорс кодов комнат | 8 символов base32 (A-Z, 2-7) = 2^40 вариантов, rate-limit на join |
| Спам комнатами | Rate limit `POST /api/rooms` |
| DoS через WS | Лимит размера сообщения 1KB, max 1 WS на session_token |
| Уязвимости Go-зависимостей | `govulncheck` в CI |
| Уязвимости npm | `npm audit` + Dependabot |

### HTTP-заголовки

```
Content-Security-Policy: default-src 'self'; connect-src 'self' wss://...;
X-Frame-Options: DENY
X-Content-Type-Options: nosniff
Strict-Transport-Security: max-age=31536000
Referrer-Policy: strict-origin-when-cross-origin
```

### Cookie

```
session_token=...; HttpOnly; Secure; SameSite=Strict; Max-Age=86400; Path=/
```

---

## Секция 8: Docker Compose deployment

### docker-compose.yml

```yaml
services:
  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_DB: backgammon
      POSTGRES_USER: bg_user
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD}
    volumes:
      - postgres_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U bg_user"]
      interval: 5s
      timeout: 5s
      retries: 5

  backend:
    build: ./backend
    environment:
      DATABASE_URL: postgres://bg_user:${POSTGRES_PASSWORD}@postgres:5432/backgammon
      PORT: 8080
      JWT_SECRET: ${JWT_SECRET}
      ALLOWED_ORIGINS: http://localhost:3000,https://yourdomain.com
    depends_on:
      postgres:
        condition: service_healthy
    ports:
      - "8080:8080"

  frontend:
    build: ./frontend
    environment:
      NEXT_PUBLIC_API_URL: http://localhost:8080
      NEXT_PUBLIC_WS_URL: ws://localhost:8080
    ports:
      - "3000:3000"
    depends_on:
      - backend

  nginx:
    image: nginx:alpine
    profiles: [production]
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf:ro
      - ./certs:/etc/letsencrypt:ro
    ports: ["80:80", "443:443"]
    depends_on: [frontend, backend]

volumes:
  postgres_data:
```

### Образы

- **Backend Dockerfile** — `golang:alpine` (build) → `gcr.io/distroless/static` (~10MB).
- **Frontend Dockerfile** — `node:alpine` (build) → `node:alpine-slim` со standalone-output Next.js (~150MB).

### Окружения

```
.env.development   # локальная разработка
.env.production    # production (gitignored, через secrets manager)
.env.test          # CI
```

---

## Секция 9: Стратегия тестирования

### Пирамида

| Уровень | Инструменты | Что покрываем |
|---|---|---|
| Unit (бэкенд) | Go `testing` + `testify` | Все правила нарды, ходы, выкидывание, дубль, "глухой забор", обязательность большего |
| Unit (фронт) | Vitest + React Testing Library | Логика сторов, утилиты, компоненты в изоляции |
| Integration (бэкенд) | Go + testcontainers (PostgreSQL) | REST endpoints + БД, WS handshake, ресинхронизация |
| E2E | Playwright | 2 браузера = 2 игрока, полная партия, мобильный viewport, чат |
| Visual regression | Playwright screenshots | Доска, шашки, кубики на iPhone/Android viewport |
| Performance | Lighthouse CI + custom WS latency tests | <100ms WS roundtrip, <2.5s LCP, бандл <300KB gzipped |

### Критические тест-кейсы для правил нарды

- Начальная расстановка верна (15 шашек на 24 для белых, 15 на 1 для чёрных).
- Белые двигаются 24→1, чёрные 1→24.
- Дубль даёт 4 хода.
- Запрет постановки на пункт с шашкой соперника.
- Запрет "глухого забора" — 6+ подряд блокируют все ходы оппонента.
- Обязательное использование обоих кубиков, если возможно.
- Если возможен только 1 кубик — обязан больший.
- Пропуск хода при отсутствии возможных ходов.
- Старт выкидывания только после захода всех 15 в дом.
- Выкидывание с более старшего пункта если кубик "промахивается".
- Победа при выкидывании всех 15 шашек.
- Марс — соперник не выкинул ни одной.
- Таймер хода: автопропуск через 60 сек.
- Disconnect/reconnect с полной ресинхронизацией.

### E2E сценарии

1. Создание комнаты → второй игрок входит → партия → победа.
2. Дисконнект во время хода → реконнект → продолжение.
3. Чат во время игры → доставка обоим.
4. Мобильный viewport (iPhone 17 Pro Max, OnePlus 9 Pro) → доска кликабельна, чат открывается.
5. Симуляция 3G throttle → корректное отображение лоадеров.

---

## Секция 10: Производительность

### Бэкенд

| Оптимизация | Метрика |
|---|---|
| Игровое состояние в памяти (in-memory hub) + async sync в БД | <5ms на ход |
| Pre-allocated message buffers для WS | Меньше GC pressure |
| Permessage-deflate compression на WS | -60% трафика для game_state |
| Connection pooling (`pgx` pool, 25 соединений) | Нет лагов на JOIN |
| Bulk insert ходов (batch при завершении партии) | Меньше нагрузка на БД |

### Фронтенд

| Оптимизация | Цель |
|---|---|
| Code splitting: лендинг отдельно от игры | Initial JS <100KB |
| SVG inline для доски | Нет лишних HTTP-запросов |
| `next/image` для статичных изображений | Автоматический WebP/AVIF |
| `next/font` для шрифтов | Нет FOUT, preload |
| React Server Components где возможно | Меньше JS на клиенте |
| `React.memo` для Checker/Point | Не ререндер 24 пунктов на каждое движение |
| Framer Motion `layoutId` для шашек | GPU-ускоренные анимации |
| Service Worker для офлайн-страницы | UX при разрыве |

### Целевые метрики

```
LCP:    < 2.5s
FID:    < 100ms
CLS:    < 0.1
WS RTT: < 100ms (RU-EU регион)
TTI:    < 3.5s на 3G
Bundle (game route, gzipped): < 300KB
```

---

## Секция 11: Инструменты и плагины

### MCP-серверы (установлены)

| MCP | Использование |
|---|---|
| **playwright** | E2E-тесты, визуальная регрессия, mobile viewport, дебаг UI |
| **context7** | Актуальная документация по Next.js 14+, Framer Motion, Zustand, nhooyr/websocket |
| **chrome-devtools** | Профилирование, network throttling, мобильная эмуляция |

### Skills (доступны)

| Skill | Применение |
|---|---|
| `superpowers:writing-plans` | Финальный implementation plan |
| `superpowers:test-driven-development` | Реализация игровых правил через TDD |
| `superpowers:systematic-debugging` | Дебаг WS/синхронизации |
| `superpowers:dispatching-parallel-agents` | Параллельная работа агентов |
| `superpowers:subagent-driven-development` | Выполнение задач плана через subagents |
| `superpowers:requesting-code-review` | Ревью между фазами |
| `vercel-react-best-practices` | Оптимизация React/Next.js |
| `frontend-design` | Неоморфные компоненты |
| `next-best-practices` | App Router, RSC, метаданные |
| `framer-motion` | Анимации шашек, кубиков |
| `shadcn-ui` | UI-компоненты лобби/чата/настроек (опционально) |

### Дополнительные инструменты

**Go:**
```
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install golang.org/x/vuln/cmd/govulncheck@latest
go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest
go install gotest.tools/gotestsum@latest
go install github.com/air-verse/air@latest
```

**Node.js (dev):**
```
npm install -D @biomejs/biome
npm install -D vitest @vitest/ui
npm install -D @testing-library/react
npm install -D @playwright/test
npm install -D @next/bundle-analyzer
```

**Frontend runtime:**
```
next@^14            framer-motion       zustand
react@^18           clsx                tailwindcss
typescript          immer               canvas-confetti
```

**Backend runtime:**
```
github.com/go-chi/chi/v5            github.com/jackc/pgx/v5
nhooyr.io/websocket                 github.com/golang-migrate/migrate/v4
github.com/google/uuid              github.com/microcosm-cc/bluemonday
github.com/joho/godotenv            golang.org/x/time
github.com/stretchr/testify         github.com/testcontainers/testcontainers-go
```

**CI/CD:**
- GitHub Actions (или GitLab CI) — линт + тесты + сборка на каждый PR.
- Docker Hub / GHCR — публикация образов.
- Sentry (опционально) — error tracking в production.

**Мониторинг (production):**
- Prometheus + Grafana — метрики Go (WS, latency, активные комнаты).
- Loki — централизованные логи.

---

## Секция 12: Пайплайн разработки с несколькими агентами

### Принцип

Главный агент (координатор) формирует контексты, использует `superpowers:dispatching-parallel-agents` для параллельной работы над независимыми модулями и `superpowers:subagent-driven-development` для последовательных задач внутри фазы.

### Фаза 0: Инициализация (без агентов)
- Структура папок `frontend/`, `backend/`.
- `go.mod`, `package.json` с базовыми зависимостями.
- `.gitignore`, `README.md`, ENV-шаблоны.
- Базовый `docker-compose.yml` с PostgreSQL.

### Фаза 1: Игровая логика (TDD, изолированно)
**Один агент, последовательно:**
- Активирует `superpowers:test-driven-development`.
- Реализует `internal/game/*` — board, dice, moves, rules, bearoff, validator, game.
- Все 14+ правил длинных нарды + edge cases как unit-тесты.
- Деливерабл: пакет `game` с публичным API, `go test ./internal/game/...` с покрытием >90%.

### Фаза 2: Параллельная разработка инфраструктуры (3 агента)

**Агент 2A — БД и миграции (backend):**
- `internal/db/*` — pgx pool, репозитории.
- Миграции (`migrations/`).
- Интеграционные тесты через testcontainers.

**Агент 2B — REST API (backend):**
- `internal/api/*` + `cmd/server/main.go`.
- Endpoints для комнат, middleware (auth, rate-limit, CORS, logger).
- Мок репозиториев на время разработки.

**Агент 2C — Каркас фронтенда:**
- Next.js scaffolding, App Router.
- Tailwind config с неоморфной палитрой.
- Базовые UI-компоненты (Button, Input, Card).
- Zustand stores (структура).
- Лендинг + страница `/room/[code]` (без WS).

**После фазы:** точка интеграции — главный агент собирает результаты, разрешает конфликты.

### Фаза 3: WebSocket-протокол
**Один агент:**
- `internal/ws/*` — hub, client, message routing.
- Использует пакет `game` из Фазы 1.
- Все 11 типов сообщений, reconnect-логика, ping/pong, grace period.
- Интеграционные тесты с реальными WS-клиентами.

### Фаза 4: Параллельная разработка UI (3 агента)

**Агент 4A — Доска и шашки:**
- Активирует `frontend-design` + `framer-motion`.
- Board, Point, Checker, CheckerStack, BearOffZone.
- SVG-доска адаптивная (`viewBox`).
- Анимации перемещения (`layoutId`).
- Тестирование на mobile viewport через Playwright MCP.

**Агент 4B — Кубики и UI вокруг доски:**
- Dice (3D-анимация), PlayerInfo, TurnTimer.
- Анимация броска через CSS transforms + Framer Motion.
- Mobile-адаптация (тап на кубик = передача хода).

**Агент 4C — Чат и лобби:**
- Лобби (ввод имени, ожидание соперника, копирование ссылки).
- Чат: desktop sidebar / mobile bottom sheet.
- Тосты при дисконнекте/реконнекте.

### Фаза 5: Интеграция фронт ↔ бэк
**Один агент:**
- `useWebSocket` hook.
- Подключение Zustand-сторов к WS-сообщениям.
- Полный flow: создание → ожидание → игра → результат.
- Обработка ошибок, reconnect-логика на клиенте.
- Деливерабл: два браузера играют полную партию.

### Фаза 6: Тестирование и полировка (2 агента параллельно)

**Агент 6A — E2E через Playwright:**
- Сценарии из Секции 9.
- Mobile viewport (iPhone 17 Pro Max, OnePlus 9 Pro).
- Slow network тесты.

**Агент 6B — Performance & a11y:**
- Профилирование через chrome-devtools MCP.
- Lighthouse CI.
- Bundle analysis.
- A11y: ARIA, keyboard nav, контрастность.
- Активирует `vercel-react-best-practices`.

### Фаза 7: Deployment и documentation
**Один агент:**
- Production `docker-compose.yml`.
- Nginx config с TLS.
- README с инструкциями.
- CI/CD workflow.

### Контракты между фазами

| Контракт | Артефакт | Когда |
|---|---|---|
| Game API | Go-интерфейсы пакета `game` | Перед Фазой 1 |
| WS-протокол | TypeScript типы + Go-структуры (общая JSON-схема) | Перед Фазой 2 |
| REST-схемы | OpenAPI / Go-типы (можно через `oapi-codegen`) | Перед Фазой 2 |
| Дизайн-токены | Tailwind config + CSS-переменные | Перед Фазой 4 |
| Zustand-схемы | TypeScript интерфейсы сторов | Перед Фазой 4 |

### Гейты между фазами

1. Все тесты фазы зелёные.
2. Линт без ошибок (`golangci-lint run`, `biome check`).
3. Code review через `superpowers:requesting-code-review`.
4. Проверка через `advisor()` перед закрытием фазы.
5. Коммит с тегом фазы (`phase-1-game-logic`, `phase-2-infra`, ...).

---

## Секция 13: Сводка пограничных случаев и рисков

| Категория | Случай | Митигация |
|---|---|---|
| **Сеть** | Потеря WS-соединения | Авто-reconnect (1s, 2s, 4s, max 30s), ресинхронизация через `game_state` |
| **Сеть** | Одновременные ходы (race) | Сервер — единственная истина, `move_count` для идемпотентности |
| **Игра** | Игрок закрыл вкладку посреди хода | 60-сек таймер → пропуск, 5-мин grace → автопоражение |
| **Игра** | Оба отключились | Комната в статусе `abandoned`, через 24ч очистка |
| **Игра** | Дубль 6-6 в начале без возможных ходов | Корректный пропуск, ход к сопернику |
| **Игра** | Все 15 в доме у обоих одновременно | Оба в `bearing_off`, кто первый выкинет — победил |
| **UI** | Тап мимо шашки на мобильном | 44×44px тач-зоны минимум, hit-area через CSS |
| **UI** | Поворот экрана во время игры | Адаптивная вёрстка, без потери состояния |
| **UI** | 15 шашек на одном пункте | Лимит видимого стека 5, остальные числом "+10" |
| **UX** | Игрок не понимает почему ход не проходит | Уведомление с причиной ("глухой забор", "нет хода", "обязан больший") |
| **Безопасность** | Подмена WS-сообщения | Серверная валидация, `session_token` + `move_count` |
| **Безопасность** | Подбор кода комнаты | 8 символов base32, rate-limit, после партии `finished` |
| **Производительность** | Лаги анимации на старых Android | Опция "минимальные анимации", `prefers-reduced-motion` |
| **Производительность** | Тормоза при 50+ комнатах | In-memory hub с шардированием по `room_id` |
| **Локализация** | Имена с кириллицей/эмодзи | UTF-8 валидация, санитизация, max 40 символов |
| **Браузеры** | Safari iOS не поддерживает что-то | Целимся в last 2 версии Safari, Chrome, Firefox, Samsung Internet |
| **PWA** | Добавление на экран | `manifest.json` + service worker (опционально в v2) |

---

## Следующие шаги

1. Утверждение этого design doc.
2. Создание детального implementation plan через `superpowers:writing-plans` — с конкретными задачами для каждого агента по фазам.
3. Настройка CI и установка инструментов из Секции 11.
4. Старт Фазы 0 — scaffolding проекта.
