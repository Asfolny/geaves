package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Asfolny/geaves"
)

func itemCommand(s state) error {
	if len(s.args) < 1 {
		// TODO print subcommand usage instead
		return fmt.Errorf("%s requires 1 argument, the subcommand", s.cmdName)
	}

	cmds := getItemCommands()
	cmd, ok := cmds[s.args[0]]
	if !ok {
		return fmt.Errorf("%s: entity command not found\n", s.args[0])
	}

	return  cmd.callback(state{s.args[0], s.args[1:], s.queries})
}

func getItemCommands() map[string]command {
	return map[string]command{
		"create": {
			name: "item create <entity id|slug>",
			description: "Create a new item by entity slug or entity id",
			callback: createItemCommand,
		},
		"update": {
			name: "item update <item id> <entity id|slug>",
			description: "Update an item's entity",
			callback: updateItemCommand,
		},
		"add": {
			name: "item add <item id> <attribute id|slug> <value>",
			description: "Try to add a new attribute value to an item",
			callback: itemAddAttributeCommand,
		},
		"del": {
			name: "item del <item id> <attribute id|slug>",
			description: "Try to delete an attribute value from an item",
			callback: itemDelAttributeCommand,
		},
		"set": {
			name: "item set <item id> <attribute id|slug> <value>",
			description: "Try to set an attribute value on an item",
			callback: itemSetAttributeCommand,
		},
		"list": {
			name: "item list",
			description: "List all items",
			callback: listItemsCommand,
		},
		"info": {
			name: "item info <id>",
			description: "Get item information by id",
			callback: infoItemCommand,
		},
		"delete": {
			name: "item delete <id>",
			description: "Delete an item by id",
			callback: deleteItemCommand,
		},
		"help": {
			name: "item help",
			description: "Displays this help message",
			callback: helpItemCommand,
		},
	}
}

func createItemCommand(s state) error {
	if (len(s.args) < 1) {
		return fmt.Errorf("%s needs 1 argument, the slug or the id of an existing entity", s.cmdName)
	}

	entity, err := getEntityByIdOrSlug(s.args[0], true, s.queries)
	if err != nil {
		return fmt.Errorf("Failed to get entity %s needed to create item: %w", entity.Name, err)
	}

	item, err := s.queries.CreateItem(context.Background(), entity.ID)
	if err != nil {
		return fmt.Errorf("Failed to create item: %w", err)
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Successfully created item %v of type %s\n", item.ID, entity.Name))
	attrs, err := entity.GetAttributes(context.Background(), s.queries)

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		fmt.Print(sb.String())
		return fmt.Errorf("Failed to find attributes under entity %s: %w", entity.Name, err)
	}

	var hasRequired bool

	if attrs != nil {
		for _, attr := range attrs {
			if attr.Required {
				if !hasRequired {
					hasRequired = true
					sb.WriteString("Entity has required attribute, to complete the item, these must be set:\n")
				}

				sb.WriteString(fmt.Sprintf("- %s (%s)\n", attr.Name, attr.Slug))
			}
		}
	}

	fmt.Print(sb.String())
	return nil
}

func listItemsCommand(s state) error {
	items, err := s.queries.ListItems(context.Background())
	if err != nil {
		return err
	}

	var sb strings.Builder

	for _, item := range items {
		entity, err := s.queries.GetEntity(context.Background(), geaves.GetEntityParam{WithAttributes: true, Field: geaves.ByID, Value: item.EntityID})
		if err != nil {
			return fmt.Errorf("Failed to get entity for item: %w", err)
		}

		values, err := s.queries.ListItemAttributes(context.Background(), item.ID)
		if err != nil {
			return fmt.Errorf("Failed to get item attributes: %w", err)
		}

		attributes, err := entity.GetAttributes(context.Background(), s.queries)
		if err != nil {
			return fmt.Errorf("Failed to get attributes from entity: %w", err)
		}

		sb.WriteString(itemToString(item, entity, values, attributes, s.queries))

	}

	sb.WriteString(fmt.Sprintln("* are required"))
	fmt.Print(sb.String())
	return nil
}

func infoItemCommand(s state) error {
	if len(s.args) < 1 {
		return fmt.Errorf("%s requires 1 argument, the item id", s.cmdName)
	}

	id, err := strconv.ParseInt(s.args[0], 10, 64)
	if err != nil {
		return fmt.Errorf("Failed to convert id to int64: %w", err)
	}

	item, err := s.queries.GetItem(context.Background(), id)
	if err != nil {
		return fmt.Errorf("Failed to get item: %w", err)
	}

	entity, err := s.queries.GetEntity(context.Background(), geaves.GetEntityParam{WithAttributes: true, Field: geaves.ByID, Value: item.EntityID})
	if err != nil {
		return fmt.Errorf("Failed to get entity for item: %w", err)
	}

	values, err := s.queries.ListItemAttributes(context.Background(), item.ID)
	if err != nil {
		return fmt.Errorf("Failed to get item attributes: %w", err)
	}

	attributes, err := entity.GetAttributes(context.Background(), s.queries)
	if err != nil {
		return fmt.Errorf("Failed to get attributes from entity: %w", err)
	}

	var sb strings.Builder
	sb.WriteString(itemToString(item, entity, values, attributes, s.queries))
	sb.WriteString(fmt.Sprintln("* are required"))
	fmt.Print(sb.String())
	return nil
}

