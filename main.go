package main

import (
  "fmt"
  "time"
  "net/http"
  "database/sql"
  "github.com/gin-gonic/gin"
  "github.com/fatih/color"
  "github.com/dgrijalva/jwt-go"
  stock "github.com/JominJun/goStock/module"
)

// Claims for JWT login
type Claims struct {
  ID string
  PW string
  jwt.StandardClaims
}

// JwtKey is JWT Key
var JwtKey = []byte("JWT_SECRET_KEY")
var expirationTime = 5 * time.Minute

// GetJwtToken gets JWT Token
func (user *Claims) GetJwtToken() (string, error) {
  expirationTime := time.Now().Add(expirationTime)
  claims := &Claims{
    ID: user.ID,
    PW: user.PW,
    StandardClaims: jwt.StandardClaims{
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

func runRouter(dbInfo *sql.DB) {
  var router *gin.Engine
  router = gin.Default()
  router.LoadHTMLGlob("templates/*")

  router.GET("/", func(c *gin.Context) {
    c.HTML(http.StatusOK, "index.html", gin.H{})
  })

  router.POST("/login", func(c *gin.Context) {
    id := c.Query("id")
    pw := c.Query("pw")

    query := fmt.Sprintf("SELECT COUNT(*) as count FROM public.user WHERE id='%s' AND pw='%s'",
    id, stock.HashPW(pw))
    rows, err := dbInfo.Query(query)
    stock.CheckErr(err)

    if stock.CountRows(rows) > 0 {   
      user := Claims{ID: id, PW: pw}
      jwtToken, err := user.GetJwtToken()
      stock.CheckErr(err)

      c.Header("Cache-Control", "no-store, no-cache, must-revalidate, post-check=0, pre-check=0, max-age=0")
      c.Header("Last-Modified", time.Now().String())
      c.Header("Pragma", "no-cache")
      c.SetCookie("access-token", jwtToken, 1800, "", "", false, false)

      c.JSON(http.StatusOK, gin.H{"status": http.StatusOK})
      color.Green("[로그인] ID: " + id)
    } else {
      c.JSON(http.StatusOK, gin.H{"status": http.StatusUnauthorized, "error": "Authentication failed"})
      color.Red("잘못된 ID 또는 PW입니다")
    }
  })

  router.Run(":8080")
}