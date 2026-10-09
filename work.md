# План доработок и технический долг: Tic-Tac-Toe

Документ содержит полный перечень несоответствий требованиям ТЗ ([`task.md`](task.md)), нарушений практик Go и недостатков, препятствующих эксплуатации сервиса в Production-среде, с практическими рекомендациями по их исправлению.

---

## Содержание

1. [Блок 1: Доработки для 100% соответствия ТЗ (task.md)](#1-доработки-для-100-соответствия-тз-taskmd)
2. [Блок 2: Доработки до уровня Production-Ready](#2-доработки-до-уровня-production-ready)
3. [Блок 3: Соответствие конвенциям Go (Idiomatic Go)](#3-соответствие-конвенциям-go-idiomatic-go)
4. [Блок 4: Улучшение покрытия тестами](#4-улучшение-покрытия-тестами)
5. [Инженерный чек-лист задач](#инженерный-чек-лист-задач)

---

## 1. Доработки для 100% соответствия ТЗ (task.md)

### 1.1. Выделение отдельного сервиса аутентификации (`AuthenticationService`)
* **Проблема:**  
  В ТЗ (Task 2) явно указано:  
  > *«Create an authentication service that uses UserService with the following methods:*  
  > *- A registration method that takes a SignUpRequest and returns an indicator of successful registration.*  
  > *- An authentication method that receives login and password in the header encoded as base64(login:password) and returns the user’s UUID.»*  
  В текущей реализации отдельный сервис не создан: логика находится напрямую внутри `UserService`, а декодирование Base64 размазано по `UserHandler` и `UserAuthenticator`.
* **Что исправить:**  
  1. Создать интерфейс и структуру `AuthService` (например, в `internal/usecase/auth/auth.go`).
  2. Внедрить в него зависимость от `user.UserService`.
  3. Реализовать метод регистрации, принимающий `dto.SignUpRequest` и возвращающий `(bool, error)` или `error`.
  4. Реализовать метод аутентификации, принимающий заголовок `Authorization` (или пару login/pass из заголовка), валидирующий его и возвращающий `uuid.UUID`.
  5. Подключить `AuthService` в граф зависимостей Fx (`internal/di/register.go`).

### 1.2. Статусы игры и UUID победителя в API
* **Проблема:**  
  В ТЗ (Task 3) определены состояния текущей игры:  
  > *- Waiting for players*  
  > *- Player with UUID to move*  
  > *- Draw*  
  > *- Player with UUID wins*  
  В текущей реализации:
  - Статусы заданы статическими константами (`player_turn`, `player_wins`, `draw`, `waiting_for_players`).
  - Поле победителя в DTO и БД — это целое число `winner: int` (`0`, `1` или `-1`), а не UUID игрока.
  - В ответе `GameBoardResponse` (`internal/transport/http/dto/game.go`) отсутствуют `player1_id` и `player2_id`. Клиент видит `winner: 0`, но не имеет возможности понять, какой именно UUID выиграл.
* **Что исправить:**  
  1. Добавить в `GameBoardResponse` идентификаторы игроков: `Player1ID uuid.UUID` и `Player2ID *uuid.UUID`.
  2. Заменить поле победителя в API на `WinnerID *uuid.UUID` (или маппить `0` и `1` в соответствующий UUID игрока при сериализации).
  3. Сформировать понятный статус (либо включить UUID в строку статуса / поле `StatusDescription`, либо обогатить модель полями `ActivePlayerID *uuid.UUID` и `WinnerID *uuid.UUID`, документировав схему).

---

## 2. Доработки до уровня Production-Ready

### 2.1. Устранение Race Condition в логике игры (Concurrency Control)
* **Проблема:**  
  В `JoinGame` и `ProcessPlayerMove` (`internal/usecase/app/app_impl.go`) используется антипаттерн `Read-Modify-Write` без транзакций и блокировок:
  ```go
  cg, err := a.repository.Get(ctx, gameID)     // Чтение
  // проверка и модификация данных в памяти
  err = a.repository.Update(ctx, cg)            // Запись
  ```
  - Если два игрока одновременно вызовут `POST /games/{uuid}/join`, оба прочитают `player2_id == nil`, оба успешно пройдут проверки и оба выполнят `UPDATE`. Один молча перезапишет другого.
  - Если игрок пошлёт два параллельных хода, состояние доски и счётчик ходов могут рассинхронизироваться.
* **Что исправить:**  
  1. Для `JoinGame`: использовать атомарный запрос с условием (Optimistic Locking):
     ```sql
     UPDATE current_game
     SET player2_id = $2, player2_real = $3, player2_symbol = $4, status = $5, active_player_id = $6
     WHERE id = $1 AND player2_id IS NULL;
     ```
     Проверять `commandTag.RowsAffected()`. Если 0 — возвращать ошибку, что игра уже занята.
  2. Для `ProcessPlayerMove`: обновлять доску только с проверкой версии/номера хода:
     ```sql
     UPDATE current_game
     SET board = $2, number_of_turn = $3, status = $4, winner = $5, active_player_id = $6
     WHERE id = $1 AND number_of_turn = $expectedPreviousTurn;
     ```
     Либо обернуть метод в транзакцию с `SELECT ... WHERE id = $1 FOR UPDATE`.

### 2.2. Автоматическое применение миграций при развёртывании
* **Проблема:**  
  В `README.md` рекомендуется запуск через `docker compose up --build`. При этом ни в `docker-compose.yml`, ни в `cmd/main.go`, ни в `Dockerfile` миграции `goose` **не применяются**. При первом запуске в пустой базе таблицы отсутствуют, и все запросы завершаются ошибкой `500/400`.
* **Что исправить:**  
  1. **Вариант А (предпочтительный для микросервисов):** добавить хук применения миграций при старте в `internal/di/register.go`:
     ```go
     lc.Append(fx.Hook{
         OnStart: func(ctx context.Context) error {
             db := stdlib.OpenDBFromPool(pool)
             defer db.Close()
             return goose.Up(db, "migration")
         },
     })
     ```
  2. **Вариант Б:** добавить отдельный сервис `migration` в `docker-compose.yml` на базе образа `pressly/goose`.

### 2.3. Корректный Graceful Shutdown для пула БД
* **Проблема:**  
  В `RegisterServer` (`internal/di/register.go`) хук `OnStop` зарегистрирован только для `http.Server`. Пул соединений `*pgxpool.Pool` никогда не закрывается через `pool.Close()`, что приводит к обрыву соединений с PostgreSQL при выключении пода/контейнера.
* **Что исправить:**  
  Зарегистрировать хук остановки для пула:
  ```go
  lc.Append(fx.Hook{
      OnStop: func(ctx context.Context) error {
          pool.Close()
          return nil
      },
  })
  ```

### 2.4. Вынос порта и параметров пула в конфигурацию (12-Factor App)
* **Проблема:**  
  В `internal/di/register.go:78` адрес сервера захардкожен: `Addr: ":8080"`. Переменная `APP_PORT` из `.env` и `docker-compose.yml` не используется кодом. В PaaS/K8s с динамическим портом сервис не сможет запуститься.
* **Что исправить:**  
  1. Добавить в `config.Config` секцию `Server` с полем `Port` (читать из `PORT` или `APP_PORT`, по умолчанию `8080`).
  2. Добавить параметры пула соединений БД: `MaxConns`, `MinConns`, `MaxConnIdleTime`, `MaxConnLifetime`.
  3. Передавать порт в конструктор сервера.

### 2.5. Оптимизация схемы БД и запросов
* **Проблема:**  
  1. Запрос `SELECT * FROM current_game WHERE player2_id IS NULL` (`AllGames`) делает `Seq Scan` (полный перебор таблицы). При тысячах игр в базе это создаст сильную нагрузку на диск и CPU.
  2. Нет пагинации (`LIMIT` / `OFFSET`) в ручке `GET /games`: сервис вернёт все открытые игры разом в одном JSON.
  3. Отсутствуют колонки аудита: `created_at TIMESTAMP WITH TIME ZONE DEFAULT now()`, `updated_at`.
  4. Отсутствуют внешние ключи (Foreign Keys) между `current_game` и `users`.
* **Что исправить:**  
  1. Добавить индекс в миграцию: `CREATE INDEX idx_current_game_waiting ON current_game(id) WHERE player2_id IS NULL;` (частичный индекс).
  2. Добавить `LIMIT` и `OFFSET` (с разумным дефолтом, например `LIMIT 50`) в `AllGames`.
  3. Добавить поля `created_at` и `updated_at`.

### 2.6. Безопасность (Security)
* **Проблема:**  
  1. **Утечка поля пароля в API:** в `dto.UserInfoResponse` присутствует поле `Password string json:"password"`. При вызове `GET /user/{uuid}` клиенту возвращается `"password": ""`. Поле пароля не должно присутствовать в публичной модели ответа.
  2. **Слабая валидация:** `validateData` проверяет только непустую строку. Допускаются пароли из одного символа или пробела. Не контролируется максимальная длина (у `bcrypt` ограничение 72 байта).
  3. **Отсутствие Rate Limiting:** эндпоинты `/user/auth` и `/user/register` подвержены атакам перебора паролей (Brute-force) и флуду создания пользователей.
  4. **RFC 7617 Basic Auth compliance:** при ошибке авторизации не выставляется заголовок `WWW-Authenticate: Basic realm="Tic-Tac-Toe"`.
* **Что исправить:**  
  1. Удалить поле `Password` из `dto.UserInfoResponse`.
  2. Ввести правила валидации: логин от 3 до 32 символов (без пробелов), пароль от 8 до 72 символов.
  3. Добавить middleware ограничения частоты запросов (Rate Limiting, например, `golang.org/x/time/rate`).
  4. Добавить заголовок `w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)` при возврате `401`.

### 2.7. Корректные HTTP-статусы и обработка ошибок
* **Проблема:**  
  - При регистрации существующего пользователя возвращается `400 Bad Request` вместо `409 Conflict`.
  - В `POST /user/register` тело ответа — сырой текст `"User registered"` вместо JSON, что ломает контракт клиентов (в README обещано `application/json` на всех ручках).
  - Ошибки базы в `AllGames` и `GetGame` возвращают `404 Not Found` даже при сетевом сбое или падении PostgreSQL.
  - Сырые системные ошибки возвращаются клиенту через `http.Error(w, err.Error(), ...)`, раскрывая внутреннее устройство БД.
* **Что исправить:**  
  1. Возвращать `409 Conflict` для `user.ErrUserAlreadyExists`.
  2. Возвращать JSON-ответ при регистрации: `{"message": "user registered", "status": "created"}`.
  3. Чётко разделять ошибки: `errors.Is(err, pgx.ErrNoRows)` ➔ `404 Not Found`, любые другие ошибки БД ➔ `500 Internal Server Error`.
  4. Реализовать единый хелпер/структуру для ошибок API:
     ```json
     { "error": "user-friendly message", "code": "RESOURCE_NOT_FOUND" }
     ```

### 2.8. Наблюдаемость (Observability) и Healthchecks
* **Проблема:**  
  - Используется устаревший стандартный логгер `log.Print`. Отсутствует контекстное структурированное логирование.
  - Нет middleware для логов HTTP-запросов (метод, URL, статус, время выполнения, User-Agent, IP).
  - Отсутствуют эндпоинты `/healthz` (Liveness) и `/ready` (Readiness с проверкой `pool.Ping(ctx)`).
  - Нет экспорта метрик для Prometheus.
* **Что исправить:**  
  1. Перейти на `log/slog` (стандарт Go 1.21+) или `uber-go/zap`.
  2. Добавить middleware логирования запросов.
  3. Реализовать хэндлер `/healthz` и `/ready` с проверкой доступности БД.

---

## 3. Соответствие конвенциям Go (Idiomatic Go)

### 3.1. Типизированные ключи контекста
* **Проблема:**  
  В `internal/transport/http/middleware/user.go:10` используется:
  ```go
  const UserIDContextName = "UserID"
  r = r.WithContext(context.WithValue(r.Context(), UserIDContextName, u.ID().String()))
  ```
  Это прямо нарушает рекомендации стандартной библиотеки Go: базовые типы (`string`) не должны использоваться как ключи контекста из-за риска коллизий между пакетами.
* **Что исправить:**  
  Объявить неэкспортируемый тип и хелперы доступа:
  ```go
  type ctxKeyUserID struct{}

  func ContextWithUserID(ctx context.Context, id uuid.UUID) context.Context {
      return context.WithValue(ctx, ctxKeyUserID{}, id)
  }

  func UserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
      id, ok := ctx.Value(ctxKeyUserID{}).(uuid.UUID)
      return id, ok
  }
  ```

### 3.2. Хранение типизированного `uuid.UUID` в контексте
* **Проблема:**  
  В middleware значение приводится к `string`, а затем в каждом хэндлере повторно вызывается `uuid.Parse(userStrID)`.  
* **Что исправить:**  
  Класть в контекст сразу `uuid.UUID` (или `*user.User`), исключив постоянный лишний парсинг.

### 3.3. Единообразие ресиверов методов
* **Проблема:**  
  В `internal/usecase/app/app_impl.go` методы имеют value receiver: `func (a appService) JoinGame(...)`, а в `internal/usecase/user/user_impl.go` — pointer receiver: `func (s *userService) Register(...)`.
* **Что исправить:**  
  Использовать pointer receiver `(a *appService)` во всех сервисах с зависимостями.

### 3.4. Оборачивание ошибок через `%w`
* **Проблема:**  
  В репозиториях и сервисах ошибки возвращаются "как есть" или подменяются доменными без сохранения стека вызовов.
* **Что исправить:**  
  Добавить контекст к ошибкам через `fmt.Errorf("failed to save user: %w", err)`.

### 3.5. Унификация REST путей
* **Проблема:**  
  Смешение множественного и единственного числа:
  - `POST /games`, `GET /games`, `POST /games/{uuid}/join` (множественное число).
  - `POST /game/{uuid}` (единственное число для хода).
* **Что исправить:**  
  Перевести все игровые эндпоинты на префикс `/games`: `POST /games/{uuid}/moves` или `POST /games/{uuid}/turn`.

---

## 4. Улучшение покрытия тестами

### 4.1. Добавление Happy Path для регистрации пользователя
* **Проблема:**  
  В [`internal/usecase/user/user_test.go`](internal/usecase/user/user_test.go#L104-L160) в `TestUserService_Register` протестированы исключительно ошибочные кейсы (`invalid login`, `invalid password`, `already exists`). Успешный сценарий регистрации с проверкой вызова `repo.Save` не протестирован.
* **Что исправить:**  
  Добавить позитивный тест-кейс в таблицу `TestUserService_Register`.

### 4.2. Concurrency-тесты
* **Проблема:**  
  Отсутствуют тесты на одновременные запросы к одной игре (параллельный `JoinGame` от двух пользователей, параллельные ходы).
* **Что исправить:**  
  Написать интеграционный тест с запуском нескольких горутин (`sync.WaitGroup`) на присоединение к одной игре для валидации защиты от гонок.

---

## Инженерный чек-лист задач

### Соответствие ТЗ
- [+] Выделить `AuthenticationService` (поверх `UserService`) для регистрации через `SignUpRequest` и аутентификации через Base64 header.
- [+] Добавить `player1_id` и `player2_id` в `GameBoardResponse`.
- [+] Изменить `Winner` на `WinnerID *uuid.UUID` в DTO и логике завершения игры.

### Production-Ready
- [+] Защитить `JoinGame` и `ProcessPlayerMove` от гонок данных (атомарный `UPDATE` с условием или `FOR UPDATE`).
- [+] Настроить автомиграции БД при старте приложения через Fx Lifecycle или init-контейнер Docker.
- [+] Добавить хук `pool.Close()` при завершении приложения в Fx.
- [+] Вынести порт HTTP-сервера и параметры пула БД в `internal/config`.
- [ ] Добавить частичный индекс для поиска открытых игр (`player2_id IS NULL`).
- [ ] Добавить пагинацию (`limit`, `offset`) в `GET /games`.
- [ ] Удалить поле `Password` из структуры `UserInfoResponse`.
- [ ] Усилить валидацию логина и пароля (длина, спецсимволы, ограничение 72 байт).
- [ ] Исправить HTTP-статусы: `409 Conflict` при дубликате логина, `500` при системных ошибках БД.
- [ ] Сделать ответ регистрации валидным JSON (`Content-Type: application/json`).
- [ ] Внедрить структурированный логгер `log/slog` и HTTP logging middleware.
- [ ] Добавить healthcheck ручки `/healthz` и `/ready`.

### Go-конвенции и тесты
- [ ] Заменить строковый ключ `UserIDContextName` на типизированную структуру в `middleware/user.go`.
- [ ] Передавать в контекст `uuid.UUID` напрямую без конвертации в `string`.
- [ ] Привести все методы сервисов к pointer receivers `(s *appService)`.
- [ ] Оборачивать ошибки с помощью `fmt.Errorf("...: %w", err)`.
- [ ] Унифицировать маршрут хода: заменить `/game/{uuid}` на `/games/{uuid}/moves`.
- [ ] Добавить тест успешной регистрации в `internal/usecase/user/user_test.go`.