func updateItemCommand(s state) error {
	if len(s.args) < 2 {
		return fmt.Errorf("%s requires 2 arguments, the item id and the new entity id or slug", s.cmdName)
	}

	id, err := strconv.ParseInt(s.args[0], 10, 64)
	if err != nil {
		return fmt.Errorf("Failed to convert id to int64: %w", err)
	}

	item, err := s.queries.GetItem(context.Background(), id)
	if err != nil {
		return fmt.Errorf("Failed to get item: %w", err)
	}

	entity, err := getEntityByIdOrSlug(s.args[1], true, s.queries)
	if err != nil {
		return fmt.Errorf("Failed to get entity %s needed to create item: %w", entity.Name, err)
	}

	err = item.ChangeEntity(context.Background(), entity, s.queries)
	if err != nil {
		return fmt.Errorf("Failed changing to entity (%s) on item: %w", entity.Name, err)
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Successfully created item %v of type %s\n", item.ID, entity.Name))
	attrs, err := entity.GetAttributes(context.Background(), s.queries)

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		fmt.Print(sb.String())
		return fmt.Errorf("Failed to find attributes under entity %s: %w", entity.Name, err)
	}

	var hasRequired bool

	if attrs != nil {
		for _, attr := range attrs {
			if attr.Required {
				if !hasRequired {
					hasRequired = true
					sb.WriteString("Entity has required attribute, to complete the item, these must be set:\n")
				}

				sb.WriteString(fmt.Sprintf("- %s (%s)\n", attr.Name, attr.Slug))
			}
		}
	}

	fmt.Print(sb.String())
	return nil
}

func deleteItemCommand(s state) error {
	if len(s.args) < 1 {
		return fmt.Errorf("%s requires 1 argument, the item id", s.cmdName)
	}

	id, err := strconv.ParseInt(s.args[0], 10, 64)
	if err != nil {
		return fmt.Errorf("Failed to convert id to int64: %w", err)
	}

	err = s.queries.DeleteItemAttributesByItem(context.Background(), id)
	if err != nil {
		return fmt.Errorf("Failed to delete item attributes: %w", err)
	}

	err = s.queries.DeleteItem(context.Background(), id)
	if err == nil {
		fmt.Println("Successfully deleted item and item attributes")
	}

	return err
}

