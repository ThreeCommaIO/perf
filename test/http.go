package main

import (
	"fmt"
	"log"
	"net/http"
)

func HandleRequest(w http.ResponseWriter, req *http.Request) {
	message := "hello worlddd\n"
	w.Write([]byte(message))
}

func main() {
	http.Handle("/", http.HandlerFunc(HandleRequest))
	fmt.Println("Listening on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
