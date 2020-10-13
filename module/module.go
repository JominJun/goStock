package module

import (
  "fmt"
  "time"
  "database/sql"
  //"reflect"

  "github.com/fatih/color"

  // Library for DB
  _ "github.com/lib/pq"
)

// Info is info for connecting to DB
type Info struct {
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
}

// CompanyStock is struct of a company's stock
type CompanyStock struct {
	Name	string
	StockValue	int
}

// MySQLInfo is my sql info
var MySQLInfo = Info {
	Host: "localhost",
	Port: 5432,
	User: "postgres",
	Password: "비번은비밀",
	Dbname: "goStock"}

// CheckErr checks error
func CheckErr (err error) {
  if err != nil {
    panic(err)
  }
}

// ConnectToDB connects to db
func ConnectToDB (sqlInfo Info) *sql.DB {
  db, err := sql.Open("postgres",
  fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
  sqlInfo.Host, sqlInfo.Port, sqlInfo.User, sqlInfo.Password, sqlInfo.Dbname))
  CheckErr(err)
  color.Blue("성공적으로 DB에 연결하였습니다: " + sqlInfo.Dbname)

  return db
}

// AddCompany adds company
func AddCompany (dbInfo *sql.DB, companyInfo Company, initialPrice int) {
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
func BankruptCompany (dbInfo *sql.DB, companyName string) {
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
func CountRows (rows *sql.Rows) (count int) {
  for rows.Next() {
    err := rows.Scan(&count)
    CheckErr(err)
  }

  return count
}

// SetStockInfo sets stock value
func SetStockInfo (dbInfo *sql.DB, stockInfo CompanyStock) {
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

// ShowAllCompany shows all company from DB
func ShowAllCompany (dbInfo *sql.DB) []Company {
  rows, err := dbInfo.Query("SELECT * FROM company")
  CheckErr(err)

  companyList := []Company{}

  for rows.Next() {
		var c Company
    err := rows.Scan(&c.Seq, &c.Name, &c.Description)
    CheckErr(err)
    companyList = append(companyList, c)
  }
  
  return companyList
}

// GetCompanyNameList returns companyNameList
func GetCompanyNameList (dbInfo *sql.DB) (companyNameList []string) {
  for _, company := range ShowAllCompany(dbInfo) {
    companyNameList = append(companyNameList, company.Name)
  }

  return companyNameList
}