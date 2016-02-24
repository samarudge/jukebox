package main

import (
  "os"
  "github.com/samarudge/jukebox/config"
  "github.com/samarudge/jukebox/app"
  "github.com/voxelbrain/goptions"
  log "github.com/Sirupsen/logrus"
)

type options struct {
  Verbose   bool            `goptions:"-v, --verbose, description='Log verbosely'"`
  Help      goptions.Help   `goptions:"-h, --help, description='Show help'"`
  Config    string          `goptions:"-c, --config, description='Config Yaml file to use'"`
  Bind      string          `goptions:"-b, --bind, description='Port/Address to bind on, can also be specified with JB_BIND environment variable'"`
}

func main() {
  parsedOptions := options{}

  parsedOptions.Config = "./config.yml"
  parsedOptions.Bind = os.Getenv("JB_BIND")

  goptions.ParseAndFail(&parsedOptions)

  if parsedOptions.Verbose{
    log.SetLevel(log.DebugLevel)
  } else {
    log.SetLevel(log.InfoLevel)
  }

  log.SetFormatter(&log.TextFormatter{FullTimestamp:true})

  log.Debug("Logging verbosely!")

  config.Initialize(parsedOptions.Config)

  app.Start(parsedOptions.Bind)
}
