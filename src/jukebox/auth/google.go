package auth

import(
  "fmt"
  "golang.org/x/oauth2"
  "io/ioutil"
  log "github.com/Sirupsen/logrus"
  "encoding/json"
  "time"
)

type Google struct{
  *BaseProvider
  //UserData func(token *oauth2.Token) (UserData, error)
}

func NewGoogle(p *BaseProvider) *Google{
  p.Name =        "Google Apps"
  p.AuthURL =     "https://accounts.google.com/o/oauth2/auth"
  p.TokenURL =    "https://www.googleapis.com/oauth2/v3/token"
  p.Scopes =      []string{"profile", "email"}
  p.ReauthEvery = time.Minute*60

  return &Google{
    BaseProvider: p,
  }
}

func (p *BaseProvider) GetUserData(token *oauth2.Token) (UserData, error){
  client := p.OauthClient(token)
  user := UserData{}

  rsp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
  if err != nil{
    return user, err
  }

  log.WithFields(log.Fields{
    "call": rsp.Request.URL,
    "status": rsp.StatusCode,
  }).Info("User Data Google")

  defer rsp.Body.Close()
  responseRaw, _ := ioutil.ReadAll(rsp.Body)
  userData := make(map[string]interface{})

  if rsp.StatusCode == 200 {
    if err := json.Unmarshal(responseRaw, &userData); err != nil {
      return user, fmt.Errorf("Could not decode JSON", string(responseRaw))
    }

    user.ProviderId = userData["id"].(string)
    user.ProfilePhoto = userData["picture"].(string)
    user.Name = userData["name"].(string)
  } else {
    return user, fmt.Errorf("Could not get user data: %s %s", rsp.StatusCode, responseRaw)
  }

  return user, nil
}
