package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"slices"
	"sort"

	"test-lintaspay/configs"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	down := flag.Bool("down", false, "run *.down.sql instead of *.up.sql")
	dir := flag.String("dir", "migrations", "migrations directory")
	flag.Parse()

	env, err := configs.NewViperFromEnv()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?multiStatements=true&parseTime=true",
		env.GetString("MYSQL_USER"),
		env.GetString("MYSQL_PASSWORD"),
		env.GetString("MYSQL_HOST"),
		env.GetString("MYSQL_PORT"),
		env.GetString("MYSQL_DATABASE"),
	)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("failed to ping database: %v", err)
	}

	pattern := "*.up.sql"
	if *down {
		pattern = "*.down.sql"
	}

	files, err := filepath.Glob(filepath.Join(*dir, pattern))
	if err != nil {
		log.Fatalf("failed to list migration files: %v", err)
	}
	if len(files) == 0 {
		log.Fatalf("no %s files found in %s", pattern, *dir)
	}

	sort.Strings(files)
	if *down {
		slices.Reverse(files)
	}

	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			log.Fatalf("failed to read %s: %v", file, err)
		}

		if _, err := db.Exec(string(content)); err != nil {
			log.Fatalf("failed to apply %s: %v", file, err)
		}

		fmt.Printf("applied %s\n", file)
	}
}
