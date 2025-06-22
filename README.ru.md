# Go Active Record

Полнофункциональный Active Record ORM для Go, вдохновленный Rails Active Record. Предоставляет комплексный интерфейс для работы с базами данных с продвинутыми возможностями, такими как хуки, пакетные операции, резолвер баз данных, построитель запросов и многое другое.

## 🚀 Возможности

### Основные функции
- **CRUD операции** - Создание, Чтение, Обновление, Удаление записей с автоматическим управлением ID
- **Система хуков** - Хуки Before/After для операций Create, Update, Delete, Save, Find
- **Поддержка транзакций** - Полное управление транзакциями с поддержкой контекста
- **Построитель запросов** - Плавный интерфейс для создания сложных запросов с режимом dry run
- **Пакетные операции** - Эффективная пакетная вставка, поиск пакетами и массовые операции
- **Резолвер баз данных** - Поддержка множественных баз данных с управлением primary/read/write репликами
- **Логирование и производительность** - Структурированное логирование и отслеживание метрик производительности

### Валидация данных
- **Валидации** - Встроенные валидаторы для проверки данных (наличие, длина, email, числовые значения, формат)
- **Сбор ошибок** - Комплексный сбор и отчетность об ошибках

### Управление базами данных
- **Миграции** - Управление схемой базы данных с контролем версий
- **Построитель таблиц** - DSL для создания и изменения таблиц
- **Пул соединений** - Эффективное управление соединениями с базой данных
- **Проверки здоровья** - Мониторинг состояния базы данных

### Продвинутые функции
- **Обработка NULL значений** - Правильная обработка NULL значений базы данных
- **Отображение на основе рефлексии** - Динамическое обнаружение полей для сложных структур
- **Фреймворк ассоциаций** - Управление связями между моделями
- **Поддержка контекста** - Операции с поддержкой контекста для отмены и таймаутов

## 📦 Установка

```bash
go get github.com/Forester-Co/go-active-record
```

## 🚀 Быстрый старт

### Подключение к базе данных

```go
package main

import (
    "log"
    "github.com/Forester-Co/go-active-record/activerecord"
)

func main() {
    // Подключение к SQLite (для разработки)
    db, err := activerecord.Connect("sqlite3", ":memory:")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    // Установка глобального соединения
    activerecord.SetConnection(db)
}
```

### Определение модели

```go
type User struct {
    activerecord.HookableModel  // Включает хуки и методы Active Record
    Name     string `db:"name" json:"name"`
    Email    string `db:"email" json:"email"`
    Age      int    `db:"age" json:"age"`
    Database string `db:"database" json:"database"`
}

// TableName возвращает имя таблицы
func (u *User) TableName() string {
    return "users"
}

// SetupHooks настраивает хуки модели
func (u *User) SetupHooks() {
    u.AddHook(activerecord.BeforeCreate, func(model interface{}) error {
        user := model.(*User)
        fmt.Printf("Создание пользователя: %s\n", user.Name)
        return nil
    })
    
    u.AddHook(activerecord.AfterCreate, func(model interface{}) error {
        user := model.(*User)
        fmt.Printf("Создан пользователь с ID: %v\n", user.GetID())
        return nil
    })
}
```

### CRUD операции

```go
// Создание с хуками
user := &User{
    Name:  "Иван Иванов",
    Email: "ivan@example.com",
    Age:   30,
}
user.SetupHooks()
err := user.Create()

// Чтение по ID
foundUser := &User{}
err = activerecord.Find(foundUser, 1)

// Чтение всех записей
var users []*User
err = activerecord.FindAll(&users)

// Поиск с условиями
var youngUsers []*User
err = activerecord.Where(&youngUsers, "age < ?", 25)

// Обновление
foundUser.Age = 31
err = foundUser.Update()

// Удаление
err = foundUser.Delete()

// Сохранение (создает или обновляет)
err = user.Save()
```

### Построитель запросов

```go
// Создание построителя запросов
qb := activerecord.NewQueryBuilder("users")
qb.Where("age > ?", 25).
   Where("email LIKE ?", "%@example.com").
   OrderBy("age", "ASC").
   Limit(10).
   Offset(0)

// Выполнение запроса
var users []*User
err := qb.Find(&users)

// Dry run для отладки
qb.DryRun(true)
err = qb.Find(&users) // Выводит запрос без выполнения
```

### Пакетные операции

```go
// Пакетная вставка
users := []interface{}{
    &User{Name: "Пользователь 1", Email: "user1@example.com", Age: 25},
    &User{Name: "Пользователь 2", Email: "user2@example.com", Age: 30},
    &User{Name: "Пользователь 3", Email: "user3@example.com", Age: 35},
}

result, err := activerecord.BatchInsert(users)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Вставлено %d пользователей\n", result.RowsAffected)

// Поиск пакетами
err = activerecord.FindInBatches(&User{}, 100, func(batch []interface{}) error {
    for _, user := range batch {
        // Обработка каждого пользователя
        fmt.Printf("Обработка пользователя: %v\n", user.(*User).Name)
    }
    return nil
})

// Найти или создать
user := &User{Email: "new@example.com"}
conditions := map[string]interface{}{"email": "new@example.com"}
err = activerecord.FindOrCreate(user, conditions)
```

### Транзакции

```go
// Начало транзакции
tx, err := activerecord.Begin()
if err != nil {
    log.Fatal(err)
}
defer tx.Rollback()

// Создание пользователя в транзакции
user := &User{Name: "Транзакционный пользователь", Email: "tx@example.com"}
err = user.Create()
if err != nil {
    return err
}

// Создание связанной записи
profile := &Profile{UserID: user.GetID(), Bio: "Тест транзакции"}
err = profile.Create()
if err != nil {
    return err
}

// Подтверждение транзакции
err = tx.Commit()
```

