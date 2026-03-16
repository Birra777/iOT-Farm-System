// One-off script: deduplicate fields table and rename to Field 1-4.
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	_ = godotenv.Load()
	dsn := os.Getenv("POSTGRES_DSN")
	if dsn == "" {
		dsn = "postgres://agristream:agristream@localhost:5433/agristream?sslmode=disable"
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil { log.Fatal(err) }
	defer pool.Close()

	// For each zone_code keep only the row with the smallest ctid (first inserted),
	// delete all others, then rename.
	steps := []string{
		// 1. For each zone_code, re-point sensor_readings to the canonical (earliest) field ID
		`UPDATE sensor_readings sr
		 SET field_id = canon.id
		 FROM (
		   SELECT DISTINCT ON (zone_code) id, zone_code
		   FROM fields ORDER BY zone_code, ctid
		 ) canon
		 JOIN fields f ON f.zone_code = canon.zone_code
		 WHERE sr.field_id = f.id AND sr.field_id != canon.id`,

		// 2. Re-point alerts
		`UPDATE alerts a
		 SET field_id = canon.id
		 FROM (
		   SELECT DISTINCT ON (zone_code) id, zone_code
		   FROM fields ORDER BY zone_code, ctid
		 ) canon
		 JOIN fields f ON f.zone_code = canon.zone_code
		 WHERE a.field_id = f.id AND a.field_id != canon.id`,

		// 3. Re-point notifications
		`UPDATE notifications n
		 SET field_id = canon.id
		 FROM (
		   SELECT DISTINCT ON (zone_code) id, zone_code
		   FROM fields ORDER BY zone_code, ctid
		 ) canon
		 JOIN fields f ON f.zone_code = canon.zone_code
		 WHERE n.field_id = f.id AND n.field_id != canon.id`,

		// 4. Now safe to delete duplicate rows
		`DELETE FROM fields
		 WHERE id NOT IN (
		   SELECT DISTINCT ON (zone_code) id
		   FROM fields ORDER BY zone_code, ctid
		 )`,

		// 5. Rename to simple names
		`UPDATE fields SET name = 'Field 1' WHERE zone_code = 'NB'`,
		`UPDATE fields SET name = 'Field 2' WHERE zone_code = 'RB'`,
		`UPDATE fields SET name = 'Field 3' WHERE zone_code = 'DR'`,
		`UPDATE fields SET name = 'Field 4' WHERE zone_code = 'SP'`,

		// 6. Add unique constraint so ON CONFLICT works in future migrations
		`ALTER TABLE fields DROP CONSTRAINT IF EXISTS fields_zone_code_key`,
		`ALTER TABLE fields ADD CONSTRAINT fields_zone_code_key UNIQUE (zone_code)`,
	}

	for _, q := range steps {
		if _, err := pool.Exec(ctx, q); err != nil {
			log.Fatalf("query failed:\n%s\nerror: %v", q, err)
		}
	}

	// Verify
	rows, err := pool.Query(ctx, "SELECT zone_code, name, crop_type FROM fields ORDER BY zone_code")
	if err != nil { log.Fatal(err) }
	defer rows.Close()
	fmt.Println("Fields after fix:")
	for rows.Next() {
		var zone, name, crop string
		rows.Scan(&zone, &name, &crop)
		fmt.Printf("  [%s] %s (%s)\n", zone, name, crop)
	}
}
