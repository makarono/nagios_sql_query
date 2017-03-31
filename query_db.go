package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "gopkg.in/rana/ora.v4"
)



//global variables
var print_panic bool
var dbh string

func main() {

	// definiranje varijabli za flagove
	var (
		HostFlag          string
		UserFlag          string
		PasswordFlag      string
		ShemaFlag         string
		PortFlag          int
		TimeoutFlag       string
		QueryFlag         string
		WarningFlag       int
		CriticalFlag      int
		DbTypeFlag        string
		OracleServiceName string
		MessageFlag       string
		VerboseFlag       bool
		InverseFlag       bool
		ErrorFlag         bool
	)

	flag.StringVar(&HostFlag, "host", "", "database ip")
	flag.StringVar(&UserFlag, "user", "root", "database user")
	flag.StringVar(&PasswordFlag, "password", "xxxxxx", "user password")
	flag.StringVar(&ShemaFlag, "shema", "test", "database shema")
	flag.IntVar(&PortFlag, "port", 3306, "database port default 3306")
	flag.StringVar(&TimeoutFlag, "timeout", "20s", "database connection timeout in seconds, default 20s")
	flag.StringVar(&QueryFlag, "query", "select count(1) from dual", "database query")
	flag.IntVar(&WarningFlag, "warning", 0, "warning treshold.Query result must be bigger than warning to activate WARNING")
	flag.IntVar(&CriticalFlag, "critical", 0, "critical treshold")
	flag.StringVar(&DbTypeFlag, "dbtype", "", "database type. suported databases mysql and oracle")
	flag.StringVar(&OracleServiceName, "servicename", "oracle", "oracle database service name")
	flag.StringVar(&MessageFlag, "message", "test", "nagios status information message")
	flag.BoolVar(&VerboseFlag, "verbose", false, "get more details about request, add this flag at first palce after binary name")
	flag.BoolVar(&InverseFlag, "inverse", false, "inverts operator. If query result <= critical treshold state is CRITICAL. Warning flag is not checked in inverse run ")
	flag.BoolVar(&ErrorFlag, "error", false, "show panic error output, add this flag at first palce after binary name")

	// parsanje flagova
	flag.Parse()

	// exit if database type not specified
	if len(strings.TrimSpace(DbTypeFlag)) == 0 {
		fmt.Println("unknown db type")
		//fmt.Println("tail:", flag.Args())
		os.Exit(3)
	}

	// print all parametars
	if bool(VerboseFlag) == true {
		fmt.Println("host:", HostFlag)
		fmt.Println("user:", UserFlag)
		fmt.Println("password:", PasswordFlag)
		fmt.Println("shema:", ShemaFlag)
		fmt.Println("port:", PortFlag)
		fmt.Println("timeout:", TimeoutFlag)
		fmt.Println("query:", QueryFlag)
		fmt.Println("dbtype:", DbTypeFlag)
		fmt.Println("warning:", WarningFlag)
		fmt.Println("critical:", CriticalFlag)
		fmt.Println("message:", MessageFlag)
		fmt.Println("error:", ErrorFlag)
	}

	var (
		db  *sql.DB
		err error
	)

	//fmt.Println("panic debug:", ErrorFlag)

	//get global variables into scope
	print_panic = ErrorFlag

	var dbt string = DbTypeFlag
	switch dbt {
	case "mysql":
		//database connection
		db, err = sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=true&loc=Local&timeout=%s", UserFlag, PasswordFlag, HostFlag, PortFlag, ShemaFlag, TimeoutFlag))
		checkErr(err, "cant connect to: "+dbt)
	case "oracle":
		//database connection
		db, err = sql.Open("ora", fmt.Sprintf("%s/%s@%s:%d/%s", UserFlag, PasswordFlag, HostFlag, PortFlag, OracleServiceName))
		checkErr(err, "cant connect to oracle")

		if err = db.Ping(); err != nil {
			fmt.Printf("Error connecting to the database: %s\n", err)
			return
		}
	default:
		fmt.Println("unknown db type")
		os.Exit(3)
	}
	defer db.Close()

	//get global variables into scope
	dbh = dbt

	//timer started
	start := time.Now()
	//query
	rows, err := db.Query(QueryFlag)
	checkErr(err, "db query not executed")

	//query results
	var result int // rezutat query-a
	for rows.Next() {
		err = rows.Scan(&result)
		checkErr(err, "no query result returned from: "+dbt)
	}
	//fmt.Println(result)

	//mjerenje vremena
	elapsed := time.Since(start)

	switch {
	// invalid operation: result <= WarningFlag (mismatched types int and *int) za to sam pretvorio int(*WarningFlag)
	case bool(InverseFlag) == true && result <= int(CriticalFlag): //inverse check. If query result <= critical treshold state is CRITICAL
		fmt.Printf("CRITICAL: '%d' (%d -le %d, c=%d), exec: %s inverse check: %s \n", result, result, CriticalFlag, CriticalFlag, elapsed, MessageFlag)
		os.Exit(2)
	case result <= int(WarningFlag):
		fmt.Printf("OK: '%d' (%d -le %d, w=%d/c=%d), exec: %s: %s \n", result, result, WarningFlag, WarningFlag, CriticalFlag, elapsed, MessageFlag)
		os.Exit(0)
	case result > int(WarningFlag) && result < int(CriticalFlag):
		fmt.Printf("WARNING: '%d' (%d -gt %d, w=%d/c=%d), exec: %s: %s \n", result, result, WarningFlag, WarningFlag, CriticalFlag, elapsed, MessageFlag)
		os.Exit(1)
	case result >= int(CriticalFlag):
		fmt.Printf("CRITICAL: '%d' (%d -ge %d, w=%d/c=%d), exec: %s: %s \n", result, result, CriticalFlag, WarningFlag, CriticalFlag, elapsed, MessageFlag)
		os.Exit(2)
	case int(WarningFlag) == 0 && result >= int(CriticalFlag): //if warning flag is zero check only critical treshold , warningfalag is 0 by default if ist not specified
		fmt.Printf("CRITICAL: '%d' (%d -ge %d, c=%d), exec: %s: %s \n", result, result, CriticalFlag, CriticalFlag, elapsed, MessageFlag)
		os.Exit(2)
	default:
		fmt.Println("UNKNOWN:")
		os.Exit(3)
	}

	db.Close()
}

//error function
func checkErr(err error, text string) {
	if err != nil {
		if print_panic == true {
			panic(err)
			os.Exit(3)
		} else {
			fmt.Println("UNKNOWN:", dbh, err)
			os.Exit(3)
		}

	}
}
