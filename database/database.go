package database

import (
	syslog "log"
	"os"

	"git.condensat.tech/bank"

	"github.com/jinzhu/gorm"
)

type Database struct {
	db *gorm.DB
}

// NewDatabase create new mysql connection
// pannic if failed to connect
func NewDatabase(options Options) *Database {
	db := connectMyql(
		options.HostName, options.Port,
		options.User, options.Password,
		options.Database,
	)

	db.LogMode(options.EnableLogging)
	db.SetLogger(syslog.New(os.Stderr, "", 0))

	return &Database{
		db: db,
	}
}

// DB returns subsequent db connection
// see bank.Database interface
func (d *Database) DB() bank.DB {
	return d.db
}
