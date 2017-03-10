package main

import (
	"fmt"
	"strings"
	"github.com/jessevdk/go-flags"
	"github.com/mackerelio/checkers"
	"github.com/ziutek/mymysql/mysql"
	_ "github.com/ziutek/mymysql/native"
	"os"
)

type mysqlSetting struct {
	Host string `short:"H" long:"host" default:"localhost" description:"Hostname"`
	Port string `short:"p" long:"port" default:"3306" description:"Port"`
	User string `short:"u" long:"user" default:"root" description:"Username"`
	Pass string `short:"P" long:"password" default:"" description:"Password"`
}

type connectionOpts struct {
	mysqlSetting
}

func main() {
	ckr := checkSlaveSQLerror()
	ckr.Name = "MySQL slave SQL error"
	ckr.Exit()
}

func checkSlaveSQLerror() *checkers.Checker {
	opts := connectionOpts{}
	psr := flags.NewParser(&opts, flags.Default)
	_, err := psr.Parse()
	if err != nil {
		os.Exit(1)
	}

	db := mysql.New("tcp", "", fmt.Sprintf("%s:%s", opts.Host, opts.Port), opts.User, opts.Pass, "")
	err = db.Connect()
	if err != nil {
		return checkers.Critical("couldn't connect DB")
	}
	defer db.Close()

	rows, res, err := db.Query("SHOW SLAVE STATUS")
	if err != nil {
		return checkers.Unknown("couldn't execute query")
	}

	var lastSQLErrors []string
	idxLastSQLError := res.Map("Last_SQL_Error")
	for _, row := range rows {
		lastSQLError := row.Str(idxLastSQLError)
		if lastSQLError != "" {
			lastSQLErrors = append(lastSQLErrors, lastSQLError)
		}
	}
	if len(lastSQLErrors) > 0 {
		msg := strings.Join(lastSQLErrors[0:]," | ")
		return checkers.Critical("Last_SQL_Error found: "+msg)
	}
	return checkers.Ok("No Last_SQL_Error exists")

}
