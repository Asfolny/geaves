package geaves

import (
	"errors"
	"fmt"
	"context"
	"encoding/json"
	"strings"
	"os"
)

type EntityAttributeEmbed struct {
	Attribute
	Required bool
}

type Entity struct {
	ID int64
	Name string
	Slug string
	attributes []EntityAttributeEmbed
	loadedAttributes bool
}

func (e *Entity) GetAttributes(ctx context.Context, q *Queries) ([]EntityAttributeEmbed, error) {
	var err error

	if !e.loadedAttributes {
		e.attributes, err = q.LoadAttributesByEntity(ctx, e.ID)
		e.loadedAttributes = true
	}

	return e.attributes, err
}

type EntityOption func(*Entity)

func WithAttribute(attr *Attribute) EntityOption {
	return func(e *Entity) {
		attrs := e.attributes
		attrs = append(attrs, EntityAttributeEmbed{*attr, false})
		e.attributes = attrs
	}
}

func WithRequiredAttribute(attr *Attribute) EntityOption {
	return func(e *Entity) {
		attrs := e.attributes
		attrs = append(attrs, EntityAttributeEmbed{*attr, true})
		e.attributes = attrs
	}
}

func NewEntity(name string, slug string, opts ...EntityOption) *Entity {
	entity := Entity{
		Name: name,
		Slug: slug,
	}

	for _, opt := range opts {
		opt(&entity)
	}

	return &entity
}

func (e *Entity) Save(q *Queries, ctx context.Context) error {
	// TODO split below between "Create" and "Update" depending on whether e.ID is 0 or > 0
	// They will be slightly similar, updated should also create missing attributes (attributes without ID) and mappings
	// but Update needs to look up
	if e.ID != 0 {
		// TODO extensive upgrade here
		return errors.New("Entity has already been saved")
	}

	entity, err := q.CreateEntity(ctx, CreateEntityParam{e.Name, e.Slug})

	if err != nil {
		return err
	}

	e.ID = entity.ID

	storedAttributes := make([]Attribute, len(e.attributes))
	for idx, attribute := range e.attributes {
		storedAttribute := attribute.Attribute

		if attribute.ID == 0 {
			storedAttribute, err = q.CreateAttribute(ctx, CreateAttributeParam{attribute.Name, attribute.Slug, attribute.Type})
			if err != nil {
				return fmt.Errorf("Failed creating new attribute: %w", err)
			}
		}

		storedAttributes[idx] = storedAttribute
	}

	mappedAttributes := make([]EntityAttributeEmbed, len(e.attributes))
	for idx, attr := range storedAttributes {
		entityAttr := EntityAttribute{
			EntityID: e.ID,
			AttributeID: attr.ID,
			Required: e.attributes[idx].Required,
		}

		mapping, err := q.CreateEntityAttribute(ctx, entityAttr)

		if err != nil {
			return fmt.Errorf("Failed creating entity->attribute map: %w", err)
		}

		mappedAttributes[idx] = EntityAttributeEmbed {
			Attribute: attr,
			Required: mapping.Required,
		}
	}

	e.attributes = mappedAttributes
	return err

}

const createEntity = `
INSERT INTO entities (name, slug) VALUES (?, ?)
RETURNING *;
`

type CreateEntityParam struct {
	Name string
	Slug string
}

func (q *Queries) CreateEntity(ctx context.Context, arg CreateEntityParam) (Entity, error) {
	row := q.db.QueryRowContext(ctx, createEntity,
		arg.Name,
		arg.Slug,
	)

	var i Entity
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Slug,
	)

	return i, err
}

const updateEntityName = `
UPDATE entities SET name = ? WHERE id = ?;
`

func (q *Queries) UpdateEntityName(ctx context.Context, name string, id int64) error {
	_, err := q.db.ExecContext(ctx, updateEntityName, name, id)
	return err
}

const updateEntitySlug = `
UPDATE entities SET slug = ? WHERE id = ?;
`

func (q *Queries) UpdateEntitySlug(ctx context.Context, slug string, id int64) error {
	_, err := q.db.ExecContext(ctx, updateEntitySlug, slug, id)
	return err
}

const deleteEntity = `
DELETE FROM entities WHERE id = ?;
`

func (q *Queries) DeleteEntity(ctx context.Context, id int64) error {
	_, err := q.db.ExecContext(ctx, deleteEntity, id)
	return err
}

const getEntityNoAttributes = `
SELECT id, name, slug, null FROM entities WHERE %s = ?;
`

const getEntityWithAttributes = `
SELECT
  entities.id,
  entities.name,
  entities.slug,
  IIF(entity_attribute.entity_id IS NOT NULL,
    JSON_GROUP_ARRAY(
      JSON_OBJECT('id', attributes.id, 'name', attributes.name, 'slug', attributes.slug, 'type', attributes.type, 'required', CAST(entity_attribute.required AS BOOLEAN))
    ),
  NULL) AS attributes
FROM entities
LEFT JOIN entity_attribute ON entity_attribute.entity_id = entities.id
INNER JOIN attributes ON entity_attribute.attribute_id = attributes.id
WHERE %s = ?;
`

type GetEntityParam struct {
	WithAttributes bool
	Field GetType
	Value any
}

