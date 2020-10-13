package main

import (
  "fmt"
  "os"
  "math/rand"
  "net/http"
  "github.com/JominJun/goStock/module"
  "github.com/gin-gonic/gin"
  "github.com/go-echarts/go-echarts/charts"

  // Library for DB
  _ "github.com/lib/pq"
)

func main() {
  db := module.ConnectToDB(module.MySQLInfo)
  defer db.Close()
  module.CheckErr(db.Ping())

  //module.BankruptCompany(db, "민준제과")
  //module.AddCompany(db, module.Company{Name: "민준귀비", Description: "마약류 제조"}, 500)
  //module.SetStockInfo(db, module.CompanyStock{Name: "sdfsdf", StockValue: 3000})
  fmt.Println(module.ShowAllCompany(db))

  var companyList [][]string
  for _, company := range module.ShowAllCompany(db) {
    var tempList []string
    tempList = append(tempList, company.Name)
    tempList = append(tempList, company.Description)
    companyList = append(companyList, tempList)
  }

  //runRouter(companyList)
}

func runRouter(companyList [][]string) {
  router := gin.Default()
  router.LoadHTMLGlob("templates/*")
  
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "home.html", gin.H{
      "title": "GoStock",
      "companyList": companyList,
		})
  })
  
  router.GET("/chart", func(c *gin.Context) {
    var nameItems []string;
    var valueItems []int;

    nameItems = append(nameItems, "A")
    nameItems = append(nameItems, "C")
    nameItems = append(nameItems, "B")

    valueItems = append(valueItems, rand.Intn(100))
    valueItems = append(valueItems, rand.Intn(100))
    valueItems = append(valueItems, rand.Intn(100))

    line := charts.NewLine()
    line.SetGlobalOptions(charts.TitleOpts{Title: "주식 차트"})
    line.AddXAxis(nameItems).
    AddYAxis("민준전자", valueItems, charts.LineOpts{Smooth: true}).
    AddYAxis("민준농업", valueItems, charts.LineOpts{Smooth: true})

    f, err := os.Create("./templates/chart.html")
    module.CheckErr(err)
    line.Render(f)

		c.HTML(http.StatusOK, "chart.html", gin.H{})
  })
  
  router.Run(":8080")
}