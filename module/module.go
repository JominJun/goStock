package module

import (
  "os"
  "fmt"
  "time"
  "crypto/sha256"
  "encoding/hex"
  "math/rand"
  "database/sql"
  "github.com/fatih/color"

  // Library for DB
  _ "github.com/lib/pq"
)

// SQLInfo is info for connecting to DB
type SQLInfo struct {
  Host      string
  Port      int
  User      string
  Password  string
  Dbname    string
}

// Company Struct
type Company struct {
  Seq         int
  Name        string
  Description string
  StockValue  int
}

// CompanyStock is struct of a company's stock
type CompanyStock struct {
	Name	string
	StockValue	int
}

// UserInfo is user info
type UserInfo struct {
  ID    string
  PW    string
  Name  string
}

// MySQLInfo is my sql info
var MySQLInfo = SQLInfo {
	Host: "localhost",
	Port: 5432,
	User: "postgres",
	Password: "비밀",
  Dbname: "goStock"}
  
// 현재 로그인 정보
var nowLoginInfo UserInfo

// CheckErr checks error
func CheckErr (err error) {
  if err != nil {
    panic(err)
  }
}

// Init ticker: loops
func Init(sc chan os.Signal, dbInfo *sql.DB) {
	ticker := time.NewTicker(1 * time.Minute) // 1분에 한 번씩
	go func() {
		for {
			select {
			case <-ticker.C:
				ResetStockValues(dbInfo)
			case <-sc:
				ticker.Stop()
				return
			}
		}
	}()
}

// ConnectToDB connects to db
func ConnectToDB(sqlInfo SQLInfo) *sql.DB {
  db, err := sql.Open("postgres",
  fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
  sqlInfo.Host, sqlInfo.Port, sqlInfo.User, sqlInfo.Password, sqlInfo.Dbname))
  CheckErr(err)
  color.Blue("성공적으로 DB에 연결하였습니다: " + sqlInfo.Dbname)

  return db
}

func hashPW(pw string) string {
  hash := sha256.New()
  hash.Write([]byte(pw))
  hashPW := hex.EncodeToString(hash.Sum(nil))

  return hashPW
}

// Register is for registering
func Register(dbInfo *sql.DB, userInfo UserInfo) {
  userInfo.PW = hashPW(userInfo.PW)

  rows, err := dbInfo.Query("SELECT COUNT(*) as count FROM public.user WHERE id='" + userInfo.ID + "'")
  CheckErr(err)

  if CountRows(rows) == 0 {
    query := fmt.Sprintf("INSERT INTO public.user (id, pw, name, money, register_date) VALUES('%s', '%s', '%s', %d, '%s')",
    userInfo.ID, userInfo.PW, userInfo.Name, 50000, getNowTime())
    fmt.Println(query)
    _, err2 := dbInfo.Exec(query)
    CheckErr(err2)

    text := fmt.Sprintf("성공적으로 추가되었습니다: %s(%s)", userInfo.Name, userInfo.ID)
    color.Blue(text)
  } else {
    color.Red("이미 존재하는 아이디입니다: " + userInfo.ID)
  }
}

// Login is for logining
func Login(dbInfo *sql.DB, userInfo UserInfo) {
  query := fmt.Sprintf("SELECT COUNT(*) as count FROM public.user WHERE id='%s' AND pw='%s'",
  userInfo.ID, hashPW(userInfo.PW))
  rows, err := dbInfo.Query(query)
  CheckErr(err)

  if CountRows(rows) > 0 {
    nowLoginInfo = GetUserInfo(dbInfo, userInfo.ID)
    
    text := fmt.Sprintf("성공적으로 로그인하였습니다: %s(%s)", nowLoginInfo.Name, nowLoginInfo.ID)
    color.Green(text)
  } else {
    color.Red("잘못된 ID 또는 PW입니다")
  }
}

// GetUserInfo gets user info
func GetUserInfo(dbInfo *sql.DB, userID string) UserInfo {
  rows, err := dbInfo.Query(fmt.Sprintf("SELECT id, name FROM public.user WHERE id='%s'", userID))
  CheckErr(err)

  var result UserInfo

  for rows.Next() {
    var u UserInfo
    err := rows.Scan(&u.ID, &u.Name)
    CheckErr(err)

    result = u
  }

  return result
}

// AddCompany adds company
func AddCompany(dbInfo *sql.DB, companyInfo Company, initialPrice int) {
  rows, err := dbInfo.Query("SELECT COUNT(*) as count FROM company WHERE name='" + companyInfo.Name + "'")
  CheckErr(err)

  if CountRows(rows) == 0 {
    // company 추가
    _, err := dbInfo.Exec("INSERT INTO company(name, description) VALUES('" + companyInfo.Name + "','" + companyInfo.Description +"')")
    CheckErr(err)

    // stock 테이블 만들기 및 초기 주식가격 정하기
    _, err2 := dbInfo.Exec("CREATE TABLE " + companyInfo.Name + "(seq integer, value integer NOT NULL, date text NOT NULL, PRIMARY KEY (seq))")
    CheckErr(err2)
    _, err3 := dbInfo.Exec("COMMENT ON TABLE " + companyInfo.Name + " IS '" + companyInfo.Description + "'")
    CheckErr(err3)
    _, err4 := dbInfo.Exec("ALTER TABLE " + companyInfo.Name + " ALTER COLUMN seq ADD GENERATED ALWAYS AS IDENTITY")
    CheckErr(err4)

    date := getNowTime()

    query := fmt.Sprintf("INSERT INTO %s(value, date) VALUES(%d, %s)", companyInfo.Name, initialPrice, date)
    _, err5 := dbInfo.Exec(query)
    CheckErr(err5)

    color.Blue("성공적으로 추가되었습니다: " + companyInfo.Name + "(" + companyInfo.Description + ")")
  } else {
    color.Red("이미 존재하는 회사입니다: " + companyInfo.Name)
  }
}

