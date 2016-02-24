package auth

import(
  "golang.org/x/oauth2"
)

/*
AuthURL:      "https://www.songkick.com/oauth2/login",
TokenURL:     "https://www.songkick.com/oauth2/token",
*/

type Songkick struct{
  *BaseProvider
  //UserData func(token *oauth2.Token) (UserData, error)
}

func NewSongkick(p *BaseProvider) *Songkick{
  p.Name =      "Songkick"
  p.AuthURL =   "https://www.songkick.com/oauth2/login"
  p.TokenURL =  "https://www.songkick.com/oauth2/token"

  return &Songkick{
    BaseProvider: p,
  }
}

func (p *Songkick) GetUserData(token *oauth2.Token) (UserData, error){
  _ = p.OauthClient(token)
  user := UserData{}

  return user, nil
}
