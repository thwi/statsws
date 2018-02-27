package main

import (
  "net/http"
  "./statsws"
)

const BIND_ADDRESS = "0.0.0.0:8080"
const COUNT        = 12
const INTERVAL     = 5

func main() {
  s := statsws.NewStatsws(COUNT, INTERVAL)
  s.Start()
  http.HandleFunc("/", s.HandleRequest)
  panic(http.ListenAndServe(BIND_ADDRESS, nil))
}
