package database

type Model interface{}

// Database (GORM)
type DB interface{}

type Context interface {
	DB() DB

	Migrate(models []Model) error
	Transaction(txFunc func(tx Context) error) error
}
