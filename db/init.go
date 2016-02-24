package db

import(
  //"github.com/mattn/go-sqlite3"
  "github.com/jinzhu/gorm"
  _ "github.com/mattn/go-sqlite3"
)

func Db() (gorm.DB){
  db, _ := gorm.Open("sqlite3", "storage.db")

  return db
}