func (q *Queries) GetEntity(ctx context.Context, arg GetEntityParam) (Entity, error) {
	switch arg.Field {
	case ByID:
		if _, ok := arg.Value.(int64); !ok {
			return Entity{}, errors.New("GetEntityParam's value has wrong type")
		}
		arg.Value = arg.Value.(int64)
		arg.Field = "entities.id"
		break;
	case BySlug:
		if _, ok := arg.Value.(string); !ok {
			return Entity{}, errors.New("GetEntityParam's vlaue has wrong type")
		}
		arg.Value = arg.Value.(string)
		arg.Field = "entities.slug"
		break;
	default:
		return Entity{}, fmt.Errorf("'%s' unsupported field to look for entity", arg.Field)
	}

	var query string
	if (arg.WithAttributes) {
		query = getEntityWithAttributes
	} else {
		query = getEntityNoAttributes
	}

	row := q.db.QueryRowContext(ctx, fmt.Sprintf(query, arg.Field), arg.Value)

	var i Entity
	var attributesJson *string
	if err := row.Scan(
	 	&i.ID,
		&i.Name,
		&i.Slug,
	 	&attributesJson,
	); err != nil {
		return i, err
	}

	if arg.WithAttributes {
		i.loadedAttributes = true

		if attributesJson != nil {
			attributes, err := parseAttributesJson(*attributesJson)
			if err != nil {
				return i, err
			}

			i.attributes = attributes
		}
	}

	return i, nil
}

const listEntitiesNoAttributes = `
SELECT id, name, slug, null FROM entities;
`

const listEntitiesWithAttributes = `
SELECT
  entities.id,
  entities.name,
  entities.slug,
  IIF(entity_attribute.entity_id IS NOT NULL,
    JSON_GROUP_ARRAY(
      JSON_OBJECT('id', attributes.id, 'name', attributes.name, 'slug', attributes.slug, 'type', attributes.type, 'required', CAST(entity_attribute.required AS BOOLEAN))
    ),
  NULL) AS attributes
FROM entities
LEFT JOIN entity_attribute ON entity_attribute.entity_id = entities.id
INNER JOIN attributes ON entity_attribute.attribute_id = attributes.id
GROUP BY entities.id;
`

func (q *Queries) ListEntities(ctx context.Context, withAttributes bool) ([]Entity, error) {
	var query string
	if withAttributes {
		query = listEntitiesWithAttributes
	} else {
		query = listEntitiesNoAttributes
	}

	rows, err := q.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []Entity
	for rows.Next() {
		var i Entity
		var attributesJson *string

		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.Slug,
			&attributesJson,
		); err != nil {
			return items, err
		}

		if withAttributes {
			i.loadedAttributes = true

			if attributesJson != nil {
				attributes, err := parseAttributesJson(*attributesJson)
				if err != nil {
					return items, err
				}

				i.attributes = attributes
			}
		}

		items = append(items, i)
	}

	return items, nil
}

const loadAttributesByEntity = `
SELECT IIF(entity_attribute.entity_id IS NOT NULL,
    JSON_GROUP_ARRAY(
      JSON_OBJECT('id', attributes.id, 'name', attributes.name, 'slug', attributes.slug, 'type', attributes.type, 'required', CAST(entity_attribute.required AS BOOLEAN))
    ),
  NULL) AS attributes
FROM entities
LEFT JOIN entity_attribute ON entity_attribute.entity_id = entities.id
INNER JOIN attributes ON entity_attribute.attribute_id = attributes.id
WHERE entities.id = ?;
`

func (q *Queries) LoadAttributesByEntity(ctx context.Context, id int64) ([]EntityAttributeEmbed, error) {
	row := q.db.QueryRowContext(ctx, loadAttributesByEntity, id)

	var attributesJson *string
	if err := row.Scan(
		&attributesJson,
	); err != nil {
		return nil, err
	}

	if attributesJson == nil {
		return nil, nil
	}

	attributes, err := parseAttributesJson(*attributesJson)
	if err != nil {
		return nil, err
	}

	return attributes, nil
}

func parseAttributesJson(attributesJson string) ([]EntityAttributeEmbed, error) {
	var parsedAttributes []struct{
		ID int64 `json:"id"`
		Name string `json:"name"`
		Slug string `json:"slug"`
		Type string `json:"type"`
		Required int `json:"required"`
	}

	if err := json.NewDecoder(strings.NewReader(attributesJson)).Decode(&parsedAttributes); err != nil {
		return nil, err
	}

	attributes := make([]EntityAttributeEmbed, len(parsedAttributes))
	for idx, parsedAttribute := range parsedAttributes {
		// TODO log this
		if !ValidAttributeType(parsedAttribute.Type) {
			fmt.Fprintf(os.Stderr, "Invalid type %s, skipping attribute %s\n", parsedAttribute.Type, parsedAttribute.Name)
			continue
		}

		attributes[idx] = EntityAttributeEmbed{
			Attribute: Attribute{
				ID: parsedAttribute.ID,
				Name: parsedAttribute.Name,
				Slug: parsedAttribute.Slug,
				Type: AttributeType(parsedAttribute.Type),
			},
			Required: parsedAttribute.Required > 0,
		}
	}

	return attributes, nil
}
