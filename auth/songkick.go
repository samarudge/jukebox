package auth

import(
  "fmt"
  "golang.org/x/oauth2"
  "io/ioutil"
  log "github.com/Sirupsen/logrus"
  "encoding/json"
  "net/url"
  "time"
)


type Songkick struct{
  BaseProvider
  ApiKey      string
}

func NewSongkick(p BaseProvider, additionalConfig map[interface{}]interface{}) *Songkick{
  oauth2.RegisterBrokenAuthHeaderProvider("https://www.songkick.com/oauth")
  p.Name =      "Songkick"
  p.AuthURL =   "https://www.songkick.com/oauth/login"
  p.TokenURL =  "https://www.songkick.com/oauth/exchange"
  p.ReauthEvery = time.Minute*30

  var apiKey string
  authConfig := additionalConfig["auth"].(map[interface{}]interface{})["songkick"].(map[interface{}]interface{})
  _, hasApiKey := authConfig["api_key"]
  if hasApiKey{
    apiKey = authConfig["api_key"].(string)
  } else {
    panic("Did not provide Songkick API key")
  }

  fmt.Println("API Key:", apiKey)

  return &Songkick{
    BaseProvider: p,
    ApiKey: apiKey,
  }
}

func (p *Songkick) GetUserData(token *oauth2.Token) (string, UserData, error){
  client := p.OauthClient(token)
  user := UserData{}
  var ProviderId string

  userDetailsUrl, _ := url.Parse("https://api.songkick.com/api/3.0/users/:me.json")
  q := userDetailsUrl.Query()
  q.Set("oauth_token", token.AccessToken)
  q.Set("oauth_version", "v2-10")
  q.Set("apikey", p.ApiKey)
  userDetailsUrl.RawQuery = q.Encode()

  rsp, err := client.Get(userDetailsUrl.String())
  if err != nil{
    return ProviderId, user, err
  }

  log.WithFields(log.Fields{
    "call": rsp.Request.URL,
    "status": rsp.StatusCode,
  }).Debug("User Data Songkick")

  defer rsp.Body.Close()
  responseRaw, _ := ioutil.ReadAll(rsp.Body)
  userData := make(map[string]interface{})

  if rsp.StatusCode == 200 {
    if err := json.Unmarshal(responseRaw, &userData); err != nil {
      return ProviderId, user, fmt.Errorf("Could not decode JSON", string(responseRaw))
    }

    skUserData := userData["resultsPage"].(map[string]interface{})["results"].(map[string]interface{})["user"].(map[string]interface{})

    skId := skUserData["id"].(string)
    ProviderId = p.MakeProviderId(skId)
    user.ProfilePhoto = fmt.Sprintf("https://images.sk-static.com/images/media/profile_images/users/%s/col2", skId)
    user.Name = skUserData["username"].(string)
    user.Username = user.Name
  } else {
    return ProviderId, user, fmt.Errorf("Could not get user data: %s %s", rsp.StatusCode, responseRaw)
  }

  return ProviderId, user, nil
}
