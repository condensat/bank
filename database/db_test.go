package database

import (
	"context"
	"fmt"

	"git.condensat.tech/bank/appcontext"
	"git.condensat.tech/bank/database/model"
	"git.condensat.tech/bank/logger"

	"github.com/jinzhu/gorm"
)

func setup(ctx context.Context, databaseName string, models []model.Model) context.Context {
	options := Options{
		HostName:      "localhost",
		Port:          3306,
		User:          "condensat",
		Password:      "condensat",
		Database:      "condensat",
		EnableLogging: false,
	}
	if databaseName == options.Database {
		panic("Wrong databaseName")
	}

	ctx = appcontext.WithDatabase(ctx, NewDatabase(options))
	db := appcontext.Database(ctx).DB().(*gorm.DB)

	createDatabase := fmt.Sprintf("create database if not exists %s; use %s;", databaseName, databaseName)
	db.Exec(createDatabase)

	err := db.Exec(createDatabase).Error
	if err != nil {
		panic(err)
	}

	migrateDatabase(ctx, models)

	return ctx
}

func teardown(ctx context.Context, databaseName string) {
	db := appcontext.Database(ctx).DB().(*gorm.DB)

	dropDatabase := fmt.Sprintf("drop database if exists %s", databaseName)
	err := db.Exec(dropDatabase).Error
	if err != nil {
		panic(err)
	}
}

func migrateDatabase(ctx context.Context, models []model.Model) {
	db := appcontext.Database(ctx)

	err := db.Migrate(models)
	if err != nil {
		logger.Logger(ctx).WithError(err).
			WithField("Method", "main.migrateDatabase").
			Panic("Failed to migrate database models")
	}
}
