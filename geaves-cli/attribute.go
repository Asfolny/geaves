package main

import (
	"context"
	"flag"
	"fmt"
	"strconv"
	"strings"
	"os"

	"github.com/Asfolny/geaves"
)

func attributeCommand(s state) error {
	if len(s.args) < 1 {
		// TODO print subcommand usage instead
		return fmt.Errorf("%s requires subcommand argument", s.cmdName)
	}

	cmds := getAttributeCommands()
	cmd, ok := cmds[s.args[0]]
	if !ok {
		return fmt.Errorf("%s: attribute command not found\n", s.args[0])
	}

	return  cmd.callback(state{s.args[0], s.args[1:], s.queries})
}

func getAttributeCommands() map[string]command {
	return map[string]command{
		"create": {
			name: "attribute create <flags>",
			description: "Create a new attribute",
			callback: createAttributeCommand,
		},
		"update": {
			name: "attribute update <flags> <slug|id>",
			description: "Update an attribute by id or slug",
			callback: updateAttributeCommand,
		},
		"list": {
			name: "attribute list <flag>",
			description: "List all entities",
			callback: listAttributesCommand,
		},
		"info": {
			name: "attribute info <flags> <slug|id>",
			description: "Get attribute information by slug or by id",
			callback: infoAttributeCommand,
		},
		"delete": {
			name: "attribute delete <slug|id>",
			description: "Delete an attribute by slug or by id",
			callback: deleteAttributeCommand,
		},
		"help": {
			name: "attribute help",
			description: "Prints this message",
			callback: helpAttributeCommand,
		},
	}
}

func createAttributeCommand(s state) error {
	registerFs := flag.NewFlagSet("attribute", flag.ExitOnError)

	var name string
	var slug string
	var typeString string

	registerFs.StringVar(&name, "name", "", "Name of new attribute")
	registerFs.StringVar(&name, "n", "", "Name of new attribute (shorthand)")

	registerFs.StringVar(&slug, "slug", "", "Slug for new attribute")
	registerFs.StringVar(&slug, "s", "", "Slug for new attribute (shorthand)")

	registerFs.StringVar(&typeString, "type", "", "Type of new attribute")
	registerFs.StringVar(&typeString, "t", "", "Type of new attribute (shorthand)")

	registerFs.Parse(s.args)

	if name == "" || slug == "" || typeString == "" {
		fmt.Fprintln(os.Stderr, "Name, slug and type must be provided, but one or more was empty")
		// TODO print usage
		os.Exit(1)
	}

	if !geaves.ValidAttributeType(typeString) {
		fmt.Fprintln(os.Stderr, "type provided was not valid, please use a valid type")
		// TODO print usage with types
		// TODO print valid types
		os.Exit(1)
	}

	attribute, err := s.queries.CreateAttribute(context.Background(), geaves.CreateAttributeParam{Name: name, Slug: slug, Type: geaves.AttributeType(typeString)})
	if err != nil {
		return err
	}

	fmt.Printf("Successfully created %s (%s), of type %s\n", attribute.Name, attribute.Slug, attribute.Type)
	return nil
}

func listAttributesCommand(s state) error {
 	listFs := flag.NewFlagSet("attribute", flag.ExitOnError)

	var entities bool

	listFs.BoolVar(&entities, "entities", false, "show entity attributes")
	listFs.BoolVar(&entities, "e", false, "show entity attributes(shorthand)")

	listFs.Parse(s.args)

	attributes, err := s.queries.ListAttributes(context.Background(), entities)
	if err != nil {
		return err
	}

	var sb strings.Builder
	for _, attribute := range attributes {
		attributeString, err := attributeToString(attribute, !entities, s.queries)
		if err != nil {
			return err
		}

		sb.WriteString(attributeString)
	}

	if entities {
		sb.WriteString(fmt.Sprintln("* required on items of entity"))
	}

	fmt.Print(sb.String())
	return nil
}

