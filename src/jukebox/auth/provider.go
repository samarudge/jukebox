package auth

import (
  "golang.org/x/oauth2"
  "net/http"
)

type UserData struct{
  ProviderId    string
  ProfilePhoto  string
  Name          string
}

type AuthProvider struct{
  Name          string
  ClientId      string
  ClientSecret  string
  AuthURL       string
  TokenURL      string
  Scopes        []string
}

func (p *AuthProvider) OauthEndpoint() oauth2.Endpoint{
  a := oauth2.Endpoint{
    AuthURL:  p.AuthURL,
    TokenURL: p.TokenURL,
  }

  return a
}

func (p *AuthProvider) OauthConfig() oauth2.Config{
  a := oauth2.Config{
    ClientID:     p.ClientId,
    ClientSecret: p.ClientSecret,
    Scopes:       p.Scopes,
    Endpoint:     p.OauthEndpoint(),
  }
  a.RedirectURL = p.RedirectURL()

  return a
}

func (p *AuthProvider) RedirectURL() string{
  return "http://localhost:8080/auth/callback"
}

func (p *AuthProvider) DoExchange(code string) (*oauth2.Token, error){
  config := p.OauthConfig()
  token, err := config.Exchange(oauth2.NoContext, code)

  return token, err
}

func (p *AuthProvider) OauthClient(token *oauth2.Token) *http.Client{
  config := p.OauthConfig()
  client := config.Client(oauth2.NoContext, token)
  return client
}