func itemAddAttributeCommand(s state) error {
	if len(s.args) < 3 {
		return fmt.Errorf("%s required 3 arguments, the item id, the attribute id or slug and the new value", s.cmdName)
	}

	id, err := strconv.ParseInt(s.args[0], 10, 64)
	if err != nil {
		return fmt.Errorf("Failed to convert id to int64: %w", err)
	}

	item, err := s.queries.GetItem(context.Background(), id)
	if err != nil {
		return fmt.Errorf("Failed to get item: %w", err)
	}

	attribute, err := getAttributeByIdOrSlug(s.args[1], false, s.queries)
	if err != nil {
		return fmt.Errorf("Failed to get attribute: %w", err)
	}

	// TODO WARN user when value is not compatible with type
	switch (geaves.AttributeType(attribute.Type)) {
	case geaves.BoolType:
		var value *bool
		p, err := strconv.ParseBool(s.args[2])
		if err == nil {
			value = &p
		}

		itemAttribute := geaves.ItemAttribute[*bool]{
			ItemID: item.ID,
			AttributeID: attribute.ID,
			Value: value,
		}

		err = itemAttribute.Create(context.Background(), s.queries)
		if err != nil {
			return fmt.Errorf("failed to create item_attribute record: %w", err)
		}
		break


	case geaves.StringType:
		value := s.args[2]
		itemAttribute := geaves.ItemAttribute[*string]{
			ItemID: item.ID,
			AttributeID: attribute.ID,
			Value: &value,
		}

		err = itemAttribute.Create(context.Background(), s.queries)
		if err != nil {
			return fmt.Errorf("failed to create item_attribute record: %w", err)
		}
		break


	case geaves.IntType:
		var value *int
		num, err := strconv.ParseInt(s.args[2], 10, 0)
		if err == nil {
			lowNum := int(num)
			value = &lowNum
		}

		itemAttribute := geaves.ItemAttribute[*int]{
			ItemID: item.ID,
			AttributeID: attribute.ID,
			Value: value,
		}


		err = itemAttribute.Create(context.Background(), s.queries)
		if err != nil {
			return fmt.Errorf("failed to create item_attribute record: %w", err)
		}
		break


	case geaves.Int8Type:
		var value *int8
		num, err := strconv.ParseInt(s.args[2], 10, 8)
		if err == nil {
			lowNum := int8(num)
			value = &lowNum
		}

		itemAttribute := geaves.ItemAttribute[*int8]{
			ItemID: item.ID,
			AttributeID: attribute.ID,
			Value: value,
		}

		err = itemAttribute.Create(context.Background(), s.queries)
		if err != nil {
			return fmt.Errorf("failed to create item_attribute record: %w", err)
		}
		break


	case geaves.Int16Type:
		var value *int16
		num, err := strconv.ParseInt(s.args[2], 10, 16)
		if err == nil {
			lowNum := int16(num)
			value = &lowNum
		}

		itemAttribute := geaves.ItemAttribute[*int16]{
			ItemID: item.ID,
			AttributeID: attribute.ID,
			Value: value,
		}


		err = itemAttribute.Create(context.Background(), s.queries)
		if err != nil {
			return fmt.Errorf("failed to create item_attribute record: %w", err)
		}
		break


	case geaves.Int32Type:
		var value *int32
		num, err := strconv.ParseInt(s.args[2], 10, 32)
		if err == nil {
			lowNum := int32(num)
			value = &lowNum
		}

		itemAttribute := geaves.ItemAttribute[*int32]{
			ItemID: item.ID,
			AttributeID: attribute.ID,
			Value: value,
		}

		err = itemAttribute.Create(context.Background(), s.queries)
		if err != nil {
			return fmt.Errorf("failed to create item_attribute record: %w", err)
		}
		break


	case geaves.Int64Type:
		var value *int64
		num, err := strconv.ParseInt(s.args[2], 10, 64)
		if err == nil {
			value = &num
		}

		itemAttribute := geaves.ItemAttribute[*int64]{
			ItemID: item.ID,
			AttributeID: attribute.ID,
			Value: value,
		}

		err = itemAttribute.Create(context.Background(), s.queries)
		if err != nil {
			return fmt.Errorf("failed to create item_attribute record: %w", err)
		}
		break


	case geaves.UintType:
		var value *uint
		num, err := strconv.ParseUint(s.args[2], 10, 0)
		if err == nil {
			lowNum := uint(num)
			value = &lowNum
		}

		itemAttribute := geaves.ItemAttribute[*uint]{
			ItemID: item.ID,
			AttributeID: attribute.ID,
			Value: value,
		}

		err = itemAttribute.Create(context.Background(), s.queries)
		if err != nil {
			return fmt.Errorf("failed to create item_attribute record: %w", err)
		}
		break


	case geaves.Uint8Type:
		var value *uint8
		num, err := strconv.ParseUint(s.args[2], 10, 8)
		if err == nil {
			lowNum := uint8(num)
			value = &lowNum
		}

		itemAttribute := geaves.ItemAttribute[*uint8]{
			ItemID: item.ID,
			AttributeID: attribute.ID,
			Value: value,
		}

		err = itemAttribute.Create(context.Background(), s.queries)
		if err != nil {
			return fmt.Errorf("failed to create item_attribute record: %w", err)
		}
		break


	case geaves.Uint16Type:
		var value *uint16
		num, err := strconv.ParseUint(s.args[2], 10, 16)
		if err == nil {
			lowNum := uint16(num)
			value = &lowNum
		}

		itemAttribute := geaves.ItemAttribute[*uint16]{
			ItemID: item.ID,
			AttributeID: attribute.ID,
			Value: value,
		}

		err = itemAttribute.Create(context.Background(), s.queries)
		if err != nil {
			return fmt.Errorf("failed to create item_attribute record: %w", err)
		}
		break


	case geaves.Uint32Type:
		var value *uint32
		num, err := strconv.ParseUint(s.args[2], 10, 32)
		if err == nil {
			lowNum := uint32(num)
			value = &lowNum
		}

		itemAttribute := geaves.ItemAttribute[*uint32]{
			ItemID: item.ID,
			AttributeID: attribute.ID,
			Value: value,
		}

		err = itemAttribute.Create(context.Background(), s.queries)
		if err != nil {
			return fmt.Errorf("failed to create item_attribute record: %w", err)
		}
		break


	case geaves.Uint64Type:
		var value *uint64
		num, err := strconv.ParseUint(s.args[2], 10, 64)
		if err == nil {
			value = &num
		}

		itemAttribute := geaves.ItemAttribute[*uint64]{
			ItemID: item.ID,
			AttributeID: attribute.ID,
			Value: value,
		}

		err = itemAttribute.Create(context.Background(), s.queries)
		if err != nil {
			return fmt.Errorf("failed to create item_attribute record: %w", err)
		}
		break


	case geaves.ByteType:
		bytes := []byte(s.args[2])

		itemAttribute := geaves.ItemAttribute[*byte]{
			ItemID: item.ID,
			AttributeID: attribute.ID,
			Value: &bytes[0],
		}

		err = itemAttribute.Create(context.Background(), s.queries)
		if err != nil {
			return fmt.Errorf("failed to create item_attribute record: %w", err)
		}
		break


	case geaves.RuneType:
		runes := []rune(s.args[2])

		itemAttribute := geaves.ItemAttribute[*rune]{
			ItemID: item.ID,
			AttributeID: attribute.ID,
			Value: &runes[0],
		}

		err = itemAttribute.Create(context.Background(), s.queries)
		if err != nil {
			return fmt.Errorf("failed to create item_attribute record: %w", err)
		}
		break


	case geaves.Float32Type:
		var value *float32
		num, err := strconv.ParseFloat(s.args[2], 32)
		if err == nil {
			lowNum := float32(num)
			value = &lowNum
		}

		itemAttribute := geaves.ItemAttribute[*float32]{
			ItemID: item.ID,
			AttributeID: attribute.ID,
			Value: value,
		}

		err = itemAttribute.Create(context.Background(), s.queries)
		if err != nil {
			return fmt.Errorf("failed to create item_attribute record: %w", err)
		}
		break


	case geaves.Float64Type:
		var value *float64
		num, err := strconv.ParseFloat(s.args[2], 64)
		if err == nil {
			value = &num
		}

		itemAttribute := geaves.ItemAttribute[*float64]{
			ItemID: item.ID,
			AttributeID: attribute.ID,
			Value: value,
		}

		err = itemAttribute.Create(context.Background(), s.queries)
		if err != nil {
			return fmt.Errorf("failed to create item_attribute record: %w", err)
		}
		break


	case geaves.BlobType:
		bytes := []byte(s.args[2])

		itemAttribute := geaves.ItemAttribute[*[]byte]{
			ItemID: item.ID,
			AttributeID: attribute.ID,
			Value: &bytes,
		}

		err = itemAttribute.Create(context.Background(), s.queries)
		if err != nil {
			return fmt.Errorf("failed to create item_attribute record: %w", err)
		}
		break


	case geaves.DateType:
		var value *time.Time
		t, err := time.Parse("2006-01-02", s.args[2])
		if err == nil {
			value = &t
		}

		itemAttribute := geaves.ItemAttribute[*time.Time]{
			ItemID: item.ID,
			AttributeID: attribute.ID,
			Value: value,
		}

		err = itemAttribute.Create(context.Background(), s.queries)
		if err != nil {
			return fmt.Errorf("failed to create item_attribute record: %w", err)
		}
		break

	case geaves.TimeType:
		var value *time.Time
		t, err := time.Parse("15:04:05", s.args[2])
		if err == nil {
			value = &t
		}

		itemAttribute := geaves.ItemAttribute[*time.Time]{
			ItemID: item.ID,
			AttributeID: attribute.ID,
			Value: value,
		}

		err = itemAttribute.Create(context.Background(), s.queries)
		if err != nil {
			return fmt.Errorf("failed to create item_attribute record: %w", err)
		}
		break

	case geaves.DatetimeType:
		var value *time.Time
		t, err := time.Parse("2006-01-02 15:04:05", s.args[2])
		if err == nil {
			value = &t
		}

		itemAttribute := geaves.ItemAttribute[*time.Time]{
			ItemID: item.ID,
			AttributeID: attribute.ID,
			Value: value,
		}

		err = itemAttribute.Create(context.Background(), s.queries)
		if err != nil {
			return fmt.Errorf("failed to create item_attribute record: %w", err)
		}
		break

	default:
		itemAttribute := geaves.ItemAttribute[*any]{
			ItemID: item.ID,
			AttributeID: attribute.ID,
			Value: nil,
		}

		err := itemAttribute.Create(context.Background(), s.queries)
		if err != nil {
			return fmt.Errorf("failed to create item_attribute record: %w", err)
		}
	}

	fmt.Println("Succesfully added new item attribute value to system")
	return nil
}

