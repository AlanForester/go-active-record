# Go Active Record

Библиотека Active Record для Go, вдохновленная Rails Active Record. Предоставляет удобный интерфейс для работы с базой данных, включая CRUD операции, валидации, ассоциации и миграции.

## Особенности

- 🚀 **CRUD операции** - создание, чтение, обновление, удаление записей
- ✅ **Валидации** - встроенные валидаторы для проверки данных
- 🔗 **Ассоциации** - связи между моделями (has_one, has_many, belongs_to)
- 📊 **Миграции** - управление схемой базы данных
- 🔍 **Query Builder** - удобный построитель запросов
- 🛡️ **Транзакции** - поддержка транзакций
- 📝 **Логирование** - встроенное логирование SQL запросов

## Установка

```bash
go get github.com/Forester-Co/go-active-record
```

## Быстрый старт

### Подключение к базе данных

```go
package main

import (
    "log"
    "github.com/Forester-Co/go-active-record/activerecord"
)

func main() {
    // Подключение к PostgreSQL
    db, err := activerecord.Connect("postgres", "host=localhost port=5432 user=postgres password=password dbname=testdb sslmode=disable")
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
    activerecord.ActiveRecordModel
    Name  string `db:"name" json:"name"`
    Email string `db:"email" json:"email"`
    Age   int    `db:"age" json:"age"`
}

// TableName возвращает имя таблицы
func (u *User) TableName() string {
    return "users"
}
```

### CRUD операции

```go
// Создание
user := &User{
    Name:  "John Doe",
    Email: "john@example.com",
    Age:   30,
}
err := user.Create()

// Чтение по ID
foundUser := &User{}
err = activerecord.Find(foundUser, 1)

// Чтение всех записей
var users []User
err = activerecord.FindAll(&users)

// Поиск с условиями
var youngUsers []User
err = activerecord.Where(&youngUsers, "age < ?", 25)

// Обновление
foundUser.Age = 31
err = foundUser.Update()

// Удаление
err = foundUser.Delete()
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
    u.Email("Email")
    u.Length("Name", 2, 50)
    u.Numericality("Age", 18, 100)
}

// Проверка валидности
user := &User{Name: "", Email: "invalid-email", Age: 15}
user.SetupValidations()

if !user.IsValid() {
    fmt.Println("Ошибки валидации:", user.Errors())
}
```

### Миграции

```go
type CreateUsersTable struct {
    activerecord.Migration
}

func (m *CreateUsersTable) Version() int64 {
    return 20231201000001
}

func (m *CreateUsersTable) Up() error {
    return activerecord.CreateTable("users", func(t *activerecord.TableBuilder) {
        t.Column("id", "SERIAL", "PRIMARY KEY")
        t.Column("name", "VARCHAR(255)", "NOT NULL")
        t.Column("email", "VARCHAR(255)", "UNIQUE", "NOT NULL")
        t.Column("age", "INTEGER")
        t.Timestamps()
        t.Index("email")
    })
}

func (m *CreateUsersTable) Down() error {
    return activerecord.DropTable("users")
}

// Выполнение миграций
func main() {
    migrator := activerecord.NewMigrator()
    migrations := []activerecord.Migration{
        &CreateUsersTable{},
    }
    
    err := migrator.Migrate(migrations)
    if err != nil {
        log.Fatal(err)
    }
}
```

### Ассоциации

Библиотека поддерживает автоматическое определение ассоциаций и ручное определение связей.

#### Автоматическое определение ассоциаций

Вы можете определить ассоциации, просто добавив поля в структуру:

