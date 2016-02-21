package auth

type songkick struct{
  AuthProvider
}

var Songkick = songkick{
  AuthProvider: AuthProvider{
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
