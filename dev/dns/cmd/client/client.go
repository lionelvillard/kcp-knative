package main

import (
    "io"
    "log"
    "net/http"
    "time"
)

func main() {
    for {
        resp, err := http.DefaultClient.Get("http://pong.server.svc.cluster.local")
        if err != nil {
            log.Printf("ping failed: %v", err)
        } else {
            if resp.StatusCode == 200 {
                log.Println("ping succeeded")
            } else {
                payload, err := io.ReadAll(resp.Body)
                if err != nil {
                    log.Printf("an error occurred while reading the HTTP response: %v\n", err)
                } else {
                    resp.Body.Close()
                    log.Printf("ping failed: %s (code: %d)\n", string(payload), resp.StatusCode)
                }
            }
        }

        time.Sleep(2 * time.Second)
    }
}