```go
type User struct {
    activerecord.BaseModel
    Name     string
    MentorID int
    Mentor   *User  `db:"-"`  // BelongsTo ассоциация
    Mentees  []User `db:"-"`  // HasMany ассоциация
}

// Библиотека автоматически определяет и регистрирует ассоциации:
// - Поле Mentor (*User) -> BelongsTo ассоциация с внешним ключом "MentorID"
// - Поле Mentees ([]User) -> HasMany ассоциация с внешним ключом "MentorID"

// Использование
mentor := &User{Name: "Master"}
mentor.Create()

mentee := &User{Name: "Student", MentorID: mentor.GetID()}
mentee.Create()

// Загрузка ассоциаций
mentee.Load("Mentor")    // Загружает ментора
mentor.Load("Mentees")   // Загружает всех подопечных
```

#### Ручное определение ассоциаций

Вы также можете определить ассоциации вручную:

```go
type User struct {
    activerecord.ActiveRecordModel
    Name  string `db:"name" json:"name"`
    Email string `db:"email" json:"email"`
}

type Post struct {
    activerecord.ActiveRecordModel
    Title   string `db:"title" json:"title"`
    Content string `db:"content" json:"content"`
    UserID  int    `db:"user_id" json:"user_id"`
}

// Определение ассоциаций вручную
func (u *User) HasMany(name string, model interface{}, foreignKey string) {
    // Реализация has_many
}

func (p *Post) BelongsTo(name string, model interface{}, foreignKey string) {
    // Реализация belongs_to
}
```

#### Поддерживаемые типы ассоциаций

- **BelongsTo**: `*OtherModel` - связь "один-к-одному", где эта модель принадлежит другой
- **HasMany**: `[]OtherModel` - связь "один-ко-многим", где эта модель имеет много других
- **HasOne**: `*OtherModel` - связь "один-к-одному", где эта модель имеет одну другую
- **HasManyThrough**: сложные связи "многие-ко-многим" (планируется)

#### Загрузка ассоциаций

```go
// Загрузка одной ассоциации
user.Load("Mentor")

// Загрузка нескольких ассоциаций
user.Include("Mentor", "Mentees")
```

## Поддерживаемые базы данных

- PostgreSQL
- MySQL
- SQLite

## API Reference

### Основные методы

#### ActiveRecordModel

- `Create() error` - создает запись
- `Update() error` - обновляет запись
- `Delete() error` - удаляет запись
- `Save() error` - сохраняет запись (создает или обновляет)
- `IsNewRecord() bool` - проверяет, является ли запись новой
- `IsPersisted() bool` - проверяет, сохранена ли запись
- `Touch() error` - обновляет временные метки
- `Reload() error` - перезагружает данные из БД

#### Глобальные методы

- `Find(model Modeler, id interface{}) error` - поиск по ID
- `FindAll(models interface{}) error` - поиск всех записей
- `Where(models interface{}, query string, args ...interface{}) error` - поиск с условиями
- `Create(model Modeler) error` - создание записи
- `Update(model Modeler) error` - обновление записи
- `Delete(model Modeler) error` - удаление записи

### Валидаторы

- `PresenceOf(field string)` - проверка наличия значения
- `Length(field string, min, max int)` - проверка длины строки
- `Email(field string)` - проверка формата email
- `Uniqueness(field string)` - проверка уникальности
- `Numericality(field string, min, max float64)` - проверка числового значения
- `Format(field string, pattern string)` - проверка по регулярному выражению

### Миграции

- `CreateTable(tableName string, callback func(*TableBuilder)) error` - создание таблицы
- `DropTable(tableName string) error` - удаление таблицы
- `Column(name, dataType string, options ...string)` - добавление колонки
- `PrimaryKey(columns ...string)` - добавление первичного ключа
- `Index(columns ...string)` - добавление индекса
- `Timestamps()` - добавление временных меток

## Примеры

Полные примеры использования можно найти в файле `main.go`.

## Лицензия

MIT License

## Вклад в проект

Приветствуются pull request'ы и issue'ы!

## TODO

- [x] Реализация ассоциаций
- [ ] Поддержка транзакций
- [ ] Логирование SQL запросов
- [ ] Кэширование
- [ ] Поддержка других баз данных
- [ ] Тесты
- [ ] Документация API 