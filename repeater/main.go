package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/jackc/pgx"
)

func main() {

	var (
		// base
		conn *pgx.Conn
		reqs pgx.Rows

		// err
		err error

		// flags
		id *int
		re *bool

		// req
		req_id int
		method string
		uri    string
		proto  string
	)

	conn, err = pgx.Connect(context.Background(), "postgres://gel0:1337@localhost:5432/proxy")

	if err != nil {
		log.Fatal(err)
	}

	defer conn.Close(context.Background())

	id = flag.Int("i", -1, "Repeat request ")
	re = flag.Bool("r", false, "Show all requests in database")

	flag.Parse()

	if *re != false {
		fmt.Println("ID		METHOD		REQUEST 		HTTP_V")
		reqs, err = conn.Query(context.Background(),
			`SELECT * FROM "requests";`)

		for reqs.Next() {
			reqs.Scan(&req_id, &method, &uri, &proto)

			fmt.Println(req_id, method, uri, proto)
		}

		if err != nil {
			log.Fatal(err)
		}
	}

	if *id != -1 {
		fmt.Println("step 1 ")
		repeatRequestById(conn, *id)
	}
}

func getRequestById(db *pgx.Conn, id int) (r *http.Request, err error) {

	var (
		rows pgx.Rows
		key  string
		val  string

		method, uri, proto string
	)

	err = db.QueryRow(context.Background(),
		`SELECT "method", "url", "http"
		 FROM "requests"
		 WHERE id = $1;`,
		id).Scan(&method, &uri, &proto)

	r, err = http.NewRequest(method, uri, nil)

	if err != nil {
		log.Fatal(err)
	}

	if err != nil {
		log.Fatal("err to get request from db: ", err)
	}

	rows, err = db.Query(context.Background(),
		`SELECT ("key", "value")
		 FROM "headers"
		 WHERE parent_id = $1;`,
		id)

	if err != nil {
		log.Fatal("err to get headerss of req from bd: ", err)
	}

	for rows.Next() {
		rows.Scan(&key, &val)
		if key != "If-None-Match" && key != "Accept-Encoding" && key != "If-Modified-Since" {
			r.Header.Add(key, val) // Add headers to request
		}
	}

	return
}

func printRes(r *http.Response) {
	fmt.Println("Response: ", r.Proto, r.Status)

	for key, value := range r.Header {
		fmt.Println(key, ":", value)
	}

	b, err := ioutil.ReadAll(r.Body)

	if err != nil {
		log.Fatal("err to parse body:", err)
	}

	fmt.Println("Body: ", string(b))
}

func repeatRequestById(db *pgx.Conn, id int) {
	r, err := getRequestById(db, id)

	if err != nil {
		log.Fatal("err to get request: ", err)
	}

	client := &http.Client{}

	fmt.Println(r.RequestURI, r.Proto, r.Method)

	res, err := client.Do(r)

	if err != nil {
		log.Fatal("err tot send request: ", err)
	}

	defer res.Body.Close()

	printRes(res)
}
