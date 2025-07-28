package geaves

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"fmt"
)

type AttributeType string
const (
	BoolType AttributeType = "bool"
	StringType = "string"
	IntType = "int"
	Int8Type = "int8"
	Int16Type = "int16"
	Int32Type = "int32"
	Int64Type = "int64"
	UintType = "uint"
	Uint8Type = "uint8"
	Uint16Type = "uint16"
	Uint32Type = "uint32"
	Uint64Type = "uint64"
	ByteType = "byte"
	RuneType = "rune"
	Float32Type = "float32"
	Float64Type = "float64"
	BlobType = "blob"
	DateType = "date"
	TimeType = "time"
	DatetimeType = "datetime"
)

func ValidAttributeType(t string) bool {
	v := AttributeType(t)
	return BoolType == v ||
		StringType == v ||
		IntType == v ||
		Int8Type == v ||
		Int16Type == v ||
		Int32Type == v ||
		Int64Type == v ||
		UintType == v ||
		Uint8Type == v ||
		Uint16Type == v ||
		Uint32Type == v ||
		Uint64Type == v ||
		ByteType == v ||
		RuneType == v ||
		Float32Type == v ||
		Float64Type == v ||
		BlobType == v ||
		DateType == v ||
		TimeType == v ||
		DatetimeType == v
}

type attributeEntityEmbed struct {
	Entity
	Required bool
}

type Attribute struct {
	ID int64
	Name string
	Slug string
	Type AttributeType
	entities []attributeEntityEmbed
	loadedEntities bool
}

func (a *Attribute) GetEntities(ctx context.Context, q *Queries) ([]attributeEntityEmbed, error) {
	var err error

	if !a.loadedEntities {
		a.entities, err = q.LoadEntitiesByAttribute(ctx, a.ID)
		a.loadedEntities = true
	}

	return a.entities, err
}

func NewAttribute(name string, slug string, attrType AttributeType) *Attribute {
	attr := Attribute{
		Name: name,
		Slug: slug,
		Type: attrType,
	}

	return &attr
}

const createAttribute = `
INSERT INTO attributes (name, slug, type) VALUES (?, ?, ?)
RETURNING *;
`

type CreateAttributeParam struct {
	Name string
	Slug string
	Type AttributeType
}

func (q *Queries) CreateAttribute(ctx context.Context, arg CreateAttributeParam) (Attribute, error) {
	row := q.db.QueryRowContext(ctx, createAttribute,
		arg.Name,
		arg.Slug,
		arg.Type,
	)

	var i Attribute
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Slug,
		&i.Type,
	)

	return i, err
}

const updateAttributeName = `
UPDATE attributes SET name = ? WHERE id = ?;
`

func (q *Queries) UpdateAttributeName(ctx context.Context, name string, id int64) error {
	_, err := q.db.ExecContext(ctx, updateEntityName, name, id)
	return err
}

const updateAttributeSlug = `
UPDATE attributes SET slug = ? WHERE id = ?;
`

func (q *Queries) UpdateAttributeSlug(ctx context.Context, slug string, id int64) error {
	_, err := q.db.ExecContext(ctx, updateEntitySlug, slug, id)
	return err
}

const updateAttributeType = `
UPDATE attributes SET type = ? WHERE id = ?;
`

func (q *Queries) UpdateAttributeType(ctx context.Context, newType AttributeType, id int64) error {
	_, err := q.db.ExecContext(ctx, updateEntitySlug, newType, id)
	return err
}

const deleteAttribute = `
DELETE FROM attributes WHERE id = ?;
`

func (q *Queries) DeleteAttribute(ctx context.Context, id int64) error {
	_, err := q.db.ExecContext(ctx, deleteAttribute, id)
	return err
}

const getAttributeNoEntitie = `
SELECT id, name, slug, type, null FROM attributes WHERE %s = ?;
`

const getAttributesWithEntities = `
SELECT
  attributes.id,
  attributes.name,
  attributes.slug,
  attributes.type,
  IIF(entity_attribute.attribute.id IS NOT NULL
    JSON_GROUP_ARRAY(
      JSON_OBJECT('id', entities.id, 'name', entities.name, 'slug', entities.slug, 'required', entity_attribute.required)
    ), NULL
  ) AS entities
FROM attributes
LEFT JOIN entity_attribute ON entity_attribute.attribute_id = attributes.id
INNER JOIN entities ON entity_attribute.entity_id = entities.id
WHERE %s = ?;
`

type GetAttributeParam struct {
	WithEntities bool
	Field GetType
	Value any
}

