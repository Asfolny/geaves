package main

import (
	"fmt"
	"strings"

	"github.com/Asfolny/geaves"
)

func generateCommand(s state) error {
	printer := "setup-sql"

	if (len(s.args) >= 1) {
		printer = s.args[0]
	}

	switch (printer) {
	case "setup-sql":
		printSetupSQL()
		return nil
	case "reset-sql":
		printResetSQL()
		return nil
	case "goose":
		printGoose()
		return nil
	default:
		return fmt.Errorf("Invalid setup type '%s', please use one of 'setup-sql' (default), 'reset-sql' or 'goose'", printer)
	}
}

func printSetupSQL() {
	fmt.Print(geaves.SetupSQL())
}

func printResetSQL() {
	fmt.Print(geaves.ResetSQL())
}

func printGoose() {
	var sb strings.Builder

	sb.WriteString("-- +goose Up\n")
	sb.WriteString(geaves.SetupSQL())
	sb.WriteString("\n")
	sb.WriteString("-- +goose Down\n")
	sb.WriteString(geaves.ResetSQL())

	fmt.Print(sb.String())
}