// BankruptCompany bankrupts company
func BankruptCompany(dbInfo *sql.DB, companyName string) {
  rows, err := dbInfo.Query("SELECT COUNT(*) as count FROM company WHERE name='" + companyName + "'")
  CheckErr(err)

  if CountRows(rows) > 0 {
    _, err := dbInfo.Exec("DELETE FROM company WHERE name='" + companyName +"'")
    CheckErr(err)

    _, err2 := dbInfo.Exec("DROP TABLE IF EXISTS " + companyName)
    CheckErr(err2)
    
    color.Blue("성공적으로 파산처리하였습니다: " + companyName)
  } else {
    color.Red("존재하지 않는 회사입니다: " + companyName)
  }
}

func getNowTime() string {
  t := time.Now()
  result := fmt.Sprintf("%d%d%d%d%d", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute())
  return result
}

// CountRows counts rows
func CountRows(rows *sql.Rows) (count int) {
  for rows.Next() {
    err := rows.Scan(&count)
    CheckErr(err)
  }

  return count
}

// SetStockInfo sets stock value
func SetStockInfo(dbInfo *sql.DB, stockInfo CompanyStock) {
  rows, err := dbInfo.Query("SELECT COUNT(*) as count FROM company WHERE name='" + stockInfo.Name + "'")
  CheckErr(err)

  if CountRows(rows) > 0 {
    date := getNowTime()
    query := fmt.Sprintf("INSERT INTO %s(value, date) VALUES(%d, %s)",
    stockInfo.Name, stockInfo.StockValue, date)
    _, err := dbInfo.Query(query)
    CheckErr(err)

    color.Blue(fmt.Sprintf("성공적으로 %s의 값을 %d로 변경하였습니다", stockInfo.Name, stockInfo.StockValue))
  } else {
    color.Red("존재하지 않는 회사입니다: " + stockInfo.Name)
  }
}

// GetRandomStockValue gets random stock value (0.5 ~ 2)
func GetRandomStockValue(stockValue int) int {
  rand.Seed(time.Now().UnixNano())
  min := int(stockValue / 2)
  max := int(stockValue * 2)
  result := rand.Intn(max - min + 1) + min
  return result
}

// ResetStockValues re-set stock values
func ResetStockValues(dbInfo *sql.DB) {
  color.Yellow(fmt.Sprintf("=====[%s]=====", getNowTime()))
  for _, company := range ShowAllCompany(dbInfo) {
    newStockValue := GetRandomStockValue(company.StockValue)
    text := fmt.Sprintf("현재가격: %d\n변경가격: %d", company.StockValue, newStockValue)
    color.Green("\n[" + company.Name + "]")
    fmt.Println(text)

    subValue := company.StockValue - newStockValue

    if company.StockValue > newStockValue {
      color.Red(fmt.Sprintf("▼ %d원 하락", subValue))
    } else if company.StockValue < newStockValue {
      color.Green(fmt.Sprintf("▲ %d원 상승", -subValue))
    } else {
      fmt.Println("- 변동 없음")
    }

    SetStockInfo(dbInfo, CompanyStock{Name: company.Name, StockValue: newStockValue})
  }
}

// PurchaseStock : 주식 구매
func PurchaseStock(dbInfo *sql.DB, companyName string, number int, trader UserInfo) {
  company := ShowCompany(dbInfo, companyName)
  fmt.Println(company)
  //query := fmt.Sprintf("INSERT INTO stocks(company_name, number, traded_value, trader_name, date) VALUES(%s, %d, %d, %s, %s)")
}

// ShowCompany shows company from DB
func ShowCompany(dbInfo *sql.DB, companyName string) Company {
  rows, err := dbInfo.Query(fmt.Sprintf("SELECT * FROM company WHERE name='%s'", companyName))
  CheckErr(err)

  var result Company

  for rows.Next() {
    var c Company
    err := rows.Scan(&c.Seq, &c.Name, &c.Description)
    CheckErr(err)

    query := fmt.Sprintf("SELECT value FROM %s ORDER BY seq DESC LIMIT 1", c.Name)
    rows2, err2 := dbInfo.Query(query)
    CheckErr(err2)

    for rows2.Next() {
      err := rows2.Scan(&c.StockValue)
      CheckErr(err)
    }

    result = c
  }

  return result
}

// ShowAllCompany shows all company from DB
func ShowAllCompany(dbInfo *sql.DB) []Company {
  rows, err := dbInfo.Query("SELECT * FROM company")
  CheckErr(err)

  companyList := []Company{}

  for rows.Next() {
		var c Company
    err := rows.Scan(&c.Seq, &c.Name, &c.Description)
    CheckErr(err)

    query := fmt.Sprintf("SELECT value FROM %s ORDER BY seq DESC LIMIT 1", c.Name)
    rows2, err2 := dbInfo.Query(query)
    CheckErr(err2)
    
    for rows2.Next() {
      err := rows2.Scan(&c.StockValue)
      CheckErr(err)
    }

    companyList = append(companyList, c)
  }
  
  return companyList
}

// GetCompanyNameList returns companyNameList
func GetCompanyNameList(dbInfo *sql.DB) (companyNameList []string) {
  for _, company := range ShowAllCompany(dbInfo) {
    companyNameList = append(companyNameList, company.Name)
  }

  return companyNameList
}