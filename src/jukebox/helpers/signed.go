package helpers

import (
  "math/rand"
  "crypto/hmac"
  "crypto/sha256"
  "strings"
  "encoding/base64"
)

// From http://stackoverflow.com/a/22892986/744180
var letters = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

func randomString(n int) []byte{
    b := make([]byte, n)
    for i := range b {
        b[i] = letters[rand.Intn(len(letters))]
    }
    return b
}

// TODO: load this from config file instead of regenerating at runtime
var appSecret = randomString(32)

func SignValue(val string) string{
  mac := hmac.New(sha256.New, appSecret)
  mac.Write([]byte(val))
  hash := base64.StdEncoding.EncodeToString(mac.Sum(nil))

  return strings.Join([]string{val,hash}, "|")
}