### Резолвер баз данных (Поддержка множественных БД)

```go
// Создание менеджера баз данных
dm := activerecord.NewDatabaseManager()

// Настройка основной базы данных
primaryResolver := activerecord.NewDatabaseResolver()
primaryConfig := &activerecord.DatabaseConfig{
    Driver:   "sqlite3",
    DSN:      "primary.db",
    MaxOpen:  10,
    MaxIdle:  5,
    Lifetime: time.Hour,
}
primaryResolver.SetPrimary(primaryConfig)

// Добавление read реплики
readConfig := &activerecord.DatabaseConfig{
    Driver:   "sqlite3",
    DSN:      "read_replica.db",
    MaxOpen:  10,
    MaxIdle:  5,
    Lifetime: time.Hour,
}
primaryResolver.AddReadReplica(readConfig)

// Добавление write реплики
writeConfig := &activerecord.DatabaseConfig{
    Driver:   "sqlite3",
    DSN:      "write_replica.db",
    MaxOpen:  10,
    MaxIdle:  5,
    Lifetime: time.Hour,
}
primaryResolver.AddWriteReplica(writeConfig)

// Добавление в менеджер
dm.AddDatabase("myapp", primaryResolver)
activerecord.SetDatabaseManager(dm)

// Использование операций с осведомленностью о БД
user := &User{Name: "Мульти-БД пользователь", Email: "multidb@example.com"}
err := activerecord.CreateOnDatabase("myapp", user)

foundUser := &User{}
err = activerecord.FindOnDatabase("myapp", foundUser, user.GetID())
```

### Логирование и производительность

```go
// Настройка структурированного логирования
logger := activerecord.NewStructuredLogger()
logger.SetLevel(activerecord.DebugLevel)
activerecord.SetLogger(logger)

// Логированные операции
result, err := activerecord.LoggedExec("INSERT INTO users (name, email) VALUES (?, ?)", "Лог пользователь", "log@example.com")

// Метрики производительности
stats := activerecord.GetPerformanceStats()
fmt.Printf("Всего запросов: %d\n", stats["total_queries"])

// Логирование приложения
activerecord.LogInfo("Пользователь создан", map[string]interface{}{
    "user_id": user.GetID(),
    "action":  "create",
})
```

### Валидации

```go
type User struct {
    activerecord.ValidationModel
    Name  string `db:"name" json:"name"`
    Email string `db:"email" json:"email"`
    Age   int    `db:"age" json:"age"`
}

func (u *User) TableName() string {
    return "users"
}

// Настройка валидаций
func (u *User) SetupValidations() {
    u.PresenceOf("Name")
    u.AddValidation("Email", "email", "имеет неверный формат")
    u.Length("Name", 2, 50)
    u.Numericality("Age", 18, 100)
    u.Format("Email", `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`, "неверный формат email")
}

// Валидация
user := &User{Name: "", Email: "invalid-email", Age: 15}
user.SetupValidations()

if !user.IsValid() {
    fmt.Println("Ошибки валидации:", user.Errors())
}
```

### Миграции

```go
// Создание миграции
migration := activerecord.NewMigration("create_users_table")
migration.Up = func() error {
    return activerecord.CreateTable("users", func(t *activerecord.TableBuilder) {
        t.Integer("id").PrimaryKey().AutoIncrement()
        t.String("name").NotNull()
        t.String("email").NotNull().Unique()
        t.Integer("age")
        t.Timestamp("created_at").NotNull()
        t.Timestamp("updated_at").NotNull()
    })
}

migration.Down = func() error {
    return activerecord.DropTable("users")
}

// Запуск миграции
err := migration.Migrate()

// Проверка статуса миграции
status := migration.Status()
fmt.Printf("Статус миграции: %s\n", status)
```

## 🧪 Тестирование

Библиотека включает комплексные тесты, покрывающие все функции:

```bash
# Запуск всех тестов
go test ./activerecord -v

# Запуск конкретного теста
go test ./activerecord -v -run TestFullFeaturedORM

# Запуск бенчмарков
go test ./activerecord -bench=.
```

## 📊 Производительность

Библиотека оптимизирована для производительности с функциями:
- Пул соединений
- Подготовленные запросы
- Пакетные операции
- Эффективное использование рефлексии
- Памятосберегающий дизайн

## 🔧 Конфигурация

### Конфигурация базы данных

```go
// SQLite
db, err := activerecord.Connect("sqlite3", ":memory:")

// PostgreSQL
db, err := activerecord.Connect("postgres", "host=localhost port=5432 user=postgres password=password dbname=testdb sslmode=disable")

// MySQL
db, err := activerecord.Connect("mysql", "user:password@tcp(localhost:3306)/testdb")
```

### Конфигурация логирования

```go
// Структурированное логирование
logger := activerecord.NewStructuredLogger()
logger.SetLevel(activerecord.DebugLevel)
activerecord.SetLogger(logger)

// Пользовательский логгер
activerecord.SetLogger(customLogger)
```

## 🤝 Участие в разработке

1. Форкните репозиторий
2. Создайте ветку для функции
3. Внесите изменения
4. Добавьте тесты для новой функциональности
5. Убедитесь, что все тесты проходят
6. Отправьте pull request

## 📄 Лицензия

Этот проект лицензирован под MIT License - см. файл [LICENSE](LICENSE) для подробностей.

## 🙏 Благодарности

- Вдохновлен Ruby on Rails Active Record
- Построен с использованием стандартного пакета Go `database/sql`
- Использует рефлексию для динамического отображения полей
- Реализует современные паттерны и лучшие практики Go 