func itemSetAttributeCommand(s state) error {
	if len(s.args) < 3 {
		return fmt.Errorf("%s required 3 arguments, the item id, the attribute id or slug and the new value", s.cmdName)
	}

	id, err := strconv.ParseInt(s.args[0], 10, 64)
	if err != nil {
		return fmt.Errorf("Failed to convert id to int64: %w", err)
	}

	item, err := s.queries.GetItem(context.Background(), id)
	if err != nil {
		return fmt.Errorf("Failed to get item: %w", err)
	}

	attribute, err := getAttributeByIdOrSlug(s.args[1], false, s.queries)
	if err != nil {
		return fmt.Errorf("Failed to get attribute: %w", err)
	}

	// TODO WARN user when value is not compatible with type
	switch (geaves.AttributeType(attribute.Type)) {
	case geaves.BoolType:
		var value *bool
		p, err := strconv.ParseBool(s.args[2])
		if err == nil {
			value = &p
		}

		itemAttribute := geaves.ItemAttribute[*bool]{
			ItemID: item.ID,
			AttributeID: attribute.ID,
			Value: value,
		}

		err = itemAttribute.Update(context.Background(), s.queries)
		if err != nil {
			return fmt.Errorf("failed to create item_attribute record: %w", err)
		}
		break


	case geaves.StringType:
		value := s.args[2]
		itemAttribute := geaves.ItemAttribute[*string]{
			ItemID: item.ID,
			AttributeID: attribute.ID,
			Value: &value,
		}

		err = itemAttribute.Update(context.Background(), s.queries)
		if err != nil {
			return fmt.Errorf("failed to create item_attribute record: %w", err)
		}
		break


	case geaves.IntType:
		var value *int
		num, err := strconv.ParseInt(s.args[2], 10, 0)
		if err == nil {
			lowNum := int(num)
			value = &lowNum
		}

		itemAttribute := geaves.ItemAttribute[*int]{
			ItemID: item.ID,
			AttributeID: attribute.ID,
			Value: value,
		}

		err = itemAttribute.Update(context.Background(), s.queries)
		if err != nil {
			return fmt.Errorf("failed to create item_attribute record: %w", err)
		}
		break


	case geaves.Int8Type:
		var value *int8
		num, err := strconv.ParseInt(s.args[2], 10, 8)
		if err == nil {
			lowNum := int8(num)
			value = &lowNum
		}

		itemAttribute := geaves.ItemAttribute[*int8]{
			ItemID: item.ID,
			AttributeID: attribute.ID,
			Value: value,
		}

		err = itemAttribute.Update(context.Background(), s.queries)
		if err != nil {
			return fmt.Errorf("failed to create item_attribute record: %w", err)
		}
		break


	case geaves.Int16Type:
		var value *int16
		num, err := strconv.ParseInt(s.args[2], 10, 16)
		if err == nil {
			lowNum := int16(num)
			value = &lowNum
		}

		itemAttribute := geaves.ItemAttribute[*int16]{
			ItemID: item.ID,
			AttributeID: attribute.ID,
			Value: value,
		}

		err = itemAttribute.Update(context.Background(), s.queries)
		if err != nil {
			return fmt.Errorf("failed to create item_attribute record: %w", err)
		}
		break

	case geaves.Int32Type:
		var value *int32
		num, err := strconv.ParseInt(s.args[2], 10, 32)
		if err == nil {
			lowNum := int32(num)
			value = &lowNum
		}

		itemAttribute := geaves.ItemAttribute[*int32]{
			ItemID: item.ID,
			AttributeID: attribute.ID,
			Value: value,
		}

		err = itemAttribute.Update(context.Background(), s.queries)
		if err != nil {
			return fmt.Errorf("failed to create item_attribute record: %w", err)
		}
		break

	case geaves.Int64Type:
		var value *int64
		num, err := strconv.ParseInt(s.args[2], 10, 64)
		if err == nil {
			value = &num
		}

		itemAttribute := geaves.ItemAttribute[*int64]{
			ItemID: item.ID,
			AttributeID: attribute.ID,
			Value: value,
		}

		err = itemAttribute.Update(context.Background(), s.queries)
		if err != nil {
			return fmt.Errorf("failed to create item_attribute record: %w", err)
		}
		break

	case geaves.UintType:
		var value *uint
		num, err := strconv.ParseUint(s.args[2], 10, 0)
		if err == nil {
			lowNum := uint(num)
			value = &lowNum
		}

		itemAttribute := geaves.ItemAttribute[*uint]{
			ItemID: item.ID,
			AttributeID: attribute.ID,
			Value: value,
		}

		err = itemAttribute.Update(context.Background(), s.queries)
		if err != nil {
			return fmt.Errorf("failed to create item_attribute record: %w", err)
		}
		break

	case geaves.Uint8Type:
		var value *uint8
		num, err := strconv.ParseUint(s.args[2], 10, 8)
		if err == nil {
			lowNum := uint8(num)
			value = &lowNum
		}

		itemAttribute := geaves.ItemAttribute[*uint8]{
			ItemID: item.ID,
			AttributeID: attribute.ID,
			Value: value,
		}

		err = itemAttribute.Update(context.Background(), s.queries)
		if err != nil {
			return fmt.Errorf("failed to create item_attribute record: %w", err)
		}
		break

	case geaves.Uint16Type:
		var value *uint16
		num, err := strconv.ParseUint(s.args[2], 10, 16)
		if err == nil {
			lowNum := uint16(num)
			value = &lowNum
		}

		itemAttribute := geaves.ItemAttribute[*uint16]{
			ItemID: item.ID,
			AttributeID: attribute.ID,
			Value: value,
		}

		err = itemAttribute.Update(context.Background(), s.queries)
		if err != nil {
			return fmt.Errorf("failed to create item_attribute record: %w", err)
		}
		break

	case geaves.Uint32Type:
		var value *uint32
		num, err := strconv.ParseUint(s.args[2], 10, 32)
		if err == nil {
			lowNum := uint32(num)
			value = &lowNum
		}

		itemAttribute := geaves.ItemAttribute[*uint32]{
			ItemID: item.ID,
			AttributeID: attribute.ID,
			Value: value,
		}

		err = itemAttribute.Update(context.Background(), s.queries)
		if err != nil {
			return fmt.Errorf("failed to create item_attribute record: %w", err)
		}
		break

	case geaves.Uint64Type:
		var value *uint64
		num, err := strconv.ParseUint(s.args[2], 10, 64)
		if err == nil {
			value = &num
		}

		itemAttribute := geaves.ItemAttribute[*uint64]{
			ItemID: item.ID,
			AttributeID: attribute.ID,
			Value: value,
		}


		err = itemAttribute.Update(context.Background(), s.queries)
		if err != nil {
			return fmt.Errorf("failed to create item_attribute record: %w", err)
		}
		break

	case geaves.ByteType:
		bytes := []byte(s.args[2])

		itemAttribute := geaves.ItemAttribute[*byte]{
			ItemID: item.ID,
			AttributeID: attribute.ID,
			Value: &bytes[0],
		}


		err = itemAttribute.Update(context.Background(), s.queries)
		if err != nil {
			return fmt.Errorf("failed to create item_attribute record: %w", err)
		}
		break

	case geaves.RuneType:
		runes := []rune(s.args[2])

		itemAttribute := geaves.ItemAttribute[*rune]{
			ItemID: item.ID,
			AttributeID: attribute.ID,
			Value: &runes[0],
		}

		err = itemAttribute.Update(context.Background(), s.queries)
		if err != nil {
			return fmt.Errorf("failed to create item_attribute record: %w", err)
		}
		break

	case geaves.Float32Type:
		var value *float32
		num, err := strconv.ParseFloat(s.args[2], 32)
		if err == nil {
			lowNum := float32(num)
			value = &lowNum
		}

		itemAttribute := geaves.ItemAttribute[*float32]{
			ItemID: item.ID,
			AttributeID: attribute.ID,
			Value: value,
		}

		err = itemAttribute.Update(context.Background(), s.queries)
		if err != nil {
			return fmt.Errorf("failed to create item_attribute record: %w", err)
		}
		break

	case geaves.Float64Type:
		var value *float64
		num, err := strconv.ParseFloat(s.args[2], 64)
		if err == nil {
			value = &num
		}

		itemAttribute := geaves.ItemAttribute[*float64]{
			ItemID: item.ID,
			AttributeID: attribute.ID,
			Value: value,
		}


		err = itemAttribute.Update(context.Background(), s.queries)
		if err != nil {
			return fmt.Errorf("failed to create item_attribute record: %w", err)
		}
		break

	case geaves.BlobType:
		bytes := []byte(s.args[2])

		itemAttribute := geaves.ItemAttribute[*[]byte]{
			ItemID: item.ID,
			AttributeID: attribute.ID,
			Value: &bytes,
		}

		err = itemAttribute.Update(context.Background(), s.queries)
		if err != nil {
			return fmt.Errorf("failed to create item_attribute record: %w", err)
		}
		break

	case geaves.DateType:
		var value *time.Time
		t, err := time.Parse("2006-01-02", s.args[2])
		if err == nil {
			value = &t
		}

		itemAttribute := geaves.ItemAttribute[*time.Time]{
			ItemID: item.ID,
			AttributeID: attribute.ID,
			Value: value,
		}

		err = itemAttribute.Update(context.Background(), s.queries)
		if err != nil {
			return fmt.Errorf("failed to create item_attribute record: %w", err)
		}
		break

	case geaves.TimeType:
		var value *time.Time
		t, err := time.Parse("15:04:05", s.args[2])
		if err == nil {
			value = &t
		}

		itemAttribute := geaves.ItemAttribute[*time.Time]{
			ItemID: item.ID,
			AttributeID: attribute.ID,
			Value: value,
		}

		err = itemAttribute.Update(context.Background(), s.queries)
		if err != nil {
			return fmt.Errorf("failed to create item_attribute record: %w", err)
		}
		break

	case geaves.DatetimeType:
		var value *time.Time
		t, err := time.Parse("2006-01-02 15:04:05", s.args[2])
		fmt.Printf("%v\n%v\n", err, s.args[2])
		if err == nil {
			value = &t
		}

		itemAttribute := geaves.ItemAttribute[*time.Time]{
			ItemID: item.ID,
			AttributeID: attribute.ID,
			Value: value,
		}

		err = itemAttribute.Update(context.Background(), s.queries)
		if err != nil {
			return fmt.Errorf("failed to create item_attribute record: %w", err)
		}
		break

	default:
		itemAttribute := geaves.ItemAttribute[*any]{
			ItemID: item.ID,
			AttributeID: attribute.ID,
			Value: nil,
		}

		err = itemAttribute.Update(context.Background(), s.queries)
		if err != nil {
			return fmt.Errorf("failed to create item_attribute record: %w", err)
		}
	}

	fmt.Println("succesfully updated attribute on item")
	return nil
}

