package auth

import(
  "fmt"
  "golang.org/x/oauth2"
  "io/ioutil"
  log "github.com/Sirupsen/logrus"
  "encoding/json"
  "time"
)

type Spotify struct{
  BaseProvider
}

func NewSpotify(p BaseProvider, _ map[interface{}]interface{}) *Spotify{
  p.Name =        "Spotify"
  p.AuthURL =     "https://accounts.spotify.com/authorize?show_dialog=true"
  p.TokenURL =    "https://accounts.spotify.com/api/token"
  p.Scopes =      []string{"playlist-read-private", "playlist-read-collaborative", "user-follow-read", "user-library-read", "user-read-private", "user-read-email"}
  p.ReauthEvery = time.Minute*30

  return &Spotify{
    BaseProvider: p,
  }
}

func (p *Spotify) GetUserData(token *oauth2.Token) (string, UserData, error){
  client := p.OauthClient(token)
  user := UserData{}
  var ProviderId string

  rsp, err := client.Get("https://api.spotify.com/v1/me")
  if err != nil{
    return ProviderId, user, err
  }

  log.WithFields(log.Fields{
    "call": rsp.Request.URL,
    "status": rsp.StatusCode,
  }).Debug("User Data Spotify")

  defer rsp.Body.Close()
  responseRaw, _ := ioutil.ReadAll(rsp.Body)
  userData := make(map[string]interface{})

  if rsp.StatusCode == 200 {
    if err := json.Unmarshal(responseRaw, &userData); err != nil {
      return ProviderId, user, fmt.Errorf("Could not decode JSON", string(responseRaw))
    }

    ProviderId = p.MakeProviderId(userData["id"].(string))

    profileImages := userData["images"].([]interface{})
    proWidth := 0
    for _,imgIf := range(profileImages){
      im := imgIf.(map[string]interface{})
      width := im["width"].(int)
      if width > proWidth{
        proWidth = width
        user.ProfilePhoto = im["url"].(string)
      }
    }
    user.Name = userData["id"].(string)
    user.Username = userData["email"].(string)
  } else {
    return ProviderId, user, fmt.Errorf("Could not get user data: %s %s", rsp.StatusCode, responseRaw)
  }

  return ProviderId, user, nil
}
