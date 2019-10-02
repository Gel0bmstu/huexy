package app

import (
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/jackc/pgx"
)

type Server struct {
	DB  *pgx.Conn
	CA  *tls.Certificate
	Mut sync.Mutex
}

func copyHeaders(r *http.Response, w http.ResponseWriter) {
	for key, v := range r.Header { // Read response from server and add it to our response to client
		for _, value := range v {
			w.Header().Add(key, value)
		}
	}
}

func handleHTTP(w http.ResponseWriter, r *http.Request) {
	transport := http.Transport{
		ResponseHeaderTimeout: 15 * time.Second,
		DisableKeepAlives:     false,
	}

	response, err := transport.RoundTrip(r) // Roundtrip request to server

	if err != nil {
		log.Println("Couldn't make round trip to host server:", err)
		return
	}

	defer response.Body.Close()

	copyHeaders(response, w)

	w.WriteHeader(response.StatusCode) // Set response status code
	io.Copy(w, response.Body)          // Copy server response body to our response body to client
}

func (s *Server) handleHTTPS(w http.ResponseWriter, r *http.Request) {
	// Connect to server
	dest_conn, err := net.DialTimeout("tcp", r.Host, 10*time.Second)

	if err != nil {
		log.Fatal(err)
	}

	w.WriteHeader(http.StatusOK)
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return
	}

	client_conn, _, err := hijacker.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
	}

	go s.insertData(r)
	go transfer(dest_conn, client_conn)
	go transfer(client_conn, dest_conn)
}
func transfer(destination io.WriteCloser, source io.ReadCloser) {
	defer destination.Close()
	defer source.Close()
	io.Copy(destination, source)
}

//ServeHTTP - handles request methods
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Println(r.Method, r.RequestURI, r.Proto, r.URL)

	if r.Method == http.MethodConnect { // Handle CONNECT method
		s.handleHTTPS(w, r)
	} else { // Handle other method
		handleHTTP(w, r)
	}
}

func Run() {
	var (
		err  error
		conn *pgx.Conn
	)

	// Connect to DB
	conn, err = InitDatabase()

	if err != nil {
		log.Fatal(err)
	}

	// Generate root CA keys
	CA, _ := getKey()

	if err != nil {
		log.Fatal(err)
	}

	serv := &Server{
		DB: conn,
		CA: &CA,
	}

	fmt.Println("Sever start. Listen and Serve on localhost:8080")

	log.Fatal(http.ListenAndServe(":8080", serv))
}
