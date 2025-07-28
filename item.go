package geaves

import (
	"context"
	"time"
)

type Item struct {
	ID int64
	EntityID int64
	attributes []ItemAttribute[any]
	entity *Entity
	loadedEntity bool
	loadedAttributes bool
}

type ItemAttribute[T any] struct {
	ItemID int64
	AttributeID int64
	Type AttributeType
	Value T
}

func (i *Item) ChangeEntity(ctx context.Context, entity Entity, q *Queries) error {
	return i.ChangeEntityID(ctx, entity.ID, q)
}

func (i *Item) ChangeEntityID(ctx context.Context, entityId int64, q *Queries) error {
	err := q.UpdateItemEntityID(ctx, entityId, i.ID)
	if err != nil {
		return err
	}

	err = q.DeleteItemAttributesByItem(ctx, i.ID)
	if err != nil {
		return err
	}

	// TODO get "attributes" out of ctx and "add" them to the item

	i.EntityID = entityId
	return nil
}

func (i *Item) Delete(ctx context.Context, q *Queries) error {
	err := q.DeleteItemAttributesByItem(ctx, i.ID)
	if err != nil {
		return err
	}

	return q.DeleteItem(ctx, i.ID)
}

func (ia *ItemAttribute[T]) Create(ctx context.Context, q *Queries) error {
	_, err := q.db.ExecContext(ctx, "INSERT INTO item_attribute (item_id, attribute_id, value) VALUES(?, ?, ?);",
		ia.ItemID,
		ia.AttributeID,
		ia.Value,
	)
	return err
}

func (ia *ItemAttribute[T]) Update(ctx context.Context, q *Queries) error {
	_, err := q.db.ExecContext(ctx, "UPDATE item_attribute SET value = ? WHERE item_id = ? AND attribute_id = ?",
		ia.Value,
		ia.ItemID,
		ia.AttributeID,
	)
	return err
}

func (ia *ItemAttribute[T]) Load(ctx context.Context, q *Queries) error {
	row := q.db.QueryRowContext(ctx, "SELECT value FROM item_attribute WHERE item_id = ? AND attribute_id = ?", ia.ItemID, ia.AttributeID)

	switch(ia.Type) {
	case TimeType: fallthrough
	case DateType: fallthrough
	case DatetimeType:
		var i *string
		err := row.Scan(&i)
		if err != nil || i == nil {
			return err
		}

		time, err := time.Parse("2006-01-02 15:04:05 -0700 MST", *i)
		if err != nil {
			return err
		}

		// Generics bs handling...
		ia.Value = any(&time).(T)
		return nil
	default:
		return row.Scan(&ia.Value)
	}
}

const createItem = `
INSERT INTO items (entity_id) VALUES (?)
RETURNING *;
`

func (q *Queries) CreateItem(ctx context.Context, entityId int64) (Item, error) {
	row := q.db.QueryRowContext(ctx, createItem, entityId)

	var i Item
	err := row.Scan(
		&i.ID,
		&i.EntityID,
	)

	return i, err
}

const updateItemEntityID = `
UPDATE items SET entity_id = ? WHERE id = ?;
`

func (q *Queries) UpdateItemEntityID(ctx context.Context, entityId int64, itemId int64) error {
	_, err := q.db.ExecContext(ctx, updateItemEntityID, entityId, itemId)
	return err
}


const deleteItem = `
DELETE FROM items WHERE id = ?;
`

func (q *Queries) DeleteItem(ctx context.Context, id int64) error {
	_, err := q.db.ExecContext(ctx, deleteItem, id)
	return err
}

const getItem = `
SELECT * FROM items WHERE id = ?;
`

func (q *Queries) GetItem(ctx context.Context, id int64) (Item, error) {
	row := q.db.QueryRowContext(ctx, getItem, id)

	var i Item
	if err := row.Scan(
		&i.ID,
		&i.EntityID,
	); err != nil {
		return i, err
	}

	return i, nil
}

const listItems = `
SELECT * FROM items;
`

func (q *Queries) ListItems(ctx context.Context) ([]Item, error) {
	rows, err := q.db.QueryContext(ctx, listItems)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []Item
	for rows.Next() {
		var i Item

		if err := rows.Scan(
			&i.ID,
			&i.EntityID,
		); err != nil {
			return items, err
		}

		items = append(items, i)
	}

	return items, nil
}

const listItemAttributesByItem = `
SELECT item_id, attribute_id, value, type
FROM item_attribute
LEFT JOIN attributes ON attributes.id = item_attribute.attribute_id
WHERE item_id = ?;
`

func (q *Queries) ListItemAttributes(ctx context.Context, itemId int64) ([]ItemAttribute[*any], error) {
	rows, err := q.db.QueryContext(ctx, listItemAttributesByItem, itemId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []ItemAttribute[*any]
	for rows.Next() {
		var i ItemAttribute[*any]
		var attributeType string

		if err := rows.Scan(
			&i.ItemID,
			&i.AttributeID,
			&i.Value,
			&attributeType,
		); err != nil {
			return items, err
		}

		items = append(items, i)
	}

	return items, nil
}

const deleteItemAttribute = `
DELETE FROM item_attribute WHERE item_id = ? AND attribute_id = ?;
`

func (q *Queries) DeleteItemAttributes(ctx context.Context, itemId int64, attributeId int64) error {
	_, err := q.db.ExecContext(ctx, deleteItemAttributesByItem, itemId, attributeId)
	return err
}


const deleteItemAttributesByItem = `
DELETE FROM item_attribute WHERE item_id = ?
`

func (q *Queries) DeleteItemAttributesByItem(ctx context.Context, itemId int64) error {
	_, err := q.db.ExecContext(ctx, deleteItemAttributesByItem, itemId)
	return err
}

const deleteItemAttributesByAttribute = `
DELETE FROM item_attribute WHERE attribute_id = ?
`

func (q *Queries) DeleteItemAttributesByAttribute(ctx context.Context, attributeId int64) error {
	_, err := q.db.ExecContext(ctx, deleteItemAttributesByAttribute, attributeId)
	return err
}
