package main

import (
  //"syscall"
  "os"
  //"os/signal"
  "math/rand"
  "net/http"
  "github.com/gin-gonic/gin"
  "github.com/go-echarts/go-echarts/charts"

  md "github.com/JominJun/goStock/module"

  // Library for DB
  _ "github.com/lib/pq"
)

func main() {
  db := md.ConnectToDB(md.MySQLInfo)
  defer db.Close()
  md.CheckErr(db.Ping())

/*
  sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
  md.Init(sc, db)
	<-sc
*/

  //md.Register(db, md.UserInfo{ID: "test", PW: "test", Name: "test"})
  userInfo := md.UserInfo{ID: "test", PW: "test"}

  md.Login(db, userInfo)
  //md.PurchaseStock(db, "민준귀비", 7, md.NowLoginInfo)
  //md.ResetStockValues(db)
  myStocks := md.InquiryMyStocks(db, userInfo)
  md.ShowProfitTable(db, myStocks)
  //md.SellStock(db, md.MyStock{Name: "리브스엔터테인먼트", Number: 1}, md.NowLoginInfo)
  //fmt.Println("보유주식: ", md.InquiryMyStocks(db, userInfo))


  // router용으로 정제
  var companyList [][]string
  for _, company := range md.ShowAllCompany(db) {
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
    md.CheckErr(err)
    line.Render(f)

		c.HTML(http.StatusOK, "chart.html", gin.H{})
  })
  
  router.Run(":8080")
}