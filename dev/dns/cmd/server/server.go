package main

import (
    "log"
    "net/http"
)

func handle(w http.ResponseWriter, req *http.Request) {
    w.Write([]byte("Hello"))
}

func main() {
    http.HandleFunc("/", handle)
    if err := http.ListenAndServe(":8080", nil); err != nil {
        log.Fatal(err)
    }
}
