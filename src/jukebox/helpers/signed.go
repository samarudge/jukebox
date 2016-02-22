package helpers

import (
  "fmt"
  "crypto/hmac"
  "crypto/sha256"
  "strings"
  "encoding/base64"
  "os"
)

var appSecret = []byte(os.Getenv("JB_SECRET"))

func getHash(val string) string{
  mac := hmac.New(sha256.New, appSecret)
  mac.Write([]byte(val))
  return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

func SignValue(val string) string{
  hash := getHash(val)

  return strings.Join([]string{base64.StdEncoding.EncodeToString([]byte(val)),hash}, "|")
}

func VerifyValue(signed string) (string, error){
  valueParts := strings.Split(signed, "|")
  val, err := base64.StdEncoding.DecodeString(valueParts[0])
  if err != nil {
    return "", fmt.Errorf("Base64 Decode Error:", err)
  }

  targetHash := getHash(string(val))
  if hmac.Equal([]byte(targetHash), []byte(valueParts[1])) {
    return string(val), nil
  } else {
    return "", fmt.Errorf("Invalid hash")
  }
}
