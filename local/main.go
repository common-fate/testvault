package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"

	"github.com/common-fate/apikit/logger"
	"github.com/common-fate/ddb"
	"github.com/common-fate/testvault"
	"github.com/joho/godotenv"
)

func main() {
	err := run()
	if err != nil {
		log.Fatal(err)
	}
}

func run() error {
	ctx := context.Background()

	err := godotenv.Load()
	if err != nil {
		return err
	}

	log, err := logger.Build("info")
	if err != nil {
		return err
	}
	table := os.Getenv("TESTVAULT_TABLE_NAME")
	if table == "" {
		return errors.New("TESTVAULT_TABLE_NAME must be set")
	}
	db, err := ddb.New(ctx, table)
	if err != nil {
		return err
	}

	a, err := testvault.NewAPI(testvault.APIOpts{DB: db, Log: log})
	if err != nil {
		return err
	}
	host := "0.0.0.0:8085"
	log.Infow("starting testvault server", "host", host)

	return http.ListenAndServe(host, a.Server())
}
