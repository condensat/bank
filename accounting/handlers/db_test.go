package handlers

import (
	"context"
	"fmt"
	"math/rand"
	"reflect"
	"sort"

	"git.condensat.tech/bank"
	"git.condensat.tech/bank/database"
	"git.condensat.tech/bank/database/model"

	"github.com/jinzhu/gorm"
)

var testCtx = context.Background()
var testUsers = []model.User{
	bankUser,
	customerUser,
}
var bankUser = model.User{
	ID:    model.UserID(1),
	Name:  "CondensatBank",
	Email: "bank@condensat.tech",
}
var customerUser = model.User{
	ID:    model.UserID(2),
	Name:  "12345678901",
	Email: "12345678901@condensat.tech",
}

var validIban = "CH5604835012345678009"
var validBic = "BCNNCH22"

var currencies = []model.Currency{
	model.NewCurrency("CHF", "swiss franc", 0, 1, 0, 2),
	model.NewCurrency("BTC", "bitcoin", 1, 1, 1, 8),
}

const (
	initAmount = 100.0
	wAmt       = 25.0
	feeRate    = 0.002
	fiatMinFee = 0.5
)

func setup(databaseName string, models []model.Model) bank.Database {
	options := database.Options{
		HostName:      "mariadb",
		Port:          3306,
		User:          "condensat",
		Password:      "condensat",
		Database:      "condensat",
		EnableLogging: false,
	}
	if databaseName == options.Database {
		panic("Wrong databaseName")
	}

	db := database.NewDatabase(options)
	gdb := db.DB().(*gorm.DB)

	createDatabase := fmt.Sprintf("create database if not exists %s; use %s;", databaseName, databaseName)
	gdb.Exec(createDatabase)

	err := gdb.Exec(createDatabase).Error
	if err != nil {
		panic(err)
	}

	migrateDatabase(db, models)

	return db
}

func teardown(db bank.Database, databaseName string) {
	gdb := db.DB().(*gorm.DB)

	dropDatabase := fmt.Sprintf("drop database if exists %s", databaseName)
	err := gdb.Exec(dropDatabase).Error
	if err != nil {
		panic(err)
	}
}

func migrateDatabase(db bank.Database, models []model.Model) {
	err := db.Migrate(models)
	if err != nil {
		panic(err)
	}
}

func getSortedTypeFileds(t reflect.Type) []string {
	count := t.NumField()
	result := make([]string, 0, count)

	for i := 0; i < count; i++ {
		field := gorm.TheNamingStrategy.Column(t.Field(i).Name)
		result = append(result, field)
	}

	for i, field := range result {
		result[i] = gorm.TheNamingStrategy.Column(field)
	}
	sort.Strings(result)

	return result
}

var (
	letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
)

func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
