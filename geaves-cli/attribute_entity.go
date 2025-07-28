package main

import (
	"context"
	"fmt"

	"github.com/Asfolny/geaves"
)

func linkAttributeEntityCommand(s state) error {
	if len(s.args) < 2 {
		// TODO print subcommand usage instead
		return fmt.Errorf("%s requires entity slug and attribute slug", s.cmdName)
	}

	entity, attribute, err := getEntityAndAttribute(s.queries, s.args[0], s.args[1])
	if err != nil {
		return err
	}

	_, err  = s.queries.CreateEntityAttribute(context.Background(), geaves.EntityAttribute{EntityID: entity.ID, AttributeID: attribute.ID})
	if err == nil {
		fmt.Printf("Succesfully linked %s to %s\n", attribute.Name, entity.Name)
	}

	return err
}

func linkRequiredAttributeEntityCommand(s state) error {
	if len(s.args) < 2 {
		// TODO print subcommand usage instead
		return fmt.Errorf("%s requires entity slug and attribute slug", s.cmdName)
	}

	entity, attribute, err := getEntityAndAttribute(s.queries, s.args[0], s.args[1])
	if err != nil {
		return err
	}

	_, err  = s.queries.CreateEntityAttribute(context.Background(), geaves.EntityAttribute{EntityID: entity.ID, AttributeID: attribute.ID, Required: true})
	if err == nil {
		fmt.Printf("Succesfully linked %s (required) to %s\n", attribute.Name, entity.Name)
	}

	return err
}

func requireEntityAttributeCommand(s state) error {
	if len(s.args) < 2 {
		// TODO print subcommand usage instead
		return fmt.Errorf("%s requires entity slug and attribute slug", s.cmdName)
	}

	entity, attribute, err := getEntityAndAttribute(s.queries, s.args[0], s.args[1])
	if err != nil {
		return err
	}

	err = s.queries.UpdateRequireEntityAttribute(context.Background(), true, entity.ID, attribute.ID)
	if err == nil {
		fmt.Printf("Succesfully make %s required on %s\n", attribute.Name, entity.Name)
	}

	return err
}

func optionalEntityAttributeCommand(s state) error {
	if len(s.args) < 2 {
		// TODO print subcommand usage instead
		return fmt.Errorf("%s requires entity slug and attribute slug", s.cmdName)
	}

	entity, attribute, err := getEntityAndAttribute(s.queries, s.args[0], s.args[1])
	if err != nil {
		return err
	}

	err = s.queries.UpdateRequireEntityAttribute(context.Background(), false, entity.ID, attribute.ID)
	if err == nil {
		fmt.Printf("Succesfully made %s optional on %s\n", attribute.Name, entity.Name)
	}

	return err

}

func unlinkAttributeEntityCommand(s state) error {
	if len(s.args) < 2 {
		// TODO print subcommand usage instead
		return fmt.Errorf("%s requires entity slug and attribute slug", s.cmdName)
	}

	entity, attribute, err := getEntityAndAttribute(s.queries, s.args[0], s.args[1])
	if err != nil {
		return err
	}

	err = s.queries.DeleteEntityAttribute(context.Background(), entity.ID, attribute.ID)
	if err == nil {
		fmt.Printf("Succesfully unlinked %s from %s\n", attribute.Name, entity.Name)
	}

	return err
}

func getEntityAndAttribute(queries *geaves.Queries, entitySlug string, attributeSlug string) (geaves.Entity, geaves.Attribute, error) {
	var entity geaves.Entity
	var attribute geaves.Attribute
	var err error

	entity, err = queries.GetEntity(context.Background(), geaves.GetEntityParam{WithAttributes: false, Field: "slug", Value: entitySlug})
	if err != nil {
		return entity, attribute, fmt.Errorf("Failed to get entity: %w", err)
	}

	attribute, err = queries.GetAttribute(context.Background(), geaves.GetAttributeParam{Field: "slug", Value: attributeSlug})
	if err != nil {
		return entity, attribute, fmt.Errorf("Failed to get attribute: %w", err)
	}

	return entity, attribute, nil
}
