# Сервер для хранения цен криптовалют

Необходимо реализовать HTTP REST API сервер (на порту 8080) для хранения и получения цен криптовалют с системой аутентификации на любом языке программирования.

## API

### Аутентификация
- **POST /auth/register** - регистрация нового пользователя
  - Body: `{"username": "string", "password": "string"}`
  - Success (201): `{"token": "string"}`
  - Error (400/409): `{"error": "string"}`

- **POST /auth/login** - вход пользователя
  - Body: `{"username": "string", "password": "string"}`
  - Success (200): `{"token": "string"}`
  - Error (400/401): `{"error": "string"}`

### CRUD операции для криптовалют
Все операции требуют аутентификации через заголовок `Authorization: Bearer <token>`

- **GET /crypto** - получить список всех криптовалют
  - Response: `{"cryptos": [{"symbol": "BTC", "name": "Bitcoin", "current_price": 45000.50, "last_updated": "2024-01-01T12:00:00Z"}]}`

- **POST /crypto** - добавить новую криптовалюту для отслеживания
  - Body: `{"symbol": "BTC"}`
  - Success (201): `{"crypto": {...}}`
  - Error (400/409/500): `{"error": "string"}`

- **GET /crypto/{symbol}** - получить информацию о конкретной криптовалюте
  - Success (200): `{"symbol": "BTC", "name": "Bitcoin", "current_price": 45000.50, "last_updated": "2024-01-01T12:00:00Z"}`
  - Error (404): `{"error": "string"}`

- **PUT /crypto/{symbol}/refresh** - принудительно обновить цену криптовалюты
  - Success (200): `{"crypto": {...}}`
  - Error (404/500): `{"error": "string"}`

- **GET /crypto/{symbol}/history** - получить историю цен криптовалюты
  - Response: `{"symbol": "BTC", "history": [{"price": 45000.50, "timestamp": "2024-01-01T12:00:00Z"}]}`

- **GET /crypto/{symbol}/stats** - получить статистику по ценам криптовалюты
  - Response: `{"symbol": "BTC", "current_price": 45000.50, "stats": {"min_price": 44000, "max_price": 46000, "avg_price": 45000, "price_change": 1000, "price_change_percent": 2.27, "records_count": 100}}`

- **DELETE /crypto/{symbol}** - удалить криптовалюту из отслеживания (включая историю)
  - Success (200): `{}` (пустой объект)
  - Error (404): `{"error": "string"}`

## Дополнительная часть ДЗ, расписание автоматического обновления:

- **GET /schedule** - получить текущие настройки автообновления
  - Response: `{"enabled": true, "interval_seconds": 30, "last_update": "2024-01-01T12:00:00Z", "next_update": "2024-01-01T12:00:30Z"}`

- **PUT /schedule** - изменить настройки автообновления
  - Request: `{"enabled": true, "interval_seconds": 60}` (минимум 10 секунд, максимум 3600 секунд)
  - Success (200): `{"enabled": true, "interval_seconds": 60}`
  - Error (400/500): `{"error": "string"}`

- **POST /schedule/trigger** - принудительно запустить обновление всех цен
  - Success (200): `{"updated_count": 5, "timestamp": "2024-01-01T12:00:00Z"}`
  - Error (500): `{"error": "string"}`

## Детали реализации

- Все данные должны храниться в оперативной памяти (будем заменять на БД попозже, поэтому заранее пишите в стиле чистой архитектуры)
- Каждая криптовалюта хранит историю из максимум 100 последних цен
- Цены обновляются каждые 30 секунд в фоновом потоке
- Пароли должны быть хешированы (bcrypt, scrypt)
- Используйте JWT токены для аунтефикации

### CoinGecko API

Для получения всей информации по криптовалютам (включая цены) будет использоваться [CoinGecko API](https://docs.coingecko.com/reference/introduction):

CoinGecko использует уникальные ID вместо символов для идентификации криптовалют:
- Symbol: `BTC` → ID: `bitcoin`
- Symbol: `ETH` → ID: `ethereum`
- Symbol: `DOGE` → ID: `dogecoin`

На наш сервер криптовалюта при добавлении будет приходить в виде тикера (Symbol в маппинге), поэтому вам нужно умень этот маппинг запрашивать и кешировать (`/coins/list` и `/search` endpoints). Остальные методы ищите в документации ;)

## Запуск тестов

### Основные тесты (обязательная часть)
```bash
make test
```

```bash
make test SCHEDULE=1
```

Ваше решение должно содержать файл с сервером: `cryptoserver.{ext}` (`cryptoserver.py`, `cryptoserver.go` и т.д.)
