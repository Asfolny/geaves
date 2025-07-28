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

func entityCommand(s state) error {
	if len(s.args) < 1 {
		// TODO print subcommand usage instead
		return fmt.Errorf("%s requires 1 argument, the subcommand", s.cmdName)
	}

	cmds := getEntityCommands()
	cmd, ok := cmds[s.args[0]]
	if !ok {
		return fmt.Errorf("%s: entity command not found\n", s.args[0])
	}

	return  cmd.callback(state{s.args[0], s.args[1:], s.queries})
}

func getEntityCommands() map[string]command {
	return map[string]command{
		"create": {
			name: "entity create <flags>",
			description: "Create a new entity",
			callback: createEntityCommand,
		},
		"update": {
			name: "entity update <flags> <slug|id>",
			description: "Update an entity by id or slug",
			callback: updateEntityCommand,
		},
		"list": {
			name: "entity list <flags>",
			description: "List all entities",
			callback: listEntitiesCommand,
		},
		"info": {
			name: "entity info <flags> <slug|id>",
			description: "Get entity information by slug or by id",
			callback: infoEntityCommand,
		},
		"delete": {
			name: "entity delete <slug|id>",
			description: "Delete an entity by slug or by id",
			callback: deleteEntityCommand,
		},
		"help": {
			name: "entity help",
			description: "Prints this message",
			callback: helpEntityCommand,
		},
	}
}

func createEntityCommand(s state) error {
	registerFs := flag.NewFlagSet("entity", flag.ExitOnError)

	var name string
	var slug string

	registerFs.StringVar(&name, "name", "", "Name of new entity")
	registerFs.StringVar(&name, "n", "", "Name of new entity (shorthand)")

	registerFs.StringVar(&slug, "slug", "", "Slug for new entity")
	registerFs.StringVar(&slug, "s", "", "Slug for new entity (shorthand)")

	registerFs.Parse(s.args)

	if name == "" || slug == "" {
		fmt.Println("Both name and slug must be provided, but one was empty")
		os.Exit(1)
	}

	entity, err := s.queries.CreateEntity(context.Background(), geaves.CreateEntityParam{Name: name, Slug: slug})
	if err != nil {
		return err
	}

	fmt.Printf("Successfully created %s (%s)\n", entity.Name, entity.Slug)
	return nil
}

func listEntitiesCommand(s state) error {
	listFs := flag.NewFlagSet("entity", flag.ExitOnError)

	var attributes bool

	listFs.BoolVar(&attributes, "attributes", false, "Hide entity attributes (shorthand)")
	listFs.BoolVar(&attributes, "a", false, "Hide entity attributes")

	listFs.Parse(s.args)

	entities, err := s.queries.ListEntities(context.Background(), attributes)
	if err != nil {
		return err
	}

	var sb strings.Builder
	for _, entity := range entities {
		entityString, err := entityToString(entity, !attributes, s.queries)
		if err != nil {
			return err
		}

		sb.WriteString(entityString)
	}

	if attributes {
		sb.WriteString(fmt.Sprintln("* are required"))
	}

	fmt.Print(sb.String())
	return nil
}

func infoEntityCommand(s state) error {
	infoFs := flag.NewFlagSet("entity", flag.ExitOnError)

	var hideAttributes bool

	infoFs.BoolVar(&hideAttributes, "hide-attributes", false, "Hide entity attributes (shorthand)")
	infoFs.BoolVar(&hideAttributes, "A", false, "Hide entity attributes")

	infoFs.Parse(s.args)

	if infoFs.NArg() < 1 {
		return fmt.Errorf("%s requires an argument, the slug or id of the entity to look up", s.cmdName)
	}

	entity, err := getEntityByIdOrSlug(infoFs.Arg(0), !hideAttributes, s.queries)
	if err != nil {
		return err
	}

	entityString, err := entityToString(entity, hideAttributes, s.queries)
	if err != nil {
		return err
	}

	var sb strings.Builder
	sb.WriteString(entityString)

	if !hideAttributes {
		sb.WriteString(fmt.Sprintln("* are required"))
	}

	fmt.Print(sb.String())
	return nil
}

