// Fetch prints the content found at a URL.
package main

import (
    "fmt"
    "io/ioutil"
    "net/http"
    "os"
    "strings"
    "time"
)

func main() {
    time.Sleep(60*time.Second)
    for _, url := range os.Args[1:] {
        resp, err := http.Get(url)
        if err != nil {
            fmt.Fprintf(os.Stderr, "fetch: %v\n", err)
            os.Exit(1)
        }
        b, err := ioutil.ReadAll(resp.Body)
        resp.Body.Close()
        if err != nil {
            fmt.Fprintf(os.Stderr, "fetch: reading %s: %v\n", url, err)
            os.Exit(1)
        }
//        fmt.Printf("%s", b)
        if strings.Contains(string(b), "Harbor") {
          fmt.Printf("sucess!\n")
        } else {
            os.Exit(1)
        }

    }
}

