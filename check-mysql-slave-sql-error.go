package main

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jessevdk/go-flags"
	"github.com/kazeburo/go-mysqlflags"
	"github.com/mackerelio/checkers"
)

// Version by Makefile
var version string

type opts struct {
	mysqlflags.MyOpts
	Timeout time.Duration `long:"timeout" default:"10s" description:"Timeout to connect mysql"`
	Version bool          `short:"v" long:"version" description:"Show version"`
}

type slave struct {
	lastSQLError string `mysqlvar:"Last_SQL_Error"`
}

func main() {
	ckr := checkSlaveSQLerror()
	ckr.Name = "MySQL replica/slave SQL error"
	ckr.Exit()
}

func checkSlaveSQLerror() *checkers.Checker {
	opts := opts{}
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

	db, err := mysqlflags.OpenDB(opts.MyOpts, opts.Timeout, false)
	if err != nil {
		return checkers.Critical(fmt.Sprintf("couldn't connect DB: %v", err))
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), opts.Timeout)
	defer cancel()
	ch := make(chan error, 1)

	var slaves []slave
	go func() {
		ch <- mysqlflags.Query(db, "SHOW SLAVE STATUS").Scan(&slaves)
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

	var lastSQLErrors []string
	for _, slave := range slaves {
		if slave.lastSQLError != "" {
			lastSQLErrors = append(lastSQLErrors, slave.lastSQLError)
		}
	}

	if len(lastSQLErrors) > 0 {
		msg := strings.Join(lastSQLErrors[0:], " | ")
		return checkers.Critical("Last_SQL_Error found: " + msg)
	}
	return checkers.Ok("No Last_SQL_Error exists")

}
