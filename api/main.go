package main

import (
	"net/http"
	"fmt"
	"time"
	"strings"
	"strconv"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"crypto/sha256"
	"github.com/fatih/color"
	"github.com/gin-gonic/gin"
	"github.com/dgrijalva/jwt-go"
	_ "github.com/lib/pq"
)

// AuthInp - Auth API 입력 값
type AuthInp struct {
	ID	string
	PW	string
}

// AuthOup - Auth API 반환 값
type AuthOup struct {
	ID		string
	Name	string
	jwt.StandardClaims
}

// GetJwtToken gets JWT Token
func (user *AuthOup) GetJwtToken() (string, error) {
	expirationTime := time.Now().Add(30 * time.Minute)
	start := time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
	end, _ := time.Parse("2006-01-02 15:04:05", time.Now().UTC().Format("2006-01-02 15:04:05"))
	t := end.Sub(start)
	hour, _ := strconv.Atoi(strings.Split(t.String(), "h")[0])
	minute, _ := strconv.Atoi(strings.Split(strings.Split(t.String(), "m")[0], "h")[1])
	second, _ := strconv.Atoi(strings.Split(strings.Split(t.String(), "m")[1], "s")[0])
	iat := hour*3600 + minute*60 + second

	authInp := &AuthOup{
	  ID: user.ID,
	  Name: user.Name,
	  StandardClaims: jwt.StandardClaims{
		IssuedAt: int64(iat),
		ExpiresAt: expirationTime.Unix(),
	  },
	}
  
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, authInp)
	JwtKey := []byte("JWT_SECRET_KEY")
	tokenString, err := token.SignedString(JwtKey)
  
	if err != nil {
	  return "", fmt.Errorf("token signed Error")
	}
	
	return tokenString, nil
}

func main() {
	var r *gin.Engine
	r = gin.Default()

	db := ConnectToDB()
	CheckErr(db.Ping())
	defer db.Close()
	
  	r.POST("v1/auth/", func(c *gin.Context) {
		auth := c.Request.Header["Authorization"][0]
		inp := AuthInp{}
		json.Unmarshal([]byte(auth), &inp)
		
		// 시간 설정
		now, _ := time.Parse("2006-01-02 15:04:05", time.Now().UTC().Format("2006-01-02 15:04:05"))
		t := now.Sub(time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC))
		hour, _ := strconv.Atoi(strings.Split(t.String(), "h")[0])
		minute, _ := strconv.Atoi(strings.Split(strings.Split(t.String(), "m")[0], "h")[1])
		second, _ := strconv.Atoi(strings.Split(strings.Split(t.String(), "m")[1], "s")[0])
		iat := int64(hour * 3600 + minute * 60 + second)
		exp := time.Now().Add(30 * time.Minute).Unix()

		// PW 암호화
		hash := sha256.New()
		hash.Write([]byte(inp.PW))
		hashPW := hex.EncodeToString(hash.Sum(nil))

		// DB 처리
		query := fmt.Sprintf("SELECT COUNT(*) as count FROM public.user WHERE id='%s' AND pw='%s'",
		inp.ID, hashPW)
		rows, err := db.Query(query)
		CheckErr(err)

		fmt.Println(inp.ID, inp.PW)
		  
		if CountRows(rows) == 1 {
			query := fmt.Sprintf("SELECT id, name FROM public.user WHERE id='%s'", inp.ID)
			rows, err := db.Query(query)
			CheckErr(err)

			var oup AuthOup

			for rows.Next() {
				err := rows.Scan(&oup.ID, &oup.Name)
				CheckErr(err)
			}
			
			oup = AuthOup{
				StandardClaims: jwt.StandardClaims{
					IssuedAt: iat,
					ExpiresAt: exp,
				},
			}

			c.JSON(200, gin.H{
				"status": http.StatusOK,
				"identity": oup,
			})
		} else {
			c.JSON(401, gin.H{
				"status": http.StatusUnauthorized,
				"error": "Login Failed",
			})
		}
	})
	
  	r.Run(":8081") // listen and serve on 0.0.0.0:8080
}

// CheckErr : 에러 체크
func CheckErr (err error) {
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