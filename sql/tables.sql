CREATE TABLE entities (
    id INTEGER NOT NULL PRIMARY KEY,
    name STRING NOT NULL UNIQUE,
    slug STRING NOT NULL UNIQUE
);

CREATE TABLE attributes (
    id INTEGER NOT NULL PRIMARY KEY,
    name STRING NOT NULL UNIQUE,
    slug STRING NOT NULL UNIQUE,
    type STRING NOT NULL,

    CHECK (type IN (
        'bool',
        'string',
        'int',
        'int8',
        'int16',
        'int32',
        'int64',
        'uint',
        'uint8',
        'uint16',
        'uint32',
        'uint64',
        'byte',
        'rune',
        'float32',
        'float64',
        'blob',
        'date',
        'time',
        'datetime'
    ))
);

CREATE TABLE entity_attribute (
    entity_id INTEGER NOT NULL REFERENCES entities(id) ON DELETE CASCADE ON UPDATE CASCADE,
    attribute_id INTEGER NOT NULL REFERENCES attributes(id) ON DELETE CASCADE ON UPDATE CASCADE,
    required BOOLEAN NOT NULL DEFAULT FALSE,

    PRIMARY KEY (entity_id, attribute_id)
);

CREATE TABLE items (
    id INTEGER NOT NULL PRIMARY KEY,
    entity_id INTEGER NOT NULL REFERENCES entities(id) ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE TABLE item_attribute (
    attribute_id INTEGER NOT NULL REFERENCES attributes(id) ON DELETE CASCADE ON UPDATE CASCADE,
    item_id INTEGER NOT NULL REFERENCES items(id) ON DELETE CASCADE ON UPDATE CASCADE,
    value ANY,

    PRIMARY KEY (item_id, attribute_id)
);