func (q *Queries) GetAttribute(ctx context.Context, arg GetAttributeParam) (Attribute, error) {
	switch arg.Field {
	case ByID:
		if _, ok := arg.Value.(int64); !ok {
			return Attribute{}, errors.New("GetAttributeParam's value has wrong type")
		}
		arg.Value = arg.Value.(int64)
		arg.Field = "attributes.id"
		break;
	case BySlug:
		if _, ok := arg.Value.(string); !ok {
			return Attribute{}, errors.New("GetAttributeParam's vlaue has wrong type")
		}
		arg.Value = arg.Value.(string)
		arg.Field = "attributes.slug"
		break;
	default:
		return Attribute{}, fmt.Errorf("'%s' unsupported field to look for entity", arg.Field)
	}

	var query string
	if arg.WithEntities {
		query = getAttributesWithEntities
	} else {
		query = getAttributeNoEntitie
	}

	row := q.db.QueryRowContext(ctx, fmt.Sprintf(query, arg.Field), arg.Value)

	var i Attribute
	var entitiesJson *string
	if err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Slug,
		&i.Type,
		&entitiesJson,
	); err != nil {
		return i, err
	}

	if arg.WithEntities {
		i.loadedEntities = true

		if entitiesJson != nil {
			entities, err := parseEntitiesJson(*entitiesJson)
			if err != nil {
				return i, err
			}

			i.entities = entities
		}
	}

	return i, nil
}

const listAttributesNoEntities = `
SELECT id, name, slug, type, null FROM attributes
`

const listAttributesWithEntities = `
SELECT
  attributes.id,
  attributes.name,
  attributes.slug,
  attributes.type,
  IIF(entity_attribute.attribute_id IS NOT NULL
    JSON_GROUP_ARRAY(
      JSON_OBJECT('id', entities.id, 'name', entities.name, 'slug', entities.slug, 'required', entity_attribute.required)
    ), NULL
  ) AS entities
FROM attributes
LEFT JOIN entity_attribute ON entity_attribute.attribute_id = attributes.id
INNER JOIN entities ON entity_attribute.entity_id = entities.id
GROUP BY attributes.id;
`

func (q *Queries) ListAttributes(ctx context.Context, withEntities bool) ([]Attribute, error) {
	var query string
	if (withEntities) {
		query = listAttributesWithEntities
	} else {
		query = listAttributesNoEntities
	}

	rows, err := q.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []Attribute
	for rows.Next() {
		var i Attribute
		var entitiesJson *string

		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.Slug,
			&i.Type,
			&entitiesJson,
		); err != nil {
			return nil, err
		}

		if withEntities {
			i.loadedEntities = true

			if entitiesJson != nil {
				entities, err := parseEntitiesJson(*entitiesJson)
				if err != nil {
					return nil, err
				}

				i.entities = entities
			}
		}

		items = append(items, i)
	}

	return items, nil
}

const loadEntitiesByAttribute = `
SELECT IIF(entity_attribute.attribute_id IS NOT NULL,
  JSON_GROUP_ARRAY(
    JSON_OBJECT('id', entities.id, 'name', entities.name, 'slug', entities.slug, 'required', entity_attribute.required)
  ), NULL
) AS entities
FROM attributes
LEFT JOIN entity_attribute ON entity_attribute.attribute_id = attributes.id
INNER JOIN entities ON entity_attribute.entity_id = entities.id
WHERE attributes.id = ?;
`

func (q *Queries) LoadEntitiesByAttribute(ctx context.Context, id int64) ([]attributeEntityEmbed, error) {
	row := q.db.QueryRowContext(ctx, loadEntitiesByAttribute, id)

	var entitiesJson *string
	if err := row.Scan(
		&entitiesJson,
	); err != nil {
		return nil, err
	}

	entities, err := parseEntitiesJson(*entitiesJson)
	if err != nil {
		return nil, err
	}

	return entities, nil
}

func parseEntitiesJson(entitiesJson string) ([]attributeEntityEmbed, error) {
	var parsedEntities []struct{
		ID int64 `json:"id"`
		Name string `json:"name"`
		Slug string `json:"slug"`
		Required int `json:"required"`
	}

	if err := json.NewDecoder(strings.NewReader(entitiesJson)).Decode(&parsedEntities); err != nil {
		return nil, err
	}

	entities := make([]attributeEntityEmbed, len(parsedEntities))
	for idx, parsedEntity := range parsedEntities {
		entities[idx] = attributeEntityEmbed{
			Entity: Entity{
				ID: parsedEntity.ID,
				Name: parsedEntity.Name,
				Slug: parsedEntity.Slug,
			},
			Required: parsedEntity.Required > 0,
		}
	}

	return entities, nil
}
