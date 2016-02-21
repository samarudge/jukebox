package main

import (
  "os"
  "jukebox/app"
)

func main() {
  // TODO: Will panic if env var not set, validate?
  app.Start(os.Getenv("JB_PORT"))
}
