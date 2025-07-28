package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"os"

	_ "modernc.org/sqlite"
	"github.com/Asfolny/geaves"
)

func main() {
	topFs := flag.NewFlagSet("top", flag.ExitOnError)
	err := topFs.Parse(os.Args[1:])
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		usage(topFs, "<cmd>")
	}

	if (topFs.NArg() == 0) {
		usage(topFs, "<cmd>")
		os.Exit(1)
	}

	cmds := getTopCommands()
	cmd, ok := cmds[topFs.Arg(0)]
	if !ok {
		fmt.Fprintf(os.Stderr, "%s: unknown command\n", topFs.Arg(0))
		os.Exit(1)
	}

	uri, ok := os.LookupEnv("GEAVES_CONNECTION")
	if !ok {
		fmt.Fprintf(os.Stderr, "GEAVES_CONNECTION env does not exist\n")
		os.Exit(1)
	}

	db, err := sql.Open("sqlite", uri)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open sqlite connection: %s", err)
		os.Exit(1)
	}

	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start a transaction: %s", err)
		os.Exit(1)
	}

	state := state{
		cmdName: topFs.Arg(0),
		args: topFs.Args()[1:],
		queries: geaves.New(db).WithTx(tx),
	}

	err = cmd.callback(state)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)

		err = tx.Rollback()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Also failed rollback: %s\n", err)
		}

		os.Exit(1)
	}

	err = tx.Commit()

}

func usage(fl *flag.FlagSet, cmd string) {
	fmt.Fprintf(os.Stdout, "Usage of %s %s\n", os.Args[0], cmd)
	// TODO list commands first
	fl.PrintDefaults()
}
