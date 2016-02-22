package auth

type songkick struct{
  authProvider
}

var Songkick = songkick{
  authProvider: authProvider{
    Name:         "Songkick",
    ClientId:     "",
    ClientSecret: "",
    AuthURL:      "https://www.songkick.com/oauth2/login",
    TokenURL:     "https://www.songkick.com/oauth2/token",
  },
}

func (p *songkick) LoginLink(fromPage string) string{
  config := p.OauthConfig()
  return config.AuthCodeURL(fromPage)
}
