package auth

import (
  "golang.org/x/oauth2"
  "net/http"
  "time"
)

var Provider OauthProvider

func LoadProvider(providerName string, p *BaseProvider){
  switch providerName{
    case "google":
      Provider = NewGoogle(p)
    case "songkick":
      Provider = NewSongkick(p)
    default:
      Provider = NewSongkick(p)
  }
}

type UserData struct{
  ProviderId    string
  ProfilePhoto  string
  Name          string
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
  GetUserData(token *oauth2.Token)  (UserData, error)

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