func infoAttributeCommand(s state) error {
	infoFs := flag.NewFlagSet("attribute", flag.ExitOnError)

	var hideEntities bool

	infoFs.BoolVar(&hideEntities, "hide-entities", false, "Don't show entities that use this attribute (shorthand)")
	infoFs.BoolVar(&hideEntities, "E", false, "Don't show entities that use this attribute")

	infoFs.Parse(s.args)

	if infoFs.NArg() < 1 {
		return fmt.Errorf("%s requires an argument, the slug or id of the attribute to look up", s.cmdName)
	}

	attribute, err := getAttributeByIdOrSlug(infoFs.Arg(0), !hideEntities, s.queries)
	if err != nil {
		return err
	}

	var sb strings.Builder
	attributeString, err := attributeToString(attribute, hideEntities, s.queries)
	if err != nil {
		return err
	}

	sb.WriteString(attributeString)

	if !hideEntities {
		sb.WriteString(fmt.Sprintln("* required on items of entity"))
	}

	fmt.Print(sb.String())
	return nil
}

func updateAttributeCommand(s state) error {
	updateFs := flag.NewFlagSet("attribute", flag.ExitOnError)

	var name string
	var slug string
	var newType string

	updateFs.StringVar(&name, "name", "", "New name for an attribute")
	updateFs.StringVar(&name, "n", "", "New name for an attribute (shorthand)")

	updateFs.StringVar(&slug, "slug", "", "New slug for an attribute")
	updateFs.StringVar(&slug, "s", "", "New slug for an attribute (shorthand)")

	updateFs.StringVar(&newType, "type", "", "New type for an attribute")
	updateFs.StringVar(&newType, "t", "", "New type for an attribute (shorthand)")

	updateFs.Parse(s.args)

	if !geaves.ValidAttributeType(newType) {
		fmt.Println("The new type is not valid, ignoring this option")
		newType = ""
	}

	if name == "" && slug == "" && newType == "" {
		fmt.Println("No valid updating flags were given, nothing to do")
		os.Exit(1)
	}

	if updateFs.NArg() < 1 {
		return fmt.Errorf("%s requires an argument, the slug or id of the attribute to look up", s.cmdName)
	}

	attribute, err := getAttributeByIdOrSlug(updateFs.Arg(0), false, s.queries)
	if err != nil {
		return err
	}

	if (name == attribute.Name || name == "") && (slug == attribute.Slug || slug == "") && (geaves.AttributeType(newType) == attribute.Type || newType == "") {
		fmt.Println("Attribute already has these fields, nothing to do")
		return nil
	}

	// TODO goroutine, waitgroup and channel errors into []error
	// TODO replace this with attribute.Save() call for convenience
	if attribute.Name != name && name != "" {
		err := s.queries.UpdateAttributeName(context.Background(), name, attribute.ID)
		if err != nil {
			return err
		}
	}

	if attribute.Slug != slug && slug != "" {
		err := s.queries.UpdateAttributeSlug(context.Background(), slug, attribute.ID)
		if err != nil {
			return err
		}
	}

	if newType != "" && attribute.Type != geaves.AttributeType(newType) {
		err := s.queries.UpdateAttributeType(context.Background(), geaves.AttributeType(newType), attribute.ID)
		if err != nil {
			return err
		}
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Successfully updated %v", attribute.ID))

	if attribute.Name != name && name != "" {
		sb.WriteString(fmt.Sprintf(" %s -> %s", attribute.Name, name))
	} else {
		sb.WriteString(fmt.Sprintf(" %s", attribute.Name))
	}

	if attribute.Slug != slug && slug != "" {
		sb.WriteString(fmt.Sprintf(" (%s -> %s)", attribute.Slug, slug))
	} else {
		sb.WriteString(fmt.Sprintf(" (%s)", attribute.Slug))
	}

	if attribute.Type != geaves.AttributeType(newType) && newType != "" {
		sb.WriteString(fmt.Sprintf(": %s -> %s", attribute.Type, newType))
	} else {
		sb.WriteString(fmt.Sprintf(": %s", attribute.Type))
	}


	fmt.Println(sb.String())
	return nil
}

func deleteAttributeCommand(s state) error {
	if len(s.args) < 1 {
		return fmt.Errorf("%s requires 1 argument, either the id or the slug of the attribute", s.cmdName)
	}

	attribute, err := getAttributeByIdOrSlug(s.args[0], false, s.queries)
	if err != nil {
		return err
	}

	err = s.queries.DeleteAttribute(context.Background(), attribute.ID)
	if err == nil {
		fmt.Printf("Successfully delete %s\n", attribute.Name)
	}

	return err
}

func getAttributeByIdOrSlug(search string, entities bool, queries *geaves.Queries) (geaves.Attribute, error) {
	id, err := strconv.ParseInt(search, 10, 64)
	if err == nil {
		return queries.GetAttribute(context.Background(), geaves.GetAttributeParam{WithEntities: entities, Field: geaves.ByID, Value: id})
	}

	return queries.GetAttribute(context.Background(), geaves.GetAttributeParam{WithEntities: entities, Field: geaves.BySlug, Value: search})
}

func attributeToString(attribute geaves.Attribute, skipEntities bool, queries *geaves.Queries) (string, error) {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("%s\n", strings.Repeat("-", 20)))
	sb.WriteString(fmt.Sprintf("| %v.%s (%s)\n", attribute.ID, attribute.Name, attribute.Slug))
	sb.WriteString(fmt.Sprintf("%s\n", strings.Repeat("-", 20)))

	if skipEntities {
		return sb.String(), nil
	}

	entities, err := attribute.GetEntities(context.Background(), queries)
	if err != nil {
		return "", fmt.Errorf("Failed to get entities that use this attribute: %w", err)
	}
	if entities != nil {
		sb.WriteString("| Entities\n")

		for _, entity := range entities {
			reqString := " "
			if entity.Required {
				reqString = "*"
			}

			sb.WriteString(fmt.Sprintf("|  %s%s (%s)\n", reqString, entity.Name, entity.Slug))
		}

		sb.WriteString(fmt.Sprintf("%s\n", strings.Repeat("-", 20)))
	}

	return sb.String(), nil
}

