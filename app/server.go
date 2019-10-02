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

// func handleHTTP(w http.ResponseWriter, r *http.Request) {

// 	uri := r.RequestURI

// 	fmt.Println(r.Method + ": " + uri)

// 	if r.Method == "POST" {
// 		body, err := ioutil.ReadAll(r.Body)
// 		log.Fatal(err)
// 		fmt.Printf("Body: %v\n", string(body))
// 	}

// 	rr, err := http.NewRequest(r.Method, uri, r.Body)
// 	log.Fatal(err)
// 	// copyHeaders(r, rr)

// 	// Create a client and query the target
// 	var transport http.Transport
// 	resp, err := transport.RoundTrip(rr)
// 	log.Fatal(err)

// 	fmt.Printf("Resp-Headers: %v\n", resp.Header)

// 	defer resp.Body.Close()
// 	body, err := ioutil.ReadAll(resp.Body)
// 	log.Fatal(err)

// 	dH := w.Header()
// 	copyHeaders(resp, w)
// 	dH.Add("Requested-Host", rr.Host)

// 	w.Write(body)
// }

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

func (s *Server) Run() {

	fmt.Println("step 1")

	var err error

	// Connect to DB
	s.DB, err = InitDatabase()

	fmt.Println("step 2")

	if err != nil {
		log.Fatal(err)
	}

	// Generate root CA keys
	// CA, _ := getKey()

	fmt.Println("4")

	if err != nil {
		log.Fatal(err)
	}

	serv := &Server{
		DB: s.DB,
		// CA: &CA,
	}

	fmt.Println("5")

	log.Fatal(http.ListenAndServe(":8080", serv))
}
