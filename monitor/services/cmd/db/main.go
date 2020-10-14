package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"git.condensat.tech/bank/appcontext"
	"git.condensat.tech/bank/database"

	monitordb "git.condensat.tech/bank/monitor/database"
	"git.condensat.tech/bank/monitor/database/model"
)

func main() {
	var opt database.Options
	database.OptionArgs(&opt)
	flag.Parse()

	ctx := context.Background()
	ctx = appcontext.WithDatabase(ctx, database.New(opt))

	step := 15 * time.Second
	timeframe := 10 * time.Minute
	to := time.Now().UTC().Truncate(step)
	from := to.Add(-timeframe)
	round := time.Minute

	db := appcontext.Database(ctx)

	apps, err := monitordb.ListServices(db, timeframe)
	if err != nil {
		panic(err)
	}

	var serviceMap = make(map[string][]model.ProcessInfo)
	for _, appName := range apps {
		services, err := monitordb.LastServiceHistory(db, appName, from, to, step, round)
		if err != nil {
			panic(err)
		}

		for _, service := range services {
			serviceName := fmt.Sprintf("%s:%s", service.AppName, service.Hostname)
			serviceMap[serviceName] = append(serviceMap[serviceName], service)
		}
	}

	fmt.Printf("%d services:\n", len(serviceMap))
	for serviceName, history := range serviceMap {
		fmt.Printf("  %s: %d\n", serviceName, len(history))
		for _, info := range history {
			fmt.Printf("    %s, %5.2f %%, %5.1f MiB\n", info.Timestamp, info.CPUUsage, float64(info.MemAlloc)/float64(1<<20))
		}
	}
}
