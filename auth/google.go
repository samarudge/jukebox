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
  BaseProvider
}

func NewGoogle(p BaseProvider, _ map[interface{}]interface{}) *Google{
  p.Name =        "Google Apps"
  p.AuthURL =     "https://accounts.google.com/o/oauth2/auth"
  p.TokenURL =    "https://www.googleapis.com/oauth2/v3/token"
  p.Scopes =      []string{"profile", "email"}
  p.ReauthEvery = time.Minute*15

  return &Google{
    BaseProvider: p,
  }
}

func (p *Google) GetUserData(token *oauth2.Token) (string, UserData, error){
  client := p.OauthClient(token)
  user := UserData{}
  var ProviderId string

  rsp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
  if err != nil{
    return ProviderId, user, err
  }

  log.WithFields(log.Fields{
    "call": rsp.Request.URL,
    "status": rsp.StatusCode,
  }).Debug("User Data Google")

  defer rsp.Body.Close()
  responseRaw, _ := ioutil.ReadAll(rsp.Body)
  userData := make(map[string]interface{})

  if rsp.StatusCode == 200 {
    if err := json.Unmarshal(responseRaw, &userData); err != nil {
      return ProviderId, user, fmt.Errorf("Could not decode JSON", string(responseRaw))
    }

    ProviderId = p.MakeProviderId(userData["id"].(string))
    user.ProfilePhoto = userData["picture"].(string)
    user.Name = userData["name"].(string)
  } else {
    return ProviderId, user, fmt.Errorf("Could not get user data: %s %s", rsp.StatusCode, responseRaw)
  }

  return ProviderId, user, nil
}
