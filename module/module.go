package module

import (
	"strconv"
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
	Name	      string
	StockValue	int
}

// MyStock is struct of my stock
type MyStock struct {
  Name    string
  Number  int
}

// UserInfo is user info
type UserInfo struct {
  ID    string
  PW    string
  Name  string
  Money int
}

// MySQLInfo is my sql info
var MySQLInfo = SQLInfo {
	Host: "localhost",
	Port: 5432,
	User: "postgres",
	Password: "비밀",
  Dbname: "goStock"}
  
// NowLoginInfo : 현재 로그인 정보
var NowLoginInfo UserInfo

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

// FormatNumbers formats number with commas
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
    NowLoginInfo = GetUserInfo(dbInfo, userInfo.ID)
    
    text := fmt.Sprintf("성공적으로 로그인하였습니다: %s(%s)", NowLoginInfo.Name, NowLoginInfo.ID)
    color.Green(text)
  } else {
    color.Red("잘못된 ID 또는 PW입니다")
  }
}

// GetUserInfo gets user info
func GetUserInfo(dbInfo *sql.DB, userID string) UserInfo {
  rows, err := dbInfo.Query(fmt.Sprintf("SELECT id, name, money FROM public.user WHERE id='%s'", userID))
  CheckErr(err)

  var result UserInfo

  for rows.Next() {
    err := rows.Scan(&result.ID, &result.Name, &result.Money)
    CheckErr(err)
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
  ownedMoney := GetUserInfo(dbInfo, trader.ID).Money
  needMoney := company.StockValue * number

  if ownedMoney > needMoney {
    query := fmt.Sprintf("INSERT INTO stocks(company_name, number, traded_value, trader_id, date) VALUES('%s', %d, %d, '%s', '%s')",
    companyName, number, company.StockValue, trader.Name, getNowTime())
    _, err := dbInfo.Query(query)
    CheckErr(err)

    query2 := fmt.Sprintf("UPDATE public.user SET money = %d WHERE id='%s'", ownedMoney - needMoney, trader.ID)
    _, err2 := dbInfo.Query(query2)
    CheckErr(err2)

    text := fmt.Sprintf("[%s] %d주 구매하였습니다\n" +
    "잔액: %s원",
    companyName, number, FormatNumbers(GetUserInfo(dbInfo, trader.ID).Money))
    color.Green(text)
  } else {
    color.Red(fmt.Sprintf("금액이 부족합니다\n" +
    "[%s] : 1주 당 %s원\n" +
    "==> %s원 더 필요", companyName, FormatNumbers(company.StockValue), FormatNumbers(needMoney - ownedMoney)))
  }
}

// SellStock : 주식 판매
func SellStock(dbInfo *sql.DB, stockToSell MyStock, userInfo UserInfo) {
  query := fmt.Sprintf("SELECT seq, company_name, number, traded_value, trader_id FROM stocks WHERE company_name='%s' AND trader_id='%s'", stockToSell.Name, userInfo.ID)
  rows, err := dbInfo.Query(query)
  CheckErr(err)

  ownedCnt := 0
  for _, myStock := range(InquiryMyStocks(dbInfo, userInfo)) { // 판매할 개수만큼 주식을 보유했는지 확인
    if myStock.Name == stockToSell.Name {
      ownedCnt += myStock.Number
    }
  }

  stockNowValue := ShowCompany(dbInfo, stockToSell.Name).StockValue
  leftStockToSell := stockToSell.Number
  gainedMoneyBySelling := 0

  if ownedCnt >= stockToSell.Number {
    for rows.Next() {
      if leftStockToSell > 0 {
        var companyName, traderID string
        var seq, number, tradedValue, selledAmount int
        rows.Scan(&seq, &companyName, &number, &tradedValue, &traderID)

        if leftStockToSell < number {
          // UPDATE 쿼리
          query := fmt.Sprintf("UPDATE stocks SET number=%d WHERE seq=%d", number - leftStockToSell, seq)
          _, err := dbInfo.Query(query)
          CheckErr(err)

          leftStockToSell = 0
          selledAmount = leftStockToSell
        } else {
          // DELETE 쿼리
          query := fmt.Sprintf("DELETE FROM stocks WHERE seq=%d", seq)
          _, err := dbInfo.Query(query)
          CheckErr(err)

          leftStockToSell -= number
          selledAmount = number
        }

        // 판매한 만큼 돈 지급하기
        query := fmt.Sprintf("UPDATE public.user SET money=money+%d WHERE id='%s'", stockNowValue * selledAmount, userInfo.ID)
        _, err := dbInfo.Query(query)
        CheckErr(err)

        gainedMoneyBySelling += stockNowValue * selledAmount
      } else {
        break
      }
    }

    text := fmt.Sprintf("[%s] %d주 판매완료\n" +
    "%d x %d = %d원 지급하였습니다", stockToSell.Name, stockToSell.Number, stockNowValue, stockToSell.Number, gainedMoneyBySelling)
    color.Green(text)
  } else {
    color.Red("보유한 주가 판매할 주보다 적습니다")
  }
}

// InquiryMyStocks : 보유 주식 전부 조회
func InquiryMyStocks(dbInfo *sql.DB, userInfo UserInfo) []MyStock {
  query := fmt.Sprintf("SELECT company_name, number FROM public.stocks WHERE trader_id='%s'", userInfo.ID)
  rows, err := dbInfo.Query(query)
  CheckErr(err)
  
  var myStocksName []string
  var myStocksCnt []int

  for rows.Next() { // row
    // 값 받기
    var stockCompany string
    var stockCnt, blockAddress int
    isBought := false
    err := rows.Scan(&stockCompany, &stockCnt)
    CheckErr(err)

    for i, v := range myStocksName {
      // 이전에 주식을 산 적이 있다면
      if v == stockCompany {
        isBought = true
        blockAddress = i
      }
    }

    if isBought {
      // 해당 칸에 구매한 만큼 추가
      myStocksCnt[blockAddress] += stockCnt
    } else {
      myStocksName = append(myStocksName, stockCompany)
      myStocksCnt = append(myStocksCnt, stockCnt)
    }
  }

  var myStocks []MyStock
  for j, stockName := range(myStocksName) {
    var myStock MyStock
    myStock.Name = stockName
    myStock.Number = myStocksCnt[j]
    myStocks = append(myStocks, myStock)
  }

  return myStocks
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