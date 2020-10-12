package main

import (
  "fmt"
  //"reflect"
  "net/http"
  "github.com/JominJun/goStock/dbmodule"
  "github.com/gin-gonic/gin"

  // Library for DB
  _ "github.com/lib/pq"
)

var mySQLInfo = dbmodule.Info {
  Host: "localhost",
  Port: 5432,
  User: "postgres",
  Password: "비번은 비밀이에요",
  Dbname: "goStock"}

func main() {
  db := dbmodule.ConnectToDB(mySQLInfo)
  defer db.Close()
  dbmodule.CheckErr(db.Ping())

  //dbmodule.AddCompany(db, dbmodule.Company{Name: "테스트", Description: "테스트회사입니다"})

  var companyList [][]string

  for _, company := range dbmodule.ShowAllCompany(db) {
    var tempList []string
    tempList = append(tempList, company.Name)
    tempList = append(tempList, company.Description)
    companyList = append(companyList, tempList)
  }

  fmt.Println(companyList[0][0])

  // ROUTER
  router := gin.Default()
	router.LoadHTMLGlob("templates/*")
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "home.html", gin.H{
      "title": "Main website",
      "companyList": companyList,
		})
	})
  router.Run(":8080")
}