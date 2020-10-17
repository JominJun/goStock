package main

import (
	"fmt"
	"net"
	"time"
	"strings"
	"strconv"
	"net/url"
  	"net/http"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"crypto/sha256"
	"github.com/fatih/color"
	"github.com/gin-gonic/gin"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-contrib/location"
	_ "github.com/lib/pq"
)

// AuthInp - Auth/logim API 입력 값
type AuthInp struct {
	ID	string
	PW	string
}

// AuthOup - Auth/login API 반환 값
type AuthOup struct {
	ID		string
	Name	string
	jwt.StandardClaims
}

var domain = "api.localhost:8081"

// MiddleWare - 미들웨어
func MiddleWare(c *gin.Context) {
	if CheckSubdomain(location.Get(c), "api") {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Cache-Control", "no-store, no-cache, must-revalidate, post-check=0, pre-check=0, max-age=0")
		c.Header("Last-Modified", time.Now().String())
		c.Header("Pragma", "no-cache")
		c.Header("Expires", "-1")

		c.Next()
	}
}

func main() {
	var r *gin.Engine
	r = gin.Default()

	// 404 SET
	r.NoRoute(func(c *gin.Context) {
		c.JSON(404, gin.H{
			"status": http.StatusNotFound,
			"message": "Page Not Found",
		})
	})

	//r.LoadHTMLGlob("templates/*")
	r.Use(location.Default())
	r.Use(MiddleWare)

	db := ConnectToDB()
	CheckErr(db.Ping())
	defer db.Close()

	// v1/auth/register => REGISTER
	r.POST("v1/auth/register", func(c *gin.Context) {
		if CheckSubdomain(location.Get(c), "api") {
			var inp AuthInp

			auth := c.Request.Header["Authentication"][0]
			json.Unmarshal([]byte(auth), &inp)

			// DB 처리
			query := fmt.Sprintf("SELECT COUNT(*) as count FROM public.user WHERE id='%s'", inp.ID)
			rows, err := db.Query(query)
			CheckErr(err)
			
			if CountRows(rows) == 0 {
				//query := fmt.Sprintf("INSERT INTO public.user() VALUES()")
				//_, err := db.Query(query)
				//CheckErr(err)

				c.JSON(201, gin.H{
					"status": http.StatusCreated,
					"message": "Successfully Created",
				})
			} else {
				c.JSON(409, gin.H{
					"status": http.StatusConflict,
					"message": "Already Exists",
				})
			}
		}
	})
	
	// v1/auth/login => LOGIN
  	r.POST("v1/auth/login", func(c *gin.Context) {
		if CheckSubdomain(location.Get(c), "api") {
			var inp AuthInp
			var oup AuthOup

			auth := c.Request.Header["Authentication"][0]
			json.Unmarshal([]byte(auth), &inp)
			
			// 시간 설정
			iat, exp := setIatExp()

			// PW 암호화
			hash := sha256.New()
			hash.Write([]byte(inp.PW))
			hashPW := hex.EncodeToString(hash.Sum(nil))

			// DB 처리
			query := fmt.Sprintf("SELECT COUNT(*) as count FROM public.user WHERE id='%s' AND pw='%s'",
			inp.ID, hashPW)
			rows, err := db.Query(query)
			CheckErr(err)
			
			if CountRows(rows) == 1 {
				query := fmt.Sprintf("SELECT id, name FROM public.user WHERE id='%s'", inp.ID)
				rows, err := db.Query(query)
				CheckErr(err)

				for rows.Next() {
					err := rows.Scan(&oup.ID, &oup.Name)
					CheckErr(err)
				}
				
				oup.StandardClaims = jwt.StandardClaims {
					Audience: GetIP(),
					Issuer: domain,
					IssuedAt: iat,
					ExpiresAt: exp,
				}

				// JWT 설정
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, oup)
				JwtKey := []byte("JWT_SECRET_KEY")
				tokenString, err := token.SignedString(JwtKey)
				CheckErr(err)

				c.JSON(200, gin.H{
					"status": http.StatusOK,
					"access_token": tokenString,
				})

				// 쿠키
				//c.SetCookie("access_token", tokenString, 1800, "", "", false, false)
			} else {
				c.JSON(401, gin.H{
					"status": http.StatusUnauthorized,
					"message": "Login Failed",
				})
			}
		} else {
			c.JSON(404, gin.H{
				"status": http.StatusNotFound,
				"message": "Page Not Found",
			})
		}
	})
	
  	r.Run(":8081") // listen and serve on 0.0.0.0:8080
}

// CheckErr - 에러 체크
func CheckErr(err error) {
	if err != nil {
	  panic(err)
	}
}

// ConnectToDB - DB 연결
func ConnectToDB() *sql.DB {
	db, err := sql.Open("postgres",
	fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
	"localhost", 5432, "postgres", "패스워드", "goStock"))
	CheckErr(err)
	color.Blue("성공적으로 DB에 연결하였습니다")
  
	return db
}

// CountRows - 결과 개수 세기
func CountRows(rows *sql.Rows) (count int) {
	for rows.Next() {
    	err := rows.Scan(&count)
    	CheckErr(err)
  	}

  	return count
}

// CheckSubdomain - 서브도메인 체크
func CheckSubdomain(url *url.URL, check string) (isSubdomain bool) {
	isSubdomain = false
  
	if strings.Count(url.String(), ".") > 0 {
	  if check == strings.Split(url.Host, ".")[0] {
		  isSubdomain = true
	  }
	}
  
	return isSubdomain
}

// GetIP - 사용자 IP 받기
func GetIP() string {
    conn, err := net.Dial("udp", "8.8.8.8:80")
	CheckErr(err)
	
    defer conn.Close()

    localAddr := conn.LocalAddr().(*net.UDPAddr)

    return localAddr.IP.String()
}

func setIatExp() (int64, int64) {
	now, _ := time.Parse("2006-01-02 15:04:05", time.Now().UTC().Format("2006-01-02 15:04:05"))
	t := now.Sub(time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC))
	hour, _ := strconv.Atoi(strings.Split(t.String(), "h")[0])
	minute, _ := strconv.Atoi(strings.Split(strings.Split(t.String(), "m")[0], "h")[1])
	second, _ := strconv.Atoi(strings.Split(strings.Split(t.String(), "m")[1], "s")[0])
	iat := int64(hour * 3600 + minute * 60 + second)
	exp := time.Now().Add(30 * time.Minute).Unix()

	return iat, exp
}