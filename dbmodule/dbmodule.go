package dbmodule

import (
  "fmt"
  "database/sql"
  //"reflect"

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
  fmt.Println("Successfully connected to " + sqlInfo.Dbname)

  return db
}

// AddCompany adds company
func AddCompany (dbInfo *sql.DB, companyInfo Company) {
  rows, err := dbInfo.Query("SELECT COUNT(*) as count FROM company WHERE name='" + companyInfo.Name + "'")
  CheckErr(err)

  if CountRows(rows) == 0 {
    _, err := dbInfo.Exec("INSERT INTO company(name, description) VALUES('" + companyInfo.Name + "','" + companyInfo.Description +"')")
    CheckErr(err)
    fmt.Println("성공적으로 추가되었습니다: " + companyInfo.Name + "(" + companyInfo.Description + ")")
  } else {
    fmt.Println("이미 존재하는 회사입니다: " + companyInfo.Name)
  }
}

// CountRows counts rows
func CountRows (rows *sql.Rows) (count int) {
  for rows.Next() {
    err := rows.Scan(&count)
    CheckErr(err)
  }

  return count
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