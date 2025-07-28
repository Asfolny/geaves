package geaves

import (
	"context"
)

type EntityAttribute struct {
	EntityID int64
	AttributeID int64
	Required bool
}

const createEntityAttribute = `
INSERT INTO entity_attribute (entity_id, attribute_id, required) VALUES (?, ?, ?)
RETURNING *;
`

func (q *Queries) CreateEntityAttribute(ctx context.Context, arg EntityAttribute) (EntityAttribute, error) {
	row := q.db.QueryRowContext(ctx, createEntityAttribute,
		arg.EntityID,
		arg.AttributeID,
		arg.Required,
	)

	var i EntityAttribute
	err := row.Scan(
		&i.EntityID,
		&i.AttributeID,
		&i.Required,
	)

	return i, err
}

const deleteEntityAttribute = `
DELETE FROM entity_attribute WHERE entity_id = ? AND attribute_id = ?;
`

func (q *Queries) DeleteEntityAttribute(ctx context.Context, entityId int64, attributeId int64) error {
	_, err := q.db.ExecContext(ctx, deleteEntityAttribute, entityId, attributeId)
	return err
}

const updatedRequireEntityAttribute = `
UPDATE entity_attribute SET required = ? WHERE entity_id = ? AND attribute_id = ?;
`

func (q *Queries) UpdateRequireEntityAttribute(ctx context.Context, req bool, entityId int64, attributeId int64) error {
	_, err := q.db.ExecContext(ctx, updatedRequireEntityAttribute, req, entityId, attributeId)
	return err
}

const deleteEntityAttributeByAttribute = `
DELETE FROM entity_attribute WHERE attribute_id = ?
`

func (q *Queries) DeleteEntityAttributeByAttribute(ctx context.Context, attributeId int64) error {
	_, err := q.db.ExecContext(ctx, deleteEntityAttributeByAttribute, attributeId)
	return err
}

const deleteEntityAttributeByEntity = `
DELETE FROM entity_attribute WHERE entity_id = ?
`

func (q *Queries) DeleteEntityAttributeByEntity(ctx context.Context, entityId int64) error {
	_, err := q.db.ExecContext(ctx, deleteEntityAttributeByEntity, entityId)
	return err
}
