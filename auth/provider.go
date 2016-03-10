package auth

import (
  "golang.org/x/oauth2"
  "net/http"
  "time"
  "fmt"
  "github.com/Machiel/slugify"
)

var ConfiguredProviders []string
var Providers = make(map[string]OauthProvider)

func LoadProvider(providerName string, p BaseProvider, additionalConfig map[interface{}]interface{}) OauthProvider{
  providers := []OauthProvider{
    NewGoogle(p, additionalConfig),
    NewSongkick(p, additionalConfig),
    NewSpotify(p, additionalConfig),
  }

  for _, provider := range providers{
    if provider.Provider().ProviderSlug() == providerName{
      return provider
    }
  }

  panic(fmt.Sprintf("Was asked to load %s provider but it doesn't exist", providerName))
}

type UserData struct{
  ProfilePhoto  string
  Name          string
  Username      string
}

type BaseProvider struct{
  Name          string
  ClientId      string
  ClientSecret  string
  AuthURL       string
  TokenURL      string
  Scopes        []string
  ReauthEvery   time.Duration
  RedirectURL   string
}

type OauthProvider interface{
  Provider()                        *BaseProvider
  ProviderSlug()                    string
  GetUserData(token *oauth2.Token)  (string, UserData, error)

  OauthEndpoint()                   oauth2.Endpoint
  OauthConfig()                     oauth2.Config

  LoginLink(string)                 string

  DoExchange(string)                (*oauth2.Token, error)
  OauthClient(*oauth2.Token)        *http.Client
}

func (p *BaseProvider) Provider() *BaseProvider{
  return p
}

func (p *BaseProvider) OauthEndpoint() oauth2.Endpoint{
  a := oauth2.Endpoint{
    AuthURL:  p.AuthURL,
    TokenURL: p.TokenURL,
  }

  return a
}

func (p *BaseProvider) OauthConfig() oauth2.Config{
  a := oauth2.Config{
    ClientID:     p.ClientId,
    ClientSecret: p.ClientSecret,
    Scopes:       p.Scopes,
    Endpoint:     p.OauthEndpoint(),
  }
  a.RedirectURL = p.RedirectURL

  return a
}

func (p *BaseProvider) LoginLink(state string) string{
  config := p.OauthConfig()
  return config.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

func (p *BaseProvider) DoExchange(code string) (*oauth2.Token, error){
  config := p.OauthConfig()
  token, err := config.Exchange(oauth2.NoContext, code)

  return token, err
}

func (p *BaseProvider) OauthClient(token *oauth2.Token) *http.Client{
  config := p.OauthConfig()
  client := config.Client(oauth2.NoContext, token)
  return client
}

func (p *BaseProvider) MakeProviderId(id string) string{
  return fmt.Sprintf("%s/%s", p.ProviderSlug(), id)
}

func (p *BaseProvider) ProviderSlug() string{
  return slugify.Slugify(p.Name)
}
