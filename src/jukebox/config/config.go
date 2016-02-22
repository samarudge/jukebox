package config

import(
  "os"
  log "github.com/Sirupsen/logrus"
  "gopkg.in/yaml.v2"
  "io/ioutil"
  "fmt"
  "jukebox/auth"
  "net/url"
)

type config struct{
  Secret    string
  Url       string
  Auth      struct{
    Provider      string
    Client_id      string
    Client_secret  string
  }
}

var Config config

func Initialize(filePath string){
  Config = config{}
  if _, err := os.Stat(filePath); os.IsNotExist(err) {
    log.WithFields(log.Fields{
      "configFile": filePath,
      "error": "File not found",
    }).Error("Could not load config file")
    os.Exit(1)
  }

  configContent, err := ioutil.ReadFile(filePath)
  if err != nil {
    log.WithFields(log.Fields{
      "configFile": filePath,
      "error": err,
    }).Error("Could not load config file")
    os.Exit(1)
  }

  err = yaml.Unmarshal(configContent, &Config)
  if err != nil {
    log.WithFields(log.Fields{
      "configFile": filePath,
      "error": err,
    }).Error("Could not load config file")
    os.Exit(1)
  }

  fmt.Println(Config)

  // Validate config params
  if  Config.Secret == "" ||
      Config.Url == "" ||
      Config.Auth.Provider == "" ||
      Config.Auth.Client_id == "" ||
      Config.Auth.Client_secret == "" {

      log.WithFields(log.Fields{
        "configFile": filePath,
        "error": "Invalid configuration provided",
      }).Error("Could not load config file")
      os.Exit(1)
  }

  // Load the auth provider
  p := auth.BaseProvider{}
  p.ClientId = Config.Auth.Client_id
  p.ClientSecret = Config.Auth.Client_secret

  u, _ := url.Parse(Config.Url)
  u.Path = "/auth/callback"

  p.RedirectURL = u.String()
  auth.LoadProvider(Config.Auth.Provider, &p)

}
