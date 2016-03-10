package config

import(
  "os"
  log "github.com/Sirupsen/logrus"
  "gopkg.in/yaml.v2"
  "io/ioutil"
  "github.com/samarudge/jukebox/auth"
  "net/url"
  "fmt"
)

type config struct{
  Secret    string
  Url       string
  Auth      struct{
    Configured_providers       []string
  }
}

var Config config

func Initialize(filePath string){
  Config = config{}
  ConfigInterface := make(map[interface{}]interface{})
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

  _ = yaml.Unmarshal(configContent, &ConfigInterface)

  // Validate config params
  if  Config.Secret == "" ||
      Config.Url == "" ||
      len(Config.Auth.Configured_providers) == 0 {

      log.WithFields(log.Fields{
        "configFile": filePath,
        "error": "Invalid configuration provided",
        "pl":len(Config.Auth.Configured_providers),
      }).Error("Could not load config file")
      os.Exit(1)
  }

  Config.Auth.Configured_providers = append(Config.Auth.Configured_providers, "spotify")
  authConfig := ConfigInterface["auth"].(map[interface{}]interface{})

  // Load the auth providers
  for _,providerName := range Config.Auth.Configured_providers{
    providerConfig, found := authConfig[providerName].(map[interface{}]interface{})
    if !found{
      log.WithFields(log.Fields{
        "configFile": filePath,
        "provider": providerName,
      }).Error("Auth provider supplied but not configured. Add the providers auth keys to the config file.")
      os.Exit(1)
    }

    p := auth.BaseProvider{}
    p.ClientId = providerConfig["client_id"].(string)
    p.ClientSecret = providerConfig["client_secret"].(string)

    u, _ := url.Parse(Config.Url)
    u.Path = fmt.Sprintf("/auth/callback/%s", providerName)

    p.RedirectURL = u.String()

    auth.Providers[providerName] = auth.LoadProvider(providerName, p, ConfigInterface)
  }

  auth.ConfiguredProviders = Config.Auth.Configured_providers
}