func itemDelAttributeCommand(s state) error {
	if len(s.args) < 2 {
		return fmt.Errorf("%s required 2 arguments, the item id and attribute id to delete item attribute", s.cmdName)
	}

	id, err := strconv.ParseInt(s.args[0], 10, 64)
	if err != nil {
		return fmt.Errorf("Failed to convert id to int64: %w", err)
	}

	item, err := s.queries.GetItem(context.Background(), id)
	if err != nil {
		return fmt.Errorf("Failed to get item: %w", err)
	}

	attribute, err := getAttributeByIdOrSlug(s.args[1], false, s.queries)
	if err != nil {
		return fmt.Errorf("Failed to get attribute: %w", err)
	}


	err = s.queries.DeleteItemAttributes(context.Background(), item.ID, attribute.ID)
	if err == nil {
		fmt.Println("Succesfully delete item's attribute value")
	}

	return err
}

func helpItemCommand(s state) (err error) {
	if len(s.args) > 0 {
		switch(s.args[0]) {
		case "create":
			fmt.Print(`
geaves-cli item create <entity id|slug>

Creates a new item, by first looking up the entity through it's provided id or slug value
`)
			return
		case "update":
			fmt.Print(`
geaves-cli item update <item id> <entity id|slug>

Change an item's entity using a provided entity id or entity slug value
`)
			return
		case "add":
			fmt.Print(`
geaves-cli item add <item id> <attribute id|slug> <value>

Add a new item attribute value to the system using a provided item id, attribute id or attribute slug and value which can by any type
`)
			return
		case "del":
			fmt.Print(`
geaves-cli item del <item id> <attribute id|slug>

Remove an item's attribute value using the provided item id and attribute id or attribute slug
`)
			return
		case "set":
			fmt.Print(`
geaves-cli item list <item id> <attribute id|slug> <value>

Update an item's attribute value using the provided item id, attribute id or attribute slug and new value of same type
`)
			return
		case "list":
			fmt.Print(`
geaves-cli item list

List all items and values to screen
`)
			return
		case "info":
			fmt.Print(`
geaves-cli item info <item id>

Print item details to screen
`)
			return
		case "delete":
			fmt.Print(`
geaves-cli item delete <item id>

Delete an item by the provided item id
`)
			return
		default:
			fmt.Println("Unknown subcommand, usage:")
		}
	}

	fmt.Print(`
geaves-cli item [subcommand]

Listed here are general flags for all item subcommands, for help on each command try: help <subcommand>

Available subcommands
  create <entity id|slug>                   - create a new item of type using entity id or slug
  update <item id> <entity id|slug>         - change an item's entity by entity id or slug
  add <item id> <attribute id|slug> <value> - add a new attribute value to an item
  del <item id> <attribute id|slug>         - remote a value from an item
  set <item id> <attribute id|slug> <value> - update an item's value
  list                                      - prints all items
  info <id>                                 - prints item details by id
  delete <id>                               - delete an item by id
  help [subcommand]                         - prints this message or the help info on a subcommand
\n`)
	return
}

