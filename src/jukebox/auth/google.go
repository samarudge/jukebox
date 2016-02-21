package auth

import(
  "fmt"
  "golang.org/x/oauth2"
  "io/ioutil"
  log "github.com/Sirupsen/logrus"
  "encoding/json"
)

type google struct{
  AuthProvider
}

var Google = google{
  AuthProvider: AuthProvider{
    Name:         "Google Apps",
    ClientId:     "15473194917-1vtqe8qm3pjpiqre108ml72bqqoi9vhl.apps.googleusercontent.com",
    ClientSecret: "3cTZaEofrFdmt7tXLTk8386j",
    AuthURL:      "https://accounts.google.com/o/oauth2/auth",
    TokenURL:     "https://www.googleapis.com/oauth2/v3/token",
    Scopes:       []string{"profile", "email"},
  },
}

func (p *google) LoginLink(fromPage string) string{
  config := p.OauthConfig()
  return config.AuthCodeURL(fromPage)
}

func (p *google) UserData(token *oauth2.Token) (UserData, error){
  client := p.OauthClient(token)
  user := UserData{}

  rsp, _ := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")

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