func updateEntityCommand(s state) error {
	updateFs := flag.NewFlagSet("entity", flag.ExitOnError)

	var name string
	var slug string

	updateFs.StringVar(&name, "name", "", "Name of new entity")
	updateFs.StringVar(&name, "n", "", "Name of new entity (shorthand)")

	updateFs.StringVar(&slug, "slug", "", "Update the slug on an entity")
	updateFs.StringVar(&slug, "s", "", "Update the slug on an entity (shorthand)")

	updateFs.Parse(s.args)

	if name == "" && slug == "" {
		fmt.Println("No updating flags were given, nothing to do")
		os.Exit(1)
	}

	if updateFs.NArg() < 1 {
		return fmt.Errorf("%s requires an argument, the slug or id of the entity to look up", s.cmdName)
	}

	entity, err := getEntityByIdOrSlug(updateFs.Arg(0), false, s.queries)
	if err != nil {
		return err
	}

	if (name == entity.Name || name == "") && (slug == entity.Slug || slug == "") {
		fmt.Println("Entity already has these fields, nothing to do")
		return nil
	}

	// TODO goroutine, waitgroup and channel errors into []error
	// TODO replace this with (entity Entity) Save() call for convenience
	if entity.Name != name && name != "" {
		err := s.queries.UpdateEntityName(context.Background(), name, entity.ID)
		if err != nil {
			return err
		}
	}

	if entity.Slug != slug && slug != "" {
		err := s.queries.UpdateEntitySlug(context.Background(), slug, entity.ID)
		if err != nil {
			return err
		}
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Successfully updated %v", entity.ID))

	if entity.Name != name && name != "" {
		sb.WriteString(fmt.Sprintf(" %s -> %s", entity.Name, name))
	} else {
		sb.WriteString(fmt.Sprintf(" %s", entity.Name))
	}

	if entity.Slug != slug && slug != "" {
		sb.WriteString(fmt.Sprintf(" (%s -> %s)", entity.Slug, slug))
	} else {
		sb.WriteString(fmt.Sprintf(" (%s)", entity.Slug))
	}

	fmt.Println(sb.String())
	return nil
}

func deleteEntityCommand(s state) error {
	if len(s.args) < 1 {
		return fmt.Errorf("%s requires 1 argument, either the id or the slug of the entity", s.cmdName)
	}

	entity, err := getEntityByIdOrSlug(s.args[0], false, s.queries)
	if err != nil {
		return err
	}

	err = s.queries.DeleteEntity(context.Background(), entity.ID)
	if err == nil {
		fmt.Printf("Successfully delete %s\n", entity.Name)
	}

	return err
}

func getEntityByIdOrSlug(search string, attributes bool, queries *geaves.Queries) (geaves.Entity, error) {
	id, err := strconv.ParseInt(search, 10, 64)
	if err == nil {
		return queries.GetEntity(context.Background(), geaves.GetEntityParam{WithAttributes: attributes, Field: geaves.ByID, Value: id})
	}

	return queries.GetEntity(context.Background(), geaves.GetEntityParam{WithAttributes: !attributes, Field: geaves.BySlug, Value: search})
}

func entityToString(entity geaves.Entity, skipAttributes bool, queries *geaves.Queries) (string, error) {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("%s\n", strings.Repeat("-", 20)))
	sb.WriteString(fmt.Sprintf("| %v.%s (%s)\n", entity.ID, entity.Name, entity.Slug))
	sb.WriteString(fmt.Sprintf("%s\n", strings.Repeat("-", 20)))

	if skipAttributes {
		return sb.String(), nil
	}

	attributes, err := entity.GetAttributes(context.Background(), queries)
	if err != nil {
		return "", fmt.Errorf("Failed to get attributes: %w", err)
	}

	if attributes != nil {
		sb.WriteString("| Attributes\n")

		for _, attribute := range attributes {
			reqString := " "
			if attribute.Required {
				reqString = "*"
			}

			sb.WriteString(fmt.Sprintf("|  %s%s (%s): %s\n", reqString, attribute.Name, attribute.Slug, attribute.Type))
		}

		sb.WriteString(fmt.Sprintf("%s\n", strings.Repeat("-", 20)))
	}

	return sb.String(), nil
}

func helpEntityCommand(s state) (err error) {
	if len(s.args) > 0 {
		switch (s.args[0]) {
		case "create":
			fmt.Print(`
geaves-cli enitity create <flags>

Create a new enitity

Required flags
  -n | --name  - name of the new enitity
  -s | --slug  - slug of the new enitity
`)
			return
		case "update":
			fmt.Print(`
geaves-cli enitity update <flags> <slug|id>
NOTE flags must be before arguments

Update an existing enitity by slug or id

Available flags
  -n | --name  - name of the new enitity
  -s | --slug  - slug of the new enitity

Either name or slug is required
`)
			return
		case "list":
			fmt.Print(`
geaves-cli enitity list <flags>

List all enititys registered to the system

Available flags
  -e | --entities  - Show (default) or Hide entities that use each enitity (-e=false to show, -e=true to hide)
`)
			return
		case "info":
			fmt.Print(`
geaves-cli enitity list <flags>

Print details of a single entity registered in the system

Available flags
  -E | --hide-attributes  - Show (default) or hide attributes that each entity has (no flags or -E=false to show, -E=true to hide)
`)
			return
		case "delete":
			fmt.Print(`
geaves-cli enitity delete <slug|id>

Delete an enitity by it's id or slug
`)
			return
		default:
			fmt.Println("Unknown subcommand, usage:")
		}
	}

	fmt.Print(`
geaves-cli enitity [subcommand]

Available subcommands
  create <flags>           - create a new enitity with data provided in flags
  update <flags> <slug|id> - update using data provided in flags by enitity id or enitity slug
  list <flags>             - list all enititys, configurable with flags
  info <flags> <slug|id>   - details of a single flag by id or slug, configurable with flags
  delete <slug|id>         - delete an enitity by slug or id
  help [subcommand]        - Print this message or help message of a subcommand
`)
	return
}
