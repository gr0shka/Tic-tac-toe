# Tic-Tac-Toe Game Service

RESTful API сервис для игры в «Крестики-нолики» с поддержкой двух режимов игры: против **непобедимого ИИ на базе алгоритма Minimax** (`player_vs_bot`) и **многопользовательского режима между реальными игроками** (`player_vs_player`).

Сервис спроектирован в соответствии с принципами **чистой архитектуры (Clean Architecture / DDD)**, использует **PostgreSQL** для персистентного хранения данных, **Uber Fx** для сборки графа зависимостей и **JWT (RSA256)** с ротацией Refresh-токенов для безопасной аутентификации.

---

## ✨ Ключевые возможности

- **Два режима игры**:
  - `player_vs_bot`: одиночная игра против алгоритма Minimax с штрафом за глубину дерева решений (Depth Penalization) для выбора кратчайшего победного пути.
  - `player_vs_player`: игра между двумя реальными игроками с созданием открытого лобби, списком ожидающих игр, подключением второго игрока и синхронизацией ходов.
- **Безопасность и JWT-аутентификация**:
  - Выпуск асимметричных JWT-токенов доступа (Access Token), подписанных закрытым RSA-ключом (**RS256**).
  - Хранение Refresh-токенов в PostgreSQL с поддержкой ротации токенов (`POST /user/token` и `POST /user/refresh`) и безопасной передачи через HttpOnly Cookie и JSON-тело.
  - Регистрация пользователей с хешированием паролей через `bcrypt` (cost 10).
  - Получение профиля текущего пользователя через `GET /user/me` и поиск по ID через `GET /user/{uuid}`.
- **Таблица лидеров (Лидерборд)**:
  - Эндпоинт `GET /leaderboard` с расчетом рейтинга на основе процента побед:
    $$\text{Ratio} = \frac{\text{Wins}}{\text{Wins} + \text{Loss} + \text{Draw}}$$
  - Учет партий только между реальными игроками (игры с ботами исключаются из рейтинга).
- **История завершенных игр**:
  - Эндпоинт `GET /games/ended` для просмотра всех завершенных партий текущего игрока с результатами и датой.
