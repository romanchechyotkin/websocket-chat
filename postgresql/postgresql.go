package postgresql

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

func New(username, password, host, port, database string) *pgxpool.Pool {
	connString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", username, password, host, port, database)
	log.Println(connString)

	ctx := context.Background()

	log.Println("postgresql client init")
	pool, err := pgxpool.New(ctx, connString)
	err = pool.Ping(ctx)
	if err != nil {
		log.Println(err)
		log.Fatal("cannot to connect to postgres")
	}

	return pool
}
