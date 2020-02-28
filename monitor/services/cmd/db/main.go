package main

import (
	"context"
	"flag"
	"fmt"

	"git.condensat.tech/bank/appcontext"
	"git.condensat.tech/bank/database"
	"git.condensat.tech/bank/monitor"
)

func main() {
	var db database.Options
	database.OptionArgs(&db)
	flag.Parse()

	ctx := context.Background()
	ctx = appcontext.WithDatabase(ctx, database.NewDatabase(db))

	services, err := monitor.LastServicesStatus(ctx)
	if err != nil {
		panic(err)
	}
	for _, service := range services {
		fmt.Printf("  %+v\n", service)
	}
}