func helpAttributeCommand(s state) (err error) {
	if len(s.args) > 0 {
		switch (s.args[0]) {
		case "create":
			fmt.Print(`
geaves-cli attribute create <flags>

Create a new attribute

Required flags
  -n | --name  - name of the new attribute
  -s | --slug  - slug of the new attribute
  -t | --type  - type of the new attribute

Type MUST be one of bool, string, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, byte, rune, float32, float64, blob, date, time, datetime
`)
			return
		case "update":
			fmt.Print(`
geaves-cli attribute update <flags> <slug|id>
NOTE flags must be before arguments

Update an existing attribute by slug or id

Available flags
  -n | --name  - name of the new attribute
  -s | --slug  - slug of the new attribute
  -t | --type  - type of the new attribute

Type MUST be one of bool, string, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, byte, rune, float32, float64, blob, date, time, datetime

One of name, slug or type is required
`)
			return
		case "list":
			fmt.Print(`
geaves-cli attribute list <flags>

List all attributes registered to the system

Available flags
  -e | --entities  - Show (default) or Hide entities that use each attribute (-e=false to show, -e=true to hide)
`)
			return
		case "info":
			fmt.Print(`
geaves-cli attribute list <flags>

Print details of oen attribute registered in the system

Available flags
  -E | --hide-entities  - Show (default) or hide entities that use each attribute (no flags or -E=false to show, -E=true to hide)
`)
			return
		case "delete":
			fmt.Print(`
geaves-cli attribute delete <slug|id>

Delete an attribute by it's id or slug
`)
			return
		default:
			fmt.Println("Unknown subcommand, usage:")
		}
	}

	fmt.Print(`
geaves-cli attribute [subcommand]

Available subcommands
  create <flags>           - create a new attribute with data provided in flags
  update <flags> <slug|id> - update using data provided in flags by attribute id or attribute slug
  list <flags>             - list all attributes, configurable with flags
  info <flags> <slug|id>   - details of a single flag by id or slug, configurable with flags
  delete <slug|id>         - delete an attribute by slug or id
  help [subcommand]        - Print this message or help message of a subcommand
`)
	return
}
