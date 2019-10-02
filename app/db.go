package app

import (
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/jackc/pgx"
)

func InitDatabase() (conn *pgx.Conn, err error) {
	conn, err = pgx.Connect(context.Background(), "postgres://gel0:1337@localhost:5432/proxy")

	if err != nil {
		return nil, errors.New("Can't conneсt to database, aborting.\n")
	}

	// defer conn.Close(context.Background())

	_, err = conn.Exec(context.Background(), `
		DROP TABLE IF EXISTS "requsts" CASCADE;
		DROP TABLE IF EXISTS "headers" CASCADE;

		CREATE TABLE IF NOT EXISTS "headers" (
			id serial primary key, 
			parent_id int, 
			key text, 
			value text
		);

		CREATE TABLE IF NOT EXISTS "requests" (
			id serial primary key, 
			method text, 
			url text, 
			http text,
			body text
		);
	`)
	return
}

func (s *Server) insertData(r *http.Request) {
	var (
		id int
		// m  sync.Mutex
	)

	s.Mut.Lock()

	err := s.DB.QueryRow(context.Background(),
		`INSERT 
		 INTO "requests" ("method", "url", "http")
		 VALUES ($1, $2, $3)
		 RETURNING id;`,
		r.Method, r.RequestURI, r.Proto).Scan(&id)

	if err != nil {
		log.Fatal(err)
	}

	for key, val := range r.Header {
		_, err = s.DB.Exec(context.Background(),
			`INSERT 
			 INTO "headers" (parent_id, key, value) 
			 VALUES($1, $2, $3);`,
			id, key, val[0])
	}

	if err != nil {
		log.Fatal(err)
	}

	s.Mut.Unlock()
}
