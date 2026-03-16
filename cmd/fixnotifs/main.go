package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()
	dsn := os.Getenv("POSTGRES_DSN")
	if dsn == "" {
		dsn = "postgres://agristream:agristream@localhost:5433/agristream?sslmode=disable"
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	replacements := [][2]string{
		{"Kavango Maize Belt", "Field 1"},
		{"Okavango Sorghum Banks", "Field 2"},
		{"Sandveld Millet Ridge", "Field 3"},
		{"Kalahari Groundnut Flats", "Field 4"},
	}
	for _, r := range replacements {
		old, newName := r[0], r[1]
		res, err := pool.Exec(ctx,
			`UPDATE notifications SET
			   title = REPLACE(title, $1, $2),
			   body  = REPLACE(body,  $1, $2)
			 WHERE title LIKE '%' || $1 || '%' OR body LIKE '%' || $1 || '%'`,
			old, newName)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("  %-28s → %-8s  (%d rows updated)\n", old, newName, res.RowsAffected())
	}
	fmt.Println("done")
}
