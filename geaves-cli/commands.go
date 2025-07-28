package main

import (
	"fmt"

	"github.com/Asfolny/geaves"
)

type state struct {
	cmdName string
	args []string
	queries *geaves.Queries
}

type command struct {
	name string
	description string
	callback func(state) error
}

func getTopCommands() map[string]command {
	return map[string]command{
		"generate": {
			name: "generate [type]",
			description: "Generate setup script\ntype can be one of 'setup-sql' (default), 'reset-sql' or 'goose'",
			callback: generateCommand,
		},
		"entity": {
			name: "entity <sub command>",
			description: "Manage entities using sub commands, see entity help",
			callback: entityCommand,
		},
		"attribute": {
			name: "attribute <sub command>",
			description: "Manage attributes using sub commands, see attribute help",
			callback: attributeCommand,
		},
		"link": {
			name: "link <entity> <attribute>",
			description: "Add an attribute to an entity",
			callback: linkAttributeEntityCommand,
		},
		"unlink": {
			name: "unlink <entity> <attribute>",
			description: "Remove an attribute from an entity",
			callback: unlinkAttributeEntityCommand,
		},
		"linkreq": {
			name: "linkreq <entity> <attribute>",
			description: "Add a required attribute to an entity",
			callback: linkRequiredAttributeEntityCommand,
		},
		"require": {
			name: "require <entity> <attribute>",
			description: "Make an attribute required on an entity",
			callback: requireEntityAttributeCommand,
		},
		"optional": {
			name: "optional <entity> <attribute>",
			description: "Make an attribute option on an entity",
			callback: optionalEntityAttributeCommand,
		},
		"item": {
			name: "item <sub command>",
			description: "Manage items using sub commands, see item help",
			callback: itemCommand,
		},
		"help": {
			name: "help",
			description: "Prints this message, make note that some command structures are nested and may have thier own sub help commands\nitem help\n attribute help\nentity help\n",
			callback: helpCommand,
		},
	}
}

func helpCommand(s state) (err error) {
	if len(s.args) > 0 {
		switch(s.args[0]) {
		case "generate":
			fmt.Print(`
geaves-cli generate [type]

Generate migrations a user may need, and print it to stdout

Available types
  setup-sql  - Generate only the content of +goose Up
  reset-sql  - Generate only the content of +goose Down
  goose      - Generate goose format
`)
			return
		case "entity":
			fmt.Print(`
geaves-cli entity <subcommand>

Manage entities in the system, see entity help instead
`)
			return
		case "attribute":
			fmt.Print(`
geaves-cli attribute <subcommand>

Manage attributes in the system, see attribute help instead
`)
			return
		case "item":
			fmt.Print(`
geaves-cli item <subcommand>

Manage items in the system, see item help instead
`)
			return
		case "link":
			fmt.Print(`
geaves-cli link <entity> <attribute>

Link an optional attribute to an entity; must provide entity slug and attribute slug
`)
			return
		case "unlink":
			fmt.Print(`
geaves-cli optional <entity> <attribute>

Remove a link of an attribute to an entity; must provide entity slug and attribute slug
`)
			return
		case "linkreq":
			fmt.Print(`
geaves-cli linkreq <entity> <attribute>

Link an attribute to an entity and make the attribute required; must provide entity slug and attribute slug
`)
			return
		case "require":
			fmt.Print(`
geaves-cli require <entity> <attribute>

Make an attribute required on a specific entity; must provide entity slug and attribute slug
`)
			return
		case "optional":
			fmt.Print(`
geaves-cli optional <entity> <attribute>

Make an attribute optional on a specific entity; must provide entity slug and attribute slug
`)
			return
		default:
			fmt.Println("Unknown command, usage:")
		}
	}

	fmt.Print(`
geaves-cli <command>

Available commands
  generate [type]               - Generate migrations which the user may need
  entity <subcommand>           - Entity handling, see entity help for more details
  attribute <subcommand>        - Attribute handling, see attribute help for more details
  item <subcommand>             - Item handling, see item help for more details
  link <entity> <attribute>     - Link an entity to an attribute by entity slug and attribute slug
  unlink <entity> <attribute>   - Unlink an entity to an attribute by entity slug and attribute slug
  linkreq <entity> <attribute>  - Link an entity to a required attribute by entity slug and attribute slug
  require <entity> <attribute>  - Make an attribute required on an entity by entity slug and attribute slug
  optional <entity> <attribute> - Make an attribute optional on an entity by entity slug and attribute slug
  help [command]                - Prints this message, or the help details of a command
`)
	return
}