- **Персистентность и надежность в PostgreSQL**:
  - Драйвер `jackc/pgx/v5` (`pgxpool`), транзакционный контекстный менеджер (`Transactor`), пессимистические блокировки `SELECT ... FOR UPDATE` для защиты от race conditions при совершении ходов.
  - Версионирование схемы БД через миграции [Goose](https://github.com/pressly/goose), встроенные напрямую в бинарник через `//go:embed`.
- **CORS Middleware с белым списком Origin**:
  - Поддержка локальных клиентов (`http://localhost:3000`, `http://localhost:63342` для JetBrains IDE, `http://localhost:8080`) с флагом `Access-Control-Allow-Credentials: true`.
- **Строгая серверная валидация**:
  - Проверка очередности ходов, запрет перезаписи клеток, защита от множественных ходов за один раунд.
- **Комплексное тестирование**:
  - Модульные тесты доменных правил и мапперов.
  - Table-Driven интеграционные тесты репозиториев и usecase-сценариев на реальной базе PostgreSQL через [testcontainers-go](https://github.com/testcontainers/testcontainers-go).
- **Интерактивный веб-клиент**:
  - Автономный SPA-клиент [`tic-tac-toe.html`](tic-tac-toe.html) с 4 разделами (Вход и JWT, Игра, Лобби и История, Лидерборд) для ручного тестирования API.

---

## 🏛️ Архитектура проекта

Структура репозитория организована по принципам Clean Architecture:

```text
├── cmd/
│   └── main.go                         # Точка входа в сервис (fx.New().Run())
├── internal/
│   ├── config/                         # Конфигурация приложения, .env и параметры JWT
│   ├── di/                             # Сборка графа зависимостей (Uber Fx) и запуск HTTP-сервера
│   ├── domain/                         # Слой бизнес-сущностей и доменных правил (чистый Go)
│   │   ├── auth/                       # Сущность Token (Refresh Token)
│   │   ├── game/                       # Сущности GameBoard, CurrentGame, Player, LeaderBoard
│   │   │   └── rules/                  # Доменные сервисы: Minimax (бот), валидация ходов
│   │   └── user/                       # Сущность User и доменные валидации
│   ├── migration/                      # Применение миграций схемы БД через Goose
│   ├── repository/                     # Инфраструктурный слой доступа к данным
│   │   ├── postgres/                   # Реализации репозиториев на базе PostgreSQL (pgx/v5)
│   │   │   ├── game/                   # Репозиторий игр, расчет лидерборда и DTO
│   │   │   ├── token/                  # Репозиторий JWT Refresh-токенов
│   │   │   └── user/                   # Репозиторий пользователей (ON CONFLICT DO NOTHING)
│   │   ├── testutil/                   # Инициализация тестовой БД через Testcontainers
│   │   └── transactor/                 # Контекстный транзактор и управление транзакциями
│   ├── security/                       # Безопасность и криптография
│   │   └── jwt/                        # RSA JWT провайдер (генерация, валидация токенов)
│   ├── transport/                      # Слой доставки (Delivery)
│   │   └── http/                       # HTTP-хэндлеры, DTO, CORS и JWT middleware
│   │       ├── dto/                    # Запросы и ответы HTTP API
│   │       ├── handler/                # Обработчики (игры, пользователи, авторизация)
│   │       ├── mapper/                 # Преобразование доменных моделей в HTTP DTO
│   │       └── middleware/             # Middleware (CORS, RequestLogger, JWT Authenticator, Recovery)
│   └── usecase/                        # Слой прикладных сценариев (Use Cases)
│       ├── auth/                       # AuthenticateService: регистрация, аутентификация, ротация токенов
│       ├── game/                       # GameService: создание игр, ходы, join, история, лидерборд
│       └── user/                       # UserService: получение пользователей
├── migration/                          # SQL-миграции базы данных (Goose)
├── secrets/                            # RSA-ключи для подписи и проверки JWT (jwt_private.pem, jwt_public.pem)
├── Dockerfile                          # Мультистейдж сборка Docker (минимальный автономный образ Alpine)
├── docker-compose.yml                  # Сервисы приложения, PostgreSQL и Docker Secrets
├── Makefile                            # Команды генерации RSA-ключей, миграций, тестов и сборки
├── tic-tac-toe.html                    # Интерактивный SPA веб-клиент для тестирования
├── .env                                # Локальные переменные окружения
└── .env.scheme                         # Шаблон переменных окружения
```

---

## 🚀 Быстрый старт

### 1. Генерация RSA-ключей для JWT

Для подписи и проверки JWT-токенов требуются открытый и закрытый RSA-ключи. Сгенерируйте их командой:

```bash
make generate-secrets
```
Команда создаст файлы `secrets/jwt_private.pem` и `secrets/jwt_public.pem`.

---

### 2. Настройка переменных окружения

Создайте файл `.env` на основе шаблона [`.env.scheme`](.env.scheme):

```bash
cp .env.scheme .env
```

Пример содержимого `.env`:
```env
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_DATABASE=tictactoe

MIGRATION=true

APP_HOST=0.0.0.0
APP_PORT=8080

JWT_PUBLIC_FILE=secrets/jwt_public.pem
JWT_PRIVATE_FILE=secrets/jwt_private.pem
JWT_ACCESS_TIME=15m
JWT_REFRESH_TIME=720h
```

---

### 3. Запуск через Docker Compose (Рекомендуется)

Запуск сервиса вместе с PostgreSQL и автоматической проброской секретов:

```bash
docker compose up --build
```

Сервер будет доступен по адресу `http://localhost:8080`.

---

### 4. Локальный запуск (без Docker)

1. Убедитесь, что PostgreSQL запущен и база данных создана.
2. Примените миграции с помощью [Goose](https://github.com/pressly/goose):
   ```bash
   make migrate-up
   ```
3. Запустите сервис:
   ```bash
   go run cmd/main.go
   ```

---

### 5. Запуск тестов

Запуск всех модульных и интеграционных тестов (с автоматическим поднятием PostgreSQL в Docker через testcontainers):

```bash
make test
# или напрямую:
go test -v -race ./...
```

Генерация отчёта о покрытии тестами:
```bash
make test-cover-html
```

---

## 📡 Спецификация REST API

Все запросы и ответы используют заголовок `Content-Type: application/json`.  
Для всех защищённых эндпоинтов требуется передача заголовка:  
`Authorization: Bearer <access_token>`.

---

### 🔐 Аутентификация и пользователи

#### 1. Регистрация нового пользователя
* **Метод**: `POST /user/register`
* **Авторизация**: Не требуется
* **Тело запроса**:
  ```json
  {
    "login": "player1",
    "password": "secretpassword"
  }
  ```
* **Ответ**: `201 Created`
  ```json
  {
    "message": "successfully register"
  }
  ```

#### 2. Вход в систему (Аутентификация)
* **Метод**: `POST /user/auth`
* **Авторизация**: Не требуется
* **Тело запроса**:
  ```json
  {
    "login": "player1",
    "password": "secretpassword"
  }
  ```
* **Ответ**: `200 OK` (+ устанавливает HttpOnly Cookie `refresh_token`)
  ```json
  {
    "type": "Bearer",
    "accessToken": "eyJhbGciOiJSUzI1NiIs...",
    "refreshToken": "6a9f3958-45fa-4c8d-8a58-0051e89b4344"
  }
  ```

#### 3. Обновление токена доступа (Access Token)
* **Метод**: `POST /user/token`
* **Авторизация**: Не требуется (Refresh-токен передаётся через Cookie `refresh_token`, JSON-тело или заголовок `Authorization: Bearer <refresh_token>`)
* **Тело запроса (опционально)**:
  ```json
  {
    "refreshToken": "6a9f3958-45fa-4c8d-8a58-0051e89b4344"
  }
  ```
* **Ответ**: `200 OK`
  ```json
  {
    "type": "Bearer",
    "accessToken": "eyJhbGciOiJSUzI1NiIs...",
    "refreshToken": "6a9f3958-45fa-4c8d-8a58-0051e89b4344"
  }
  ```

#### 4. Ротация пары токенов (Refresh & Access)
* **Метод**: `POST /user/refresh`
* **Авторизация**: Не требуется (Refresh-токен передаётся через Cookie или тело)
* **Тело запроса (опционально)**:
  ```json
  {
    "refreshToken": "6a9f3958-45fa-4c8d-8a58-0051e89b4344"
  }
  ```
* **Ответ**: `200 OK` (выдаётся новая пара Access и Refresh токенов, старый аннулируется)

#### 5. Профиль текущего пользователя («Кто я?»)
* **Метод**: `GET /user/me`
* **Авторизация**: `Authorization: Bearer <access_token>`
* **Ответ**: `200 OK`
  ```json
  {
    "id": "e7b0e27a-8b1e-4c7e-97c2-550998f8045f",
    "login": "player1"
  }
  ```

#### 6. Получение информации о пользователе по UUID
* **Метод**: `GET /user/{uuid}`
* **Авторизация**: `Authorization: Bearer <access_token>`
* **Ответ**: `200 OK`
  ```json
  {
    "id": "e7b0e27a-8b1e-4c7e-97c2-550998f8045f",
    "login": "player1"
  }
  ```
* **Ошибки**: `404 Not Found`, если пользователь не найден.

---

### 🎮 Игровой процесс

#### 1. Создание новой игры
* **Метод**: `POST /games`
* **Авторизация**: `Authorization: Bearer <access_token>`
* **Тело запроса**:
  ```json
  {
    "mode": "player_vs_bot"
  }
  ```
  *(Допустимые значения `mode`: `"player_vs_bot"`, `"player_vs_player"`)*
* **Ответ**: `201 Created`
  ```json
  {
    "id": "a988d8b9-8e89-4bc2-9e8d-d790757a3e81",
    "status": "player_turn",
    "activePlayer": "e7b0e27a-8b1e-4c7e-97c2-550998f8045f",
    "player1ID": "e7b0e27a-8b1e-4c7e-97c2-550998f8045f",
    "player1Symbol": 0,
    "player2ID": "3196f7e4-23db-4e1b-bb3b-74895cae7f12",
    "player2Symbol": 1,
    "winner": -1,
    "winnerID": null,
    "board": [
      [-1, -1, -1],
      [-1, -1, -1],
      [-1, -1, -1]
    ],
    "createdAt": "2026-10-25T20:00:00Z"
  }
  ```

#### 2. Получение текущего состояния игры
* **Метод**: `GET /games/{uuid}`
* **Авторизация**: `Authorization: Bearer <access_token>`
* **Ответ**: `200 OK` (объект `GameBoardResponse`).

#### 3. Совершение хода
* **Метод**: `POST /games/{uuid}/moves`
* **Авторизация**: `Authorization: Bearer <access_token>`
* **Тело запроса**:
  ```json
  {
    "board": [
      [0, -1, -1],
      [-1, -1, -1],
      [-1, -1, -1]
    ]
  }
  ```
* **Ответ**: `200 OK`  
  - В режиме `player_vs_bot`: сервер валидирует ход игрока, моментально рассчитывает и применяет ответный ход Minimax-бота и возвращает итоговое состояние доски с ходами обоих участников.
  - В режиме `player_vs_player`: проверяет корректность хода, фиксирует его в БД и переключает активного игрока.
* **Ошибки**: `400 Bad Request` при нарушении правил (ход не в свою очередь, изменение ранее занятых клеток, несколько ходов одновременно).

#### 4. Присоединение к открытой PvP-игре
* **Метод**: `POST /games/{uuid}/join`
* **Авторизация**: `Authorization: Bearer <access_token>`
* **Ответ**: `200 OK` (второй игрок регистрируется со статусом `player_turn`).

#### 5. Список открытых игр (Лобби)
* **Метод**: `GET /games`
* **Авторизация**: `Authorization: Bearer <access_token>`
* **Ответ**: `200 OK` (массив открытых игр, ожидающих второго игрока: `player2ID IS NULL`).

#### 6. История сыгранных партий пользователя
* **Метод**: `GET /games/ended`
* **Авторизация**: `Authorization: Bearer <access_token>`
* **Ответ**: `200 OK` (массив завершённых игр текущего авторизованного пользователя со статусами `player_wins` или `draw`, отсортированных по дате убывания).

---

### 🏆 Таблица лидеров (Лидерборд)

#### 1. Получение рейтинга игроков
* **Метод**: `GET /leaderboard?limit=10`
* **Авторизация**: `Authorization: Bearer <access_token>`
* **Параметры строки запроса**: `limit` (число от 1 до 100, по умолчанию 10)
* **Ответ**: `200 OK`
  ```json
  {
    "players": [
      {
        "id": "e7b0e27a-8b1e-4c7e-97c2-550998f8045f",
        "login": "alice",
        "wins": 5,
        "loss": 1,
        "draw": 1,
        "ratio": 0.714
      },
      {
        "id": "3196f7e4-23db-4e1b-bb3b-74895cae7f12",
        "login": "bob",
        "wins": 3,
        "loss": 3,
        "draw": 0,
        "ratio": 0.5
      }
    ]
  }
  ```

---

### 📋 Обозначения и состояния

| Значение клетки | Описание |
| :---: | :--- |
| `-1` | Пустая клетка |
| `0`  | Первый игрок (Крестик `X`) |
| `1`  | Второй игрок / Бот (Нолик `O`) |

| Статус игры (`status`) | Описание |
| :--- | :--- |
| `waiting_for_players` | Ожидание подключения второго игрока |
| `player_turn` | Игра активна, ожидается ход активного игрока |
| `player_wins` | Игра завершена победой одного из участников |
| `draw` | Игра завершена вничью |

| Победитель (`winner`) | Описание |
| :---: | :--- |
| `-1` | Победителя нет (игра продолжается или ничья) |
| `0`  | Победил первый игрок (`X`) |
| `1`  | Победил второй игрок / бот (`O`) |

---

## 🕹️ Интерактивный веб-клиент

В корень репозитория включён удобный клиент [`tic-tac-toe.html`](tic-tac-toe.html) (чистый HTML/CSS/JS без внешних зависимостей):

1. Запустите бэкенд (`docker compose up` или `go run cmd/main.go`).
2. Откройте [`tic-tac-toe.html`](tic-tac-toe.html) в браузере (или через веб-сервер IDE, порт `63342` разрешён в CORS).
3. **Разделы интерфейса**:
   - **1. Вход и JWT**: Регистрация, вход с получением JWT, быстрые пресеты (`player1`/`player2`), ручное тестирование обновления токенов (`/user/token`, `/user/refresh`), вызов `GET /user/me` и поиск по UUID.
   - **2. Игра**: Создание игры против бота или реального игрока, интерактивное поле 3x3, автообновление каждые 2 секунды, отображение текущего хода и результата.
   - **3. Лобби и История**: Список открытых игр с кнопкой «Присоединиться», прямое подключение по UUID, а также просмотр завершённых партий текущего пользователя с кнопкой загрузки на поле.
   - **4. Лидерборд**: Таблица топа игроков с медалями (🥇, 🥈, 🥉), процентом побед и подсветкой строки текущего пользователя.
   - **Инспектор API (внизу)**: Детальный лог последнего HTTP-запроса, статус-код и форматированный JSON-ответ сервера.

---

## 🛠️ Стек технологий

- **Язык**: [Go](https://go.dev/) (1.26+)
- **База данных**: [PostgreSQL 16](https://www.postgresql.org/)
- **Драйвер БД**: [jackc/pgx/v5](https://github.com/jackc/pgx) (`pgxpool`)
- **Миграции**: [Goose](https://github.com/pressly/goose) (автономно встроены в бинарник через `//go:embed`)
- **DI-фреймворк**: [Uber Fx](https://github.com/uber-go/fx) (граф зависимостей и Graceful Shutdown)
- **Аутентификация и безопасность**:
  - [golang-jwt/jwt/v5](https://github.com/golang-jwt/jwt) (RSA256 токены)
  - `golang.org/x/crypto/bcrypt` (хеширование паролей)
  - Хранение и ротация Refresh-токенов в PostgreSQL
- **Тестирование**:
  - Стандартный `testing` в стиле Table-Driven Tests
  - [testcontainers-go](https://github.com/testcontainers/testcontainers-go) для запуска PostgreSQL в Docker при интеграционных тестах
- **Логирование**: Стандартный структурированный `log/slog`
- **Конкурентность и транзакции**: Контекстный `Transactor` (`pgx.Tx`), блокировки `SELECT ... FOR UPDATE` и защита от гонок `ON CONFLICT DO NOTHING`
- **Идентификаторы**: [google/uuid](https://github.com/google/uuid) (UUIDv4)
- **Контейнеризация**: Docker (multi-stage non-root build), Docker Compose + Docker Secrets
