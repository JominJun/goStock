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
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/location"
	_ "github.com/lib/pq"
)

// AuthInp - Auth API 입력 값
type AuthInp struct {
	ID	string
	PW	string
	Name string
}

// AuthOup - Auth API 반환 값
type AuthOup struct {
	ID		string
	Name	string
	Money	int
	IsAdmin	bool
	jwt.StandardClaims
}

// Company - 회사 API 반환 값
type Company struct {
	Seq         int
	Name        string
	Description string
	StockValue  int
}

// MyStock - 보유 주식 API 반환 값
type MyStock struct {
	Name    	string
	Number  	int
	Profit   	int
	TradedValue	int
}

var domain = "api.localhost:8081"

// MiddleWare - 미들웨어
func MiddleWare(c *gin.Context) {
	if CheckSubdomain(location.Get(c), "api") {
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

	r.Use(location.Default())
	r.Use(MiddleWare)

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"*"},
		AllowHeaders:     []string{"*"},
		AllowCredentials: true,
	}))

	// 405 SET
	r.NoMethod(func(c *gin.Context) {
		c.JSON(405, gin.H{
			"status": http.StatusMethodNotAllowed,
			"message": "This Method's Not Supported",
		})
	})

	// 404 SET
	r.NoRoute(func(c *gin.Context) {
		c.JSON(404, gin.H{
			"status": http.StatusNotFound,
			"message": "Page Not Found",
		})
	})

	db := ConnectToDB()
	CheckErr(db.Ping())
	defer db.Close()

	JwtKey := []byte("JWT_SECRET_KEY")

	v1 := r.Group("/v1")
	{
		// v1/auth/register => REGISTER
		v1.POST("/auth/register", func(c *gin.Context) {
			if CheckSubdomain(location.Get(c), "api") {
				var inp AuthInp

				if len(c.Request.Header["Authentication"]) == 0 {
					c.JSON(400 , gin.H{
						"status": http.StatusBadRequest,
						"message": "Authentication Needed. But missing.",
					})
				} else {
					auth := c.Request.Header["Authentication"][0]
					json.Unmarshal([]byte(auth), &inp)

					// DB 처리
					query := fmt.Sprintf("SELECT COUNT(*) as count FROM public.user WHERE id='%s' OR name='%s'", inp.ID, inp.Name)
					rows, err := db.Query(query)
					CheckErr(err)

					// PW 암호화
					hash := sha256.New()
					hash.Write([]byte(inp.PW))
					hashPW := hex.EncodeToString(hash.Sum(nil))
					
					if CountRows(rows) == 0 {
						if inp.ID != "" && hashPW != "" && inp.Name != "" {
							t := time.Now()
							query := fmt.Sprintf("INSERT INTO public.user(id, pw, name, money, register_date, is_admin) VALUES('%s', '%s', '%s', %d, '%d%d%d%d%d', %t)",
							inp.ID, hashPW, inp.Name, 50000, t.Year(), int(t.Month()), t.Day(), t.Hour(), t.Minute(), false)
							_, err := db.Query(query)
							CheckErr(err)

							c.JSON(201, gin.H{
								"status": http.StatusCreated,
								"message": "Successfully Created",
							})
						} else {
							c.JSON(400, gin.H{
								"status": http.StatusBadRequest,
								"message": "ID, PW, Name Needed. But something's missing",
							})
						}
					} else {
						c.JSON(409, gin.H{
							"status": http.StatusConflict,
							"message": "Already Exists. ID and Name should not be overlapped",
						})
					}
				}
			}
		})
		
		// v1/auth/login => LOGIN
		v1.POST("/auth/login", func(c *gin.Context) {
			if CheckSubdomain(location.Get(c), "api") {
				var inp AuthInp
				var oup AuthOup

				if c.Query("id") == "" && c.Query("pw") == "" {
					c.JSON(400 , gin.H{
						"status": http.StatusBadRequest,
						"message": "id and pw Needed. But something's missing.",
					})
				} else {
					inp.ID = c.Query("id")
					inp.PW = c.Query("pw")
					
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
						query := fmt.Sprintf("SELECT id, name, money, is_admin FROM public.user WHERE id='%s'", inp.ID)
						rows, err := db.Query(query)
						CheckErr(err)

						for rows.Next() {
							err := rows.Scan(&oup.ID, &oup.Name, &oup.Money, &oup.IsAdmin)
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
							"id": inp.ID,
							"pw": inp.PW,
						})
					}
				}
			} else {
				c.JSON(404, gin.H{
					"status": http.StatusNotFound,
					"message": "Page Not Found",
				})
			}
		})

		// v1/auth/info => My Info
		v1.GET("/auth/info", func(c *gin.Context) {
			if CheckSubdomain(location.Get(c), "api") {
				if len(c.Request.Header["Authorization"]) == 0 {
					c.JSON(400 , gin.H{
						"status": http.StatusBadRequest,
						"message": "Authorization Needed. But missing.",
					})
				} else {
					auth := c.Request.Header["Authorization"][0]

					claims := jwt.MapClaims{}
					_, err := jwt.ParseWithClaims(auth, claims, func(token *jwt.Token) (interface{}, error) {
						return JwtKey, nil
					})

					if err != nil {
						c.JSON(403, gin.H{
							"status": http.StatusForbidden,
							"message": "Token is Expired.",
						})
					} else {
						isValid := false
						id := ""

						for key, val := range claims {
							if key == "ID" {
								query := fmt.Sprintf("SELECT COUNT(*) as count FROM public.user WHERE id='%s'", val)
								rows, err := db.Query(query)
								CheckErr(err)

								id = fmt.Sprintf("%s", val)

								if CountRows(rows) == 1 {
									isValid = true
								}
							}
						}

						if isValid {
							// DB 처리
							query := fmt.Sprintf("SELECT id, name, money, is_admin FROM public.user WHERE id='%s'", id)
							rows, err := db.Query(query)
							CheckErr(err)

							var result AuthOup

							for rows.Next() {
								rows.Scan(&result.ID, &result.Name, &result.Money, &result.IsAdmin)
							}

							c.JSON(http.StatusOK, gin.H{
								"status": http.StatusOK,
								"result": result,
							})
						} else {
							c.JSON(403, gin.H{
								"status": http.StatusForbidden,
								"message": "Token is Expired.",
							})
						}
					}
				}
			} else {
				c.JSON(404, gin.H{
					"status": http.StatusNotFound,
					"message": "Page Not Found",
				})
			}
		})

		// v1/auth/validate => Check JWT
		v1.GET("/auth/validate", func(c *gin.Context) {
			if CheckSubdomain(location.Get(c), "api") {
				if len(c.Request.Header["Authorization"]) == 0 {
					c.JSON(400 , gin.H{
						"status": http.StatusBadRequest,
						"message": "Authorization Needed. But missing.",
					})
				} else {
					auth := c.Request.Header["Authorization"][0]

					claims := jwt.MapClaims{}
					_, err := jwt.ParseWithClaims(auth, claims, func(token *jwt.Token) (interface{}, error) {
						return JwtKey, nil
					})

					if err != nil {
						c.JSON(403, gin.H{
							"status": http.StatusForbidden,
							"message": "Invalid Token",
						})
					} else {
						isValid := false

						for key, val := range claims {
							if key == "ID" {
								query := fmt.Sprintf("SELECT COUNT(*) as count FROM public.user WHERE id='%s'", val)
								rows, err := db.Query(query)
								CheckErr(err)

								if CountRows(rows) == 1 {
									isValid = true
								}
							}
						}

						if isValid {
							c.JSON(200, gin.H{
								"status": http.StatusOK,
								"message": "Valid Token",
							})
						}
					}
				}
			}
		})

		// v1/company => Company 전체조회
		v1.GET("/company", func(c *gin.Context) {
			if CheckSubdomain(location.Get(c), "api") {
				if len(c.Request.Header["Authorization"]) == 0 {
					c.JSON(400 , gin.H{
						"status": http.StatusBadRequest,
						"message": "Authorization Needed. But missing.",
					})
				} else {
					auth := c.Request.Header["Authorization"][0]

					claims := jwt.MapClaims{}
					_, err := jwt.ParseWithClaims(auth, claims, func(token *jwt.Token) (interface{}, error) {
						return JwtKey, nil
					})

					if err != nil {
						c.JSON(403, gin.H{
							"status": http.StatusForbidden,
							"message": "Token is Expired.",
						})
					} else {
						isValid := false

						for key, val := range claims {
							if key == "ID" {
								query := fmt.Sprintf("SELECT COUNT(*) as count FROM public.user WHERE id='%s'", val)
								rows, err := db.Query(query)
								CheckErr(err)

								if CountRows(rows) == 1 {
									isValid = true
								}
							}
						}

						if isValid {
							// DB 처리
							query := fmt.Sprintf("SELECT seq, name, description FROM public.company")
							rows, err := db.Query(query)
							CheckErr(err)

							var companyList = []Company{}

							for rows.Next() {
								var c Company
								errScan := rows.Scan(&c.Seq, &c.Name, &c.Description)
								CheckErr(errScan)

								query2 := fmt.Sprintf("SELECT value FROM %s ORDER BY seq DESC LIMIT 1", c.Name)
								rows2, err2 := db.Query(query2)
								CheckErr(err2)

								for rows2.Next() {
									errScan2 := rows2.Scan(&c.StockValue)
									CheckErr(errScan2)
								}

								companyList = append(companyList, c)
							}

							c.JSON(http.StatusOK, gin.H{
								"status": http.StatusOK,
								"result": companyList,
							})
						} else {
							c.JSON(403, gin.H{
								"status": http.StatusForbidden,
								"message": "Token is Expired.",
							})
						}
					}
				}
			} else {
				c.JSON(404, gin.H{
					"status": http.StatusNotFound,
					"message": "Page Not Found",
				})
			}
		})

		// v1/company/:name => Company 조건조회
		v1.GET("/company/:name", func(c *gin.Context) {
			if CheckSubdomain(location.Get(c), "api") {
				if len(c.Request.Header["Authorization"]) == 0 {
					c.JSON(400 , gin.H{
						"status": http.StatusBadRequest,
						"message": "Authorization Needed. But missing.",
					})
				} else {
					auth := c.Request.Header["Authorization"][0]

					claims := jwt.MapClaims{}
					_, err := jwt.ParseWithClaims(auth, claims, func(token *jwt.Token) (interface{}, error) {
						return JwtKey, nil
					})

					if err != nil {
						c.JSON(401, gin.H{
							"status": http.StatusUnauthorized,
							"message": "Token is Expired.",
						})
					} else {
						isValid := false

						for key, val := range claims {
							if key == "ID" {
								query := fmt.Sprintf("SELECT COUNT(*) as count FROM public.user WHERE id='%s'", val)
								rows, err := db.Query(query)
								CheckErr(err)

								if CountRows(rows) == 1 {
									isValid = true
								}
							}
						}

						if isValid {
							name := c.Param("name")
							query := fmt.Sprintf("SELECT COUNT(*) as count FROM public.company WHERE name='%s'", name)
							rows, err := db.Query(query)
							CheckErr(err)

							if CountRows(rows) == 1 {
								// DB 처리
								query2 := fmt.Sprintf("SELECT seq, name, description FROM public.company WHERE name='%s'", name)
								rows2, err2 := db.Query(query2)
								CheckErr(err2)

								var companyList = []Company{}

								for rows2.Next() {
									var c Company
									errScan := rows2.Scan(&c.Seq, &c.Name, &c.Description)
									CheckErr(errScan)

									query3 := fmt.Sprintf("SELECT value FROM %s ORDER BY seq DESC LIMIT 1", c.Name)
									rows3, err3 := db.Query(query3)
									CheckErr(err3)

									for rows3.Next() {
										errScan2 := rows3.Scan(&c.StockValue)
										CheckErr(errScan2)
									}

									companyList = append(companyList, c)
								}

								c.JSON(http.StatusOK, gin.H{
									"status": http.StatusOK,
									"result": companyList,
								})
							} else {
								c.JSON(400, gin.H{
									"status": http.StatusBadRequest,
									"message": fmt.Sprintf("No Company Named '%s'", name),
								})
							}
						} else {
							c.JSON(403, gin.H{
								"status": http.StatusForbidden,
								"message": "Forbidden. Token is invalid.",
							})
						}
					}
				}
			} else {
				c.JSON(404, gin.H{
					"status": http.StatusNotFound,
					"message": "Page Not Found",
				})
			}
		})

		// v1/company => Company 추가 [[ADMIN]]
		v1.POST("/company", func(c *gin.Context) {
			if CheckSubdomain(location.Get(c), "api") {
				if len(c.Request.Header["Authorization"]) == 0 {
					c.JSON(400 , gin.H{
						"status": http.StatusBadRequest,
						"message": "Authorization Needed. But missing.",
					})
				} else {
					auth := c.Request.Header["Authorization"][0]

					claims := jwt.MapClaims{}
					_, err := jwt.ParseWithClaims(auth, claims, func(token *jwt.Token) (interface{}, error) {
						return JwtKey, nil
					})

					if err != nil {
						c.JSON(403, gin.H{
							"status": http.StatusForbidden,
							"message": "Token is Expired.",
						})
					} else {
						isAdmin := false

						for key, val := range claims {
							if key == "ID" {
								query := fmt.Sprintf("SELECT COUNT(*) as count FROM public.user WHERE id='%s' AND is_admin=true", val)
								rows, err := db.Query(query)
								CheckErr(err)

								if CountRows(rows) == 1 {
									isAdmin = true
								}
							}
						}

						if isAdmin {
							name := c.Query("name")
							description := c.Query("description")

							if name != "" && description != "" {
								// DB 처리
								query := fmt.Sprintf("SELECT COUNT(*) as count FROM public.company WHERE name='%s'", name)
								rows, err := db.Query(query)
								CheckErr(err)

								if CountRows(rows) == 0 {
									query2 := fmt.Sprintf("INSERT INTO public.company(name, description) VALUES('%s', '%s')", name, description)
									_, err2 := db.Query(query2)
									CheckErr(err2)

									query3 := fmt.Sprintf("CREATE TABLE %s(seq integer, value integer NOT NULL, date text NOT NULL, PRIMARY KEY (seq))", name)
									_, err3 := db.Exec(query3)
									CheckErr(err3)

									query4 := fmt.Sprintf("COMMENT ON TABLE %s IS '%s'", name, description)
									_, err4 := db.Exec(query4)
									CheckErr(err4)

									query5 := fmt.Sprintf("ALTER TABLE %s ALTER COLUMN seq ADD GENERATED ALWAYS AS IDENTITY", name)
									_, err5 := db.Exec(query5)
									CheckErr(err5)

									c.JSON(200, gin.H{
										"status": http.StatusOK,
										"message": "Successfully Added.",
									})
								} else {
									c.JSON(409, gin.H{
										"status": http.StatusConflict,
										"message": "Already Exists.",
									})
								}
							} else {
								c.JSON(400, gin.H{
									"status": http.StatusBadRequest,
									"message": "name, description needed. But something's missing.",
								})
							}
						} else {
							c.JSON(401, gin.H{
								"status": http.StatusUnauthorized,
								"message": "Admin Auth Needed.",
							})
						}
					}
				}
			} else {
				c.JSON(404, gin.H{
					"status": http.StatusNotFound,
					"message": "Page Not Found",
				})
			}
		})

		// v1/company => Company 수정 [[ADMIN]]
		v1.PUT("/company", func(c *gin.Context) {
			if CheckSubdomain(location.Get(c), "api") {
				if len(c.Request.Header["Authorization"]) == 0 {
					c.JSON(400 , gin.H{
						"status": http.StatusBadRequest,
						"message": "Authorization Needed. But missing.",
					})
				} else {
					auth := c.Request.Header["Authorization"][0]

					claims := jwt.MapClaims{}
					_, err := jwt.ParseWithClaims(auth, claims, func(token *jwt.Token) (interface{}, error) {
						return JwtKey, nil
					})

					if err != nil {
						c.JSON(403, gin.H{
							"status": http.StatusForbidden,
							"message": "Token is Expired.",
						})
					} else {
						isAdmin := false

						for key, val := range claims {
							if key == "ID" {
								query := fmt.Sprintf("SELECT COUNT(*) as count FROM public.user WHERE id='%s' AND is_admin=true", val)
								rows, err := db.Query(query)
								CheckErr(err)

								if CountRows(rows) == 1 {
									isAdmin = true
								}
							}
						}

						if isAdmin {
							name := c.Query("name")
							description := c.Query("description")

							if name != "" && description != "" {
								// DB 처리
								query := fmt.Sprintf("SELECT COUNT(*) as count FROM public.company WHERE name='%s'", name)
								rows, err := db.Query(query)
								CheckErr(err)

								if CountRows(rows) == 1 {
									query2 := fmt.Sprintf("UPDATE public.company SET description='%s' WHERE name='%s'", description, name)
									_, err2 := db.Query(query2)
									CheckErr(err2)

									c.JSON(200, gin.H{
										"status": http.StatusOK,
										"message": "Successfully Modified.",
									})
								} else {
									c.JSON(400, gin.H{
										"status": http.StatusBadRequest,
										"message": fmt.Sprintf("No Company Named '%s'", name),
									})
								}
							} else {
								c.JSON(400, gin.H{
									"status": http.StatusBadRequest,
									"message": "name, description needed. But something's missing.",
								})
							}
						} else {
							c.JSON(401, gin.H{
								"status": http.StatusUnauthorized,
								"message": "Admin Auth Needed.",
							})
						}
					}
				}
			}
		})

		// v1/company => Company 파산 [[ADMIN]]
		v1.DELETE("/company", func(c *gin.Context) {
			if CheckSubdomain(location.Get(c), "api") {
				if len(c.Request.Header["Authorization"]) == 0 {
					c.JSON(400 , gin.H{
						"status": http.StatusBadRequest,
						"message": "Authorization Needed. But missing.",
					})
				} else {
					auth := c.Request.Header["Authorization"][0]

					claims := jwt.MapClaims{}
					_, err := jwt.ParseWithClaims(auth, claims, func(token *jwt.Token) (interface{}, error) {
						return JwtKey, nil
					})

					if err != nil {
						c.JSON(403, gin.H{
							"status": http.StatusForbidden,
							"message": "Token is Expired.",
						})
					} else {
						isAdmin := false

						for key, val := range claims {
							if key == "ID" {
								query := fmt.Sprintf("SELECT COUNT(*) as count FROM public.user WHERE id='%s' AND is_admin=true", val)
								rows, err := db.Query(query)
								CheckErr(err)

								if CountRows(rows) == 1 {
									isAdmin = true
								}
							}
						}

						if isAdmin {
							name := c.Query("name")

							if name != "" {
								// DB 처리
								query := fmt.Sprintf("SELECT COUNT(*) as count FROM public.company WHERE name='%s'", name)
								rows, err := db.Query(query)
								CheckErr(err)

								if CountRows(rows) == 1 {
									query2 := fmt.Sprintf("DELETE FROM public.company WHERE name='%s'", name)
									_, err2 := db.Query(query2)
									CheckErr(err2)

									query3 := fmt.Sprintf("DROP TABLE IF EXISTS %s", name)
									_, err3 := db.Query(query3)
									CheckErr(err3)

									c.JSON(200, gin.H{
										"status": http.StatusOK,
										"message": "Successfully Bankrupted.",
									})
								} else {
									c.JSON(400, gin.H{
										"status": http.StatusBadRequest,
										"message": fmt.Sprintf("No Company Named '%s'", name),
									})
								}
							} else {
								c.JSON(400, gin.H{
									"status": http.StatusBadRequest,
									"message": "name, description needed. But something's missing.",
								})
							}
						} else {
							c.JSON(401, gin.H{
								"status": http.StatusUnauthorized,
								"message": "Admin Auth Needed.",
							})
						}
					}
				}
			}
		})
		
		// v1/stocks => 내 Stock 조회
		v1.GET("/stocks", func(c *gin.Context) {
			if CheckSubdomain(location.Get(c), "api") {
				if len(c.Request.Header["Authorization"]) == 0 {
					c.JSON(400 , gin.H{
						"status": http.StatusBadRequest,
						"message": "Authorization Needed. But missing.",
					})
				} else {
					auth := c.Request.Header["Authorization"][0]

					claims := jwt.MapClaims{}
					_, err := jwt.ParseWithClaims(auth, claims, func(token *jwt.Token) (interface{}, error) {
						return JwtKey, nil
					})

					if err != nil {
						c.JSON(403, gin.H{
							"status": http.StatusForbidden,
							"message": "Token is Expired.",
						})
					} else {
						var id string
						isValid := false

						for key, val := range claims {
							if key == "ID" {
								query := fmt.Sprintf("SELECT COUNT(*) as count FROM public.user WHERE id='%s'", val)
								rows, err := db.Query(query)
								CheckErr(err)

								if CountRows(rows) == 1 {
									isValid = true
									id = fmt.Sprintf("%s", val)
								}
							}
						}

						if isValid {
							query := fmt.Sprintf("SELECT company_name, number, traded_value FROM stocks WHERE trader_id='%s'", id)
							rows, err := db.Query(query)
							CheckErr(err)

							result := []MyStock{}

							for rows.Next() {
								var m MyStock
								rows.Scan(&m.Name, &m.Number, &m.TradedValue)

								query2 := fmt.Sprintf("SELECT value FROM %s ORDER BY seq DESC LIMIT 1", m.Name)
								rows2, err2 := db.Query(query2)
								CheckErr(err2)

								for rows2.Next() {
									var nowValue int
									rows2.Scan(&nowValue)

									m.Profit = (nowValue - m.TradedValue) * m.Number
								}

								result = append(result, m)
							}

							c.JSON(200, gin.H{
								"status": http.StatusOK,
								"result": result,
							})
						} else {
							c.JSON(403, gin.H{
								"status": http.StatusUnauthorized,
								"message": "Forbidden. Token is invalid.",
							})
						}
					}
				}
			} else {
				c.JSON(404, gin.H{
					"status": http.StatusNotFound,
					"message": "Page Not Found",
				})
			}
		})

		// v1/stocks => Stock 구매
		v1.POST("/stocks/:name/:number", func(c *gin.Context) {
			if CheckSubdomain(location.Get(c), "api") {
				if len(c.Request.Header["Authorization"]) == 0 {
					c.JSON(400 , gin.H{
						"status": http.StatusBadRequest,
						"message": "Authorization Needed. But missing.",
					})
				} else {
					auth := c.Request.Header["Authorization"][0]

					claims := jwt.MapClaims{}
					_, err := jwt.ParseWithClaims(auth, claims, func(token *jwt.Token) (interface{}, error) {
						return JwtKey, nil
					})

					if err != nil {
						c.JSON(401, gin.H{
							"status": http.StatusUnauthorized,
							"message": "Token is Expired.",
						})
					} else {
						isValid := false
						id := ""

						for key, val := range claims {
							if key == "ID" {
								query := fmt.Sprintf("SELECT COUNT(*) as count FROM public.user WHERE id='%s'", val)
								rows, err := db.Query(query)
								CheckErr(err)

								id = fmt.Sprintf("%s", val)

								if CountRows(rows) == 1 {
									isValid = true
								}
							}
						}

						if isValid {
							name := c.Param("name")
							number, errParam := strconv.Atoi(c.Param("number"))

							if errParam != nil {
								c.JSON(400, gin.H{
									"status": http.StatusBadRequest,
									"message": "number should be numbers.",
								})
							} else {
								if name != "" && number != 0 {
									query2 := fmt.Sprintf("SELECT value FROM public.%s ORDER BY seq DESC LIMIT 1", name)
									rows2, err2 := db.Query(query2)
									
									if err2 != nil {
										c.JSON(400, gin.H{
											"status": http.StatusBadRequest,
											"message": fmt.Sprintf("No Company Named '%s'", name),
										})
									} else {
										var nowPrice int

										for rows2.Next() {
											rows2.Scan(&nowPrice)
										}

										query3 := fmt.Sprintf("SELECT money FROM public.user WHERE id='%s'", id)
										rows3, err3 := db.Query(query3)
										CheckErr(err3)

										var ownedMoney int

										for rows3.Next() {
											rows3.Scan(&ownedMoney)
										}

										// 보유한 돈이 주식을 사기에 충분하다면
										if nowPrice * number < ownedMoney {
											query4 := fmt.Sprintf("INSERT INTO public.stocks(company_name, number, traded_value, trader_id, date) VALUES('%s', %d, %d, '%s', '%s')",
											name, number, nowPrice, id, getNowTime())
											_, err4 := db.Query(query4)
											CheckErr(err4)

											c.JSON(200, gin.H{
												"status": http.StatusOK,
												"message": "Successfully Bought.",
											})
										} else {
											c.JSON(400, gin.H{
												"status": http.StatusBadRequest,
												"message": "More Money Needed.",
											})
										}
									}
								} else {
									c.JSON(400, gin.H{
										"status": http.StatusBadRequest,
										"message": "name, number Needed. But something's missing or zero.",
									})
								}
							}
						} else {
							c.JSON(403, gin.H{
								"status": http.StatusForbidden,
								"message": "Forbidden. Token is invalid.",
							})
						}
					}
				}
			}
		})

		// v1/stock => Stock 판매
		v1.DELETE("/stocks/:name/:number", func(c *gin.Context) {
			if CheckSubdomain(location.Get(c), "api") {
				if len(c.Request.Header["Authorization"]) == 0 {
					c.JSON(400 , gin.H{
						"status": http.StatusBadRequest,
						"message": "Authorization Needed. But missing.",
					})
				} else {
					auth := c.Request.Header["Authorization"][0]

					claims := jwt.MapClaims{}
					_, err := jwt.ParseWithClaims(auth, claims, func(token *jwt.Token) (interface{}, error) {
						return JwtKey, nil
					})

					if err != nil {
						c.JSON(401, gin.H{
							"status": http.StatusUnauthorized,
							"message": "Token is Expired.",
						})
					} else {
						isValid := false
						id := ""

						for key, val := range claims {
							if key == "ID" {
								query := fmt.Sprintf("SELECT COUNT(*) as count FROM public.user WHERE id='%s'", val)
								rows, err := db.Query(query)
								CheckErr(err)

								id = fmt.Sprintf("%s", val)

								if CountRows(rows) == 1 {
									isValid = true
								}
							}
						}

						if isValid {
							name := c.Param("name")
							number, errParam := strconv.Atoi(c.Param("number"))

							if errParam != nil {
								c.JSON(400, gin.H{
									"status": http.StatusBadRequest,
									"message": "number should be numbers.",
								})
							} else {
								if name != "" && number != 0 {
									query2 := fmt.Sprintf("SELECT number FROM public.stocks WHERE trader_id='%s' AND company_name='%s'", id, name)
									rows2, err2 := db.Query(query2)
									CheckErr(err2)

									ownedStockCount := 0
									forCnt := 0

									for rows2.Next() {
										rows2.Scan(&forCnt)
										ownedStockCount += forCnt
									}

									if ownedStockCount >= number {
										var tradedValueList []int
										processedNumber := 0

										for {
											if processedNumber == number {
												break
											}

											query3 := fmt.Sprintf("SELECT number, traded_value FROM public.stocks WHERE trader_id='%s' AND company_name='%s' ORDER BY seq DESC LIMIT 1", id, name)
											rows3, err3 := db.Query(query3)
											CheckErr(err3)

											var n, tradedValue int

											for rows3.Next() {
												rows3.Scan(&n, &tradedValue)
												tradedValueList = append(tradedValueList, tradedValue)
												
												if processedNumber + n > number {
													query4 := fmt.Sprintf("UPDATE public.stocks SET number=number-%d WHERE ctid IN (SELECT ctid FROM public.stocks WHERE trader_id='%s' AND company_name='%s' ORDER BY seq DESC LIMIT 1)",
													number-processedNumber, id, name)
													_, err4 := db.Query(query4)
													CheckErr(err4)

													processedNumber = number

													break
												} else {
													query4 := fmt.Sprintf("DELETE FROM public.stocks WHERE ctid IN (SELECT ctid FROM public.stocks WHERE trader_id='%s' AND company_name='%s' ORDER BY seq DESC LIMIT 1)",
													id, name)
													_, err4 := db.Query(query4)
													CheckErr(err4)
													
													processedNumber += n
												}
											}
										}

										query5 := fmt.Sprintf("SELECT value FROM public.%s ORDER BY seq DESC LIMIT 1", name)
										rows5, err5 := db.Query(query5)
										CheckErr(err5)

										var nowPrice int
										var totalProfit int

										for rows5.Next() {
											rows5.Scan(&nowPrice)
										}

										for value := range(tradedValueList) {
											totalProfit += nowPrice-value
										}

										query6 := fmt.Sprintf("UPDATE public.user SET money=money+%d WHERE id='%s'", totalProfit, id)
										_, err6 := db.Query(query6)
										CheckErr(err6)

										c.JSON(200, gin.H{
											"status": http.StatusOK,
											"message": "Successfully Selled.",
										})
									} else {
										c.JSON(400, gin.H{
											"status": http.StatusBadRequest,
											"message": "Does not have enough stocks to sell.",
											"ownedStockCount": ownedStockCount,
										})
									}
								} else {
									c.JSON(400, gin.H{
										"status": http.StatusBadRequest,
										"message": "name, number Needed. But something's missing or zero.",
									})
								}
							}
						} else {
							c.JSON(403, gin.H{
								"status": http.StatusForbidden,
								"message": "Forbidden. Token is invalid.",
							})
						}
					}
				}
			}
		})
	}

  	r.Run(":8081")
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

// FormatNumbers - 숫자 천단위 쉼표 포맷팅
func FormatNumbers(n int) string {
	in := strconv.FormatInt(int64(n), 10)
	numOfDigits := len(in)
	if n < 0 {
		numOfDigits-- // First character is the - sign (not a digit)
	}
	numOfCommas := (numOfDigits - 1) / 3

	out := make([]byte, len(in)+numOfCommas)
	if n < 0 {
		in, out[0] = in[1:], '-'
	}

	for i, j, k := len(in)-1, len(out)-1, 0; ; i, j = i-1, j-1 {
		out[j] = in[i]
		if i == 0 {
			return string(out)
		}
		if k++; k == 3 {
			j, k = j-1, 0
			out[j] = ','
		}
	}
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

func getNowTime() string {
	t := time.Now()
	result := fmt.Sprintf("%d%d%d%d%d", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute())
	return result
}