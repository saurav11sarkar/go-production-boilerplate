package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/saurav11sarkar/go-production-boilerplate/internal/config"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: go run ./cmd/migrate [up|down|force VERSION]")
		os.Exit(2)
	}
	cfg := config.MustLoad()
	m, err := migrate.New("file://migrations", cfg.Database.URL)
	if err != nil {
		panic(err)
	}
	defer m.Close()

	switch os.Args[1] {
	case "up":
		err = m.Up()
	case "down":
		err = m.Steps(-1)
	case "force":
		if len(os.Args) != 3 {
			panic("force requires a migration version")
		}
		var version int
		if _, scanErr := fmt.Sscanf(os.Args[2], "%d", &version); scanErr != nil {
			panic(scanErr)
		}
		err = m.Force(version)
	default:
		panic("unknown migration command")
	}
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		panic(err)
	}
	fmt.Println("migration complete")
}