func itemToString(item geaves.Item, entity geaves.Entity, itemAttributes []geaves.ItemAttribute[*any], attributes []geaves.EntityAttributeEmbed, queries *geaves.Queries) string  {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("%s\n", strings.Repeat("-", 20)))
	sb.WriteString(fmt.Sprintf("| Item %v (%s)\n", item.ID, entity.Name))

	for _, itemAttribute := range itemAttributes {
		var attribute geaves.EntityAttributeEmbed
		for _, attr := range attributes {
			if attr.ID == itemAttribute.AttributeID {
				attribute = attr
				break
			}
		}

		if attribute.ID == 0 {
			if itemAttribute.Value == nil {
				sb.WriteString("|   Invalid attribute: nil\n")
			} else {
				sb.WriteString(fmt.Sprintf("|   Invalid attribute: %v", *itemAttribute.Value))
			}
			continue
		}

		reqString := " "
		if attribute.Required {
			reqString = "*"
		}

		var value any

		switch (geaves.AttributeType(attribute.Type)) {
		case geaves.BoolType:
			itemAttribute := geaves.ItemAttribute[*bool]{
				ItemID: item.ID,
				AttributeID: attribute.ID,
				Type: attribute.Type,
			}
			err := itemAttribute.Load(context.Background(), queries)
			if err != nil {
				sb.WriteString(fmt.Sprintf("|  Failed to load value of %s: %v\n", itemAttribute.Type, err))
			}

			if itemAttribute.Value == nil {
				value = nil
			} else {
				value = *itemAttribute.Value
			}
			break

		case geaves.StringType:
			itemAttribute := geaves.ItemAttribute[*string]{
				ItemID: item.ID,
				AttributeID: attribute.ID,
				Type: attribute.Type,
			}

			err := itemAttribute.Load(context.Background(), queries)
			if err != nil {
				sb.WriteString(fmt.Sprintf("|  Failed to load value of %s: %v\n", itemAttribute.Type, err))
			}

			if itemAttribute.Value == nil {
				value = nil
			} else {
				value = *itemAttribute.Value
			}

			break


		case geaves.IntType:
			itemAttribute := geaves.ItemAttribute[*int]{
				ItemID: item.ID,
				AttributeID: attribute.ID,
				Type: attribute.Type,
			}

			err := itemAttribute.Load(context.Background(), queries)
			if err != nil {
				sb.WriteString(fmt.Sprintf("|  Failed to load value of %s: %v\n", itemAttribute.Type, err))
			}

			if itemAttribute.Value == nil {
				value = nil
			} else {
				value = *itemAttribute.Value
			}
			break

		case geaves.Int8Type:
			itemAttribute := geaves.ItemAttribute[*int8]{
				ItemID: item.ID,
				AttributeID: attribute.ID,
				Type: attribute.Type,
			}

			err := itemAttribute.Load(context.Background(), queries)
			if err != nil {
				sb.WriteString(fmt.Sprintf("|  Failed to load value of %s: %v\n", itemAttribute.Type, err))
			}

			if itemAttribute.Value == nil {
				value = nil
			} else {
				value = *itemAttribute.Value
			}
			break

		case geaves.Int16Type:
			itemAttribute := geaves.ItemAttribute[*int16]{
				ItemID: item.ID,
				AttributeID: attribute.ID,
				Type: attribute.Type,
			}

			err := itemAttribute.Load(context.Background(), queries)
			if err != nil {
				sb.WriteString(fmt.Sprintf("|  Failed to load value of %s: %v\n", itemAttribute.Type, err))
			}

			if itemAttribute.Value == nil {
				value = nil
			} else {
				value = *itemAttribute.Value
			}
			break

		case geaves.Int32Type:
			itemAttribute := geaves.ItemAttribute[*int32]{
				ItemID: item.ID,
				AttributeID: attribute.ID,
				Type: attribute.Type,
			}

			err := itemAttribute.Load(context.Background(), queries)
			if err != nil {
				sb.WriteString(fmt.Sprintf("|  Failed to load value of %s: %v\n", itemAttribute.Type, err))
			}

			if itemAttribute.Value == nil {
				value = nil
			} else {
				value = *itemAttribute.Value
			}
			break

		case geaves.Int64Type:
			itemAttribute := geaves.ItemAttribute[*int64]{
				ItemID: item.ID,
				AttributeID: attribute.ID,
				Type: attribute.Type,
			}

			err := itemAttribute.Load(context.Background(), queries)
			if err != nil {
				sb.WriteString(fmt.Sprintf("|  Failed to load value of %s: %v\n", itemAttribute.Type, err))
			}

			if itemAttribute.Value == nil {
				value = nil
			} else {
				value = *itemAttribute.Value
			}
			break

		case geaves.UintType:
			itemAttribute := geaves.ItemAttribute[*uint]{
				ItemID: item.ID,
				AttributeID: attribute.ID,
				Type: attribute.Type,
			}

			err := itemAttribute.Load(context.Background(), queries)
			if err != nil {
				sb.WriteString(fmt.Sprintf("|  Failed to load value of %s: %v\n", itemAttribute.Type, err))
			}

			if itemAttribute.Value == nil {
				value = nil
			} else {
				value = *itemAttribute.Value
			}
			break

		case geaves.Uint8Type:
			itemAttribute := geaves.ItemAttribute[*uint8]{
				ItemID: item.ID,
				AttributeID: attribute.ID,
				Type: attribute.Type,
			}

			err := itemAttribute.Load(context.Background(), queries)
			if err != nil {
				sb.WriteString(fmt.Sprintf("|  Failed to load value of %s: %v\n", itemAttribute.Type, err))
			}

			if itemAttribute.Value == nil {
				value = nil
			} else {
				value = *itemAttribute.Value
			}
			break

		case geaves.Uint16Type:
			itemAttribute := geaves.ItemAttribute[*uint16]{
				ItemID: item.ID,
				AttributeID: attribute.ID,
				Type: attribute.Type,
			}

			err := itemAttribute.Load(context.Background(), queries)
			if err != nil {
				sb.WriteString(fmt.Sprintf("|  Failed to load value of %s: %v\n", itemAttribute.Type, err))
			}

			if itemAttribute.Value == nil {
				value = nil
			} else {
				value = *itemAttribute.Value
			}
			break

		case geaves.Uint32Type:
			itemAttribute := geaves.ItemAttribute[*uint32]{
				ItemID: item.ID,
				AttributeID: attribute.ID,
				Type: attribute.Type,
			}

			err := itemAttribute.Load(context.Background(), queries)
			if err != nil {
				sb.WriteString(fmt.Sprintf("|  Failed to load value of %s: %v\n", itemAttribute.Type, err))
			}

			if itemAttribute.Value == nil {
				value = nil
			} else {
				value = *itemAttribute.Value
			}
			break

		case geaves.Uint64Type:
			itemAttribute := geaves.ItemAttribute[*uint64]{
				ItemID: item.ID,
				AttributeID: attribute.ID,
				Type: attribute.Type,
			}

			err := itemAttribute.Load(context.Background(), queries)
			if err != nil {
				sb.WriteString(fmt.Sprintf("|  Failed to load value of %s: %v\n", itemAttribute.Type, err))
			}

			if itemAttribute.Value == nil {
				value = nil
			} else {
				value = *itemAttribute.Value
			}
			break

		case geaves.ByteType:
			itemAttribute := geaves.ItemAttribute[*byte]{
				ItemID: item.ID,
				AttributeID: attribute.ID,
				Type: attribute.Type,
			}

			err := itemAttribute.Load(context.Background(), queries)
			if err != nil {
				sb.WriteString(fmt.Sprintf("|  Failed to load value of %s: %v\n", itemAttribute.Type, err))
			}

			if itemAttribute.Value == nil {
				value = nil
			} else {
				value = *itemAttribute.Value
			}
			break

		case geaves.RuneType:
			itemAttribute := geaves.ItemAttribute[*rune]{
				ItemID: item.ID,
				AttributeID: attribute.ID,
				Type: attribute.Type,
			}

			err := itemAttribute.Load(context.Background(), queries)
			if err != nil {
				sb.WriteString(fmt.Sprintf("|  Failed to load value of %s: %v\n", itemAttribute.Type, err))
			}

			if itemAttribute.Value == nil {
				value = nil
			} else {
				value = *itemAttribute.Value
			}
			break

		case geaves.Float32Type:
			itemAttribute := geaves.ItemAttribute[*float32]{
				ItemID: item.ID,
				AttributeID: attribute.ID,
				Type: attribute.Type,
			}

			err := itemAttribute.Load(context.Background(), queries)
			if err != nil {
				sb.WriteString(fmt.Sprintf("|  Failed to load value of %s: %v\n", itemAttribute.Type, err))
			}

			if itemAttribute.Value == nil {
				value = nil
			} else {
				value = *itemAttribute.Value
			}
			break

		case geaves.Float64Type:
			itemAttribute := geaves.ItemAttribute[*float64]{
				ItemID: item.ID,
				AttributeID: attribute.ID,
				Type: attribute.Type,
			}

			err := itemAttribute.Load(context.Background(), queries)
			if err != nil {
				sb.WriteString(fmt.Sprintf("|  Failed to load value of %s: %v\n", itemAttribute.Type, err))
			}

			if itemAttribute.Value == nil {
				value = nil
			} else {
				value = *itemAttribute.Value
			}
			break

		case geaves.BlobType:
			itemAttribute := geaves.ItemAttribute[*[]byte]{
				ItemID: item.ID,
				AttributeID: attribute.ID,
				Type: attribute.Type,
			}

			err := itemAttribute.Load(context.Background(), queries)
			if err != nil {
				sb.WriteString(fmt.Sprintf("|  Failed to load value of %s: %v\n", itemAttribute.Type, err))
			}

			if itemAttribute.Value == nil {
				value = nil
			} else {
				value = *itemAttribute.Value
			}
			break

		case geaves.DateType:
			itemAttribute := geaves.ItemAttribute[*time.Time]{
				ItemID: item.ID,
				AttributeID: attribute.ID,
				Type: attribute.Type,
			}

			err := itemAttribute.Load(context.Background(), queries)
			if err != nil {
				sb.WriteString(fmt.Sprintf("|  Failed to load value of %s: %v\n", itemAttribute.Type, err))
			}

			if itemAttribute.Value == nil {
				value = nil
			} else {
				value = *itemAttribute.Value
			}
			break

		case geaves.TimeType:
			itemAttribute := geaves.ItemAttribute[*time.Time]{
				ItemID: item.ID,
				AttributeID: attribute.ID,
				Type: attribute.Type,
			}

			err := itemAttribute.Load(context.Background(), queries)
			if err != nil {
				sb.WriteString(fmt.Sprintf("|  Failed to load value of %s: %v\n", itemAttribute.Type, err))
			}

			if itemAttribute.Value == nil {
				value = nil
			} else {
				value = *itemAttribute.Value
			}
			break

		case geaves.DatetimeType:
			itemAttribute := geaves.ItemAttribute[*time.Time]{
				ItemID: item.ID,
				AttributeID: attribute.ID,
				Type: attribute.Type,
			}

			err := itemAttribute.Load(context.Background(), queries)
			if err != nil {
				sb.WriteString(fmt.Sprintf("|  Failed to load value of %s: %v\n", itemAttribute.Type, err))
			}

			if itemAttribute.Value == nil {
				value = nil
			} else {
				value = *itemAttribute.Value
			}
			break

		default:
			itemAttribute := geaves.ItemAttribute[*any]{
				ItemID: item.ID,
				AttributeID: attribute.ID,
			}

			err := itemAttribute.Load(context.Background(), queries)
			if err != nil {
				sb.WriteString(fmt.Sprintf("|  Failed to load value of %s: %v\n", attribute.Type, err))
			}

			if itemAttribute.Value == nil {
				value = nil
			} else {
				value = *itemAttribute.Value
			}
		}

		var valStr string
		if value == nil {
			valStr = "nil"
		} else {
			valStr = fmt.Sprintf("%v", value)
		}

		sb.WriteString(fmt.Sprintf("|  %s%s: %v\n", reqString, attribute.Name, valStr))
	}

	sb.WriteString(fmt.Sprintf("%s\n", strings.Repeat("-", 20)))
	return sb.String()
}
