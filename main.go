package main

import (
  "fmt"
  "time"
  "strconv"
  "strings"
  "net/url"
  "net/http"
  "database/sql"
  "github.com/gin-gonic/gin"
  "github.com/gin-contrib/location"
  "github.com/fatih/color"
  "github.com/dgrijalva/jwt-go"
  stock "github.com/JominJun/goStock/module"
)

// Claims for JWT login
type Claims struct {
  ID string
  NAME string
  jwt.StandardClaims
}

// JwtKey is JWT Key
var JwtKey = []byte("JWT_SECRET_KEY")
var expirationTime = 5 * time.Minute

// GetJwtToken gets JWT Token
func (user *Claims) GetJwtToken() (string, error) {
  expirationTime := time.Now().Add(expirationTime)
  start := time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
  end, _ := time.Parse("2006-01-02 15:04:05", time.Now().UTC().Format("2006-01-02 15:04:05"))
  t := end.Sub(start)
  hour, _ := strconv.Atoi(strings.Split(t.String(), "h")[0])
  minute, _ := strconv.Atoi(strings.Split(strings.Split(t.String(), "m")[0], "h")[1])
  second, _ := strconv.Atoi(strings.Split(strings.Split(t.String(), "m")[1], "s")[0])
  iat := hour*3600 + minute*60 + second
  claims := &Claims{
    ID: user.ID,
    NAME: user.NAME,
    StandardClaims: jwt.StandardClaims{
      IssuedAt: int64(iat),
      ExpiresAt: expirationTime.Unix(),
    },
  }

  token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
  tokenString, err := token.SignedString(JwtKey)

  if err != nil {
    return "", fmt.Errorf("token signed Error")
  }
  
  return tokenString, nil
}

func main() {
  db := stock.ConnectToDB(stock.MySQLInfo)
  defer db.Close()
  stock.CheckErr(db.Ping())

  runRouter(db)
}

func checkSubdomain(url *url.URL) (subdomain string, isSubdomain bool) {
  subdomain = ""
  isSubdomain = false

  if strings.Count(url.String(), ".") > 0 {
    // ~~.localhost 라면
    subdomain = strings.Split(url.Host, ".")[0]
    isSubdomain = true
  }

  return subdomain, isSubdomain
}

func runRouter(dbInfo *sql.DB) {
  var router *gin.Engine
  router = gin.Default()
  router.LoadHTMLGlob("templates/*")
  router.Use(location.Default())

  router.GET("/", func(c *gin.Context) {
    _, isSubdomain := checkSubdomain(location.Get(c))
    if !isSubdomain {
      c.HTML(http.StatusOK, "index.html", gin.H{})
    } else {
      c.JSON(http.StatusNotFound, gin.H{"status": http.StatusNotFound, "error": "Not Found"})
    }
  })

  router.POST("/auth", func(c *gin.Context) {
    subdomain, _ := checkSubdomain(location.Get(c))

    if subdomain == "api" {
      id := c.Query("id")
      pw := c.Query("pw")

      query := fmt.Sprintf("SELECT COUNT(*) as count FROM public.user WHERE id='%s' AND pw='%s'",
      id, stock.HashPW(pw))
      rows, err := dbInfo.Query(query)
      stock.CheckErr(err)

      if stock.CountRows(rows) > 0 {  
        name := stock.GetUserInfo(dbInfo, id).Name 
        user := Claims{ID: id, NAME: name}
        jwtToken, err := user.GetJwtToken()
        stock.CheckErr(err)

        c.Header("Cache-Control", "no-store, no-cache, must-revalidate, post-check=0, pre-check=0, max-age=0")
        c.Header("Last-Modified", time.Now().String())
        c.Header("Pragma", "no-cache")
        c.SetCookie("access-token", jwtToken, 1800, "", "", false, false)

        c.JSON(http.StatusOK, gin.H{"status": http.StatusOK})
        color.Green(fmt.Sprintf("[로그인] %s(%s)", name, id))
      } else {
        c.SetCookie("access-token", "", -1, "", "", false, false) // 이전에 있던 cookie가 그대로 유지되어서 넣어줌
        c.JSON(http.StatusUnauthorized, gin.H{"status": http.StatusUnauthorized, "error": "Authentication failed"})
        color.Red("잘못된 ID 또는 PW입니다")
      }
    } else {
      c.JSON(http.StatusNotFound, gin.H{"status": http.StatusNotFound, "error": "Not Found"})
    }
  })

  router.Run(":8080")
}