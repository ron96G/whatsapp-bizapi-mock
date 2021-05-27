package main

import (
	"bytes"
	"compress/gzip"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

func webhook(w http.ResponseWriter, r *http.Request) {

	w.WriteHeader(200)

	fmt.Printf("Header:\n %v\n", r.Header)
	content, _ := ioutil.ReadAll(r.Body)
	fmt.Printf("Body:\n %v\n", string(content))

	if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
		buf := bytes.NewBuffer(content)
		gz, _ := gzip.NewReader(buf)
		defer gz.Close()
		content, _ := ioutil.ReadAll(gz)
		fmt.Printf("Gzip-Body:\n %v\n", string(content))
	}
}

func main() {
	http.HandleFunc("/webhook", webhook)
	s, _ := GenerateServerTLS()
	l, _ := tls.Listen("tcp", "0.0.0.0:9000", s)
	http.Serve(l, nil)
}
