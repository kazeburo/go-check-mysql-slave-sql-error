package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jessevdk/go-flags"
	"github.com/mackerelio/checkers"
)

// Version by Makefile
var version string

type mysqlSetting struct {
	Host    string        `short:"H" long:"host" default:"localhost" description:"Hostname"`
	Port    string        `short:"p" long:"port" default:"3306" description:"Port"`
	User    string        `short:"u" long:"user" default:"root" description:"Username"`
	Pass    string        `short:"P" long:"password" default:"" description:"Password"`
	Timeout time.Duration `long:"timeout" default:"10s" description:"Timeout to connect mysql"`
}

type connectionOpts struct {
	mysqlSetting
	Version bool `short:"v" long:"version" description:"Show version"`
}

func main() {
	ckr := checkSlaveSQLerror()
	ckr.Name = "MySQL replica/slave SQL error"
	ckr.Exit()
}

func checkSlaveSQLerror() *checkers.Checker {
	opts := connectionOpts{}
	psr := flags.NewParser(&opts, flags.HelpFlag|flags.PassDoubleDash)
	_, err := psr.Parse()
	if opts.Version {
		fmt.Fprintf(os.Stderr, "Version: %s\nCompiler: %s %s\n",
			version,
			runtime.Compiler,
			runtime.Version())
		os.Exit(0)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	db, err := sql.Open(
		"mysql",
		fmt.Sprintf(
			"%s:%s@tcp(%s:%s)/",
			opts.mysqlSetting.User,
			opts.mysqlSetting.Pass,
			opts.mysqlSetting.Host,
			opts.mysqlSetting.Port,
		),
	)
	if err != nil {
		return checkers.Critical(fmt.Sprintf("couldn't connect DB: %v", err))
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), opts.Timeout)
	defer cancel()
	ch := make(chan error, 1)

	var lastSQLErrors []string
	go func() {
		rows, e := db.Query("SHOW SLAVE STATUS")
		if e != nil {
			ch <- e
			return
		}
		defer rows.Close()
		cols, e := rows.Columns()
		if e != nil {
			ch <- e
			return
		}
		vals := make([]interface{}, len(cols))
		idxLastSQLError := -1
		for i, v := range cols {
			vals[i] = new(sql.RawBytes)
			if v == "Last_SQL_Error" {
				idxLastSQLError = i
			}
		}
		if idxLastSQLError < 0 {
			ch <- fmt.Errorf("Could not find Last_SQL_Error in columns")
			return
		}
		for rows.Next() {
			e = rows.Scan(vals...)
			if e != nil {
				ch <- e
				return
			}
			lastSQLError := string(*vals[idxLastSQLError].(*sql.RawBytes))
			if lastSQLError != "" {
				lastSQLErrors = append(lastSQLErrors, lastSQLError)
			}
		}
		ch <- nil
	}()

	select {
	case err = <-ch:
		// nothing
	case <-ctx.Done():
		err = fmt.Errorf("connection or query timeout")
	}

	if err != nil {
		return checkers.Critical(fmt.Sprintf("Couldn't fetch replica/slave status: %v", err))
	}

	if len(lastSQLErrors) > 0 {
		msg := strings.Join(lastSQLErrors[0:], " | ")
		return checkers.Critical("Last_SQL_Error found: " + msg)
	}
	return checkers.Ok("No Last_SQL_Error exists")

}
