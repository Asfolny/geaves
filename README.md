# geaves
An zero dependency sqlite eav library for Go

## Requirements
- Go 1.24.5+

## Usage
### CLI
For demonstration purposes an (almost*) zero dependencies is provided under geaves-cli, this tool is meant to be used for 
CLI administration of the EAV once set up in a database

For further demo showcase, a geaves-cli/seed.sh is provided which will set up and run commands on it's own to provide an example usecase of 
the EAV library, along with printing it's final state

After installation, `geaves-cli` contains comprehensive `help` commands, simply run `./geaves-cli help` to find out how to use the program

Once you have `geaves-cli`, you will need to set an environment variable `GEAVE_CONNECTION` which contains the connection to open sqlite on

#### Installation
Currently using git itself is currently the easiest installation method, this will change in the future
```bash
$ git clone https://github.com/Asfolny/geaves.git
$ cd geaves/geaves-cli
$ go build -o geaves ./...
```

#### seed.sh
It is worth nothing that `seed.sh` will not exit on error
`seed.sh` is entirely self-contained and can just be run on it's own
```bash
$ ./seed.sh
```

`seed.sh` will leave a test.sql you can inspect with an `sqlite3 test.sql` if you so desire

`seed.sh` will also build geaves-cli into a binary called `geaves`

### geaves
To use `geaves` you must set up an sqlite database for it to be used on,
`geaves-cli` provides a handy command to generate the needed SQL, this has not been adopted to other ways of using a DB except raw sql and goose

The easiest way to get set up with the table, even on an existing database, is to simply run `geaves-cli generate | sqlite <database>`, (see geaves-cli/seed.sh for the full example)

geaves itself is a library has to be integrated into another system to shine, geaves-cli is an excellent example of how to do this

## Rationale
### But what _is_ an eav?

An **E**ntity **A**ttribute **V**alue system is a mapping of Entities (types) -> Attributes (properties) -> Values (data)

### Okay but who is this for?
An EAV is very useful for when you need system managers (who may not be programmers)

Or when significant time could be saved by allowing the data to give itself shape, rather than deploying database and code changes 
to update properteis of stored items

### Terminology
Not everything is as it seems, because I like my naming to be different, this EAV names things slightly different
- Entity
- Attribute
- EntityAttribute
- Item
- ItemAttribute

Entity and Attribute are self-describing, EntitiyAttribute is the "link" that puts that attribute on all Items that are of Entity (as an Entity is like a Type)

Item is the link between concrete data and Entity (it's data type if you will)

ItemAttribute is the Value that an Item _has_ and an Attribute _describes_

## Limitations
Due to Go's generic system, a programmer must in advance decide what type a generic must be before attempting to use it

This causes one large pain point in this library in particular, ItemAttributes (the V in EAV) are generic, 
and can be one of many types (see the docs (TODO LINK)), the programmer using this library must at least twice implement a large `switch ... case`, 
once to put data into the sqlite `value ANY` column and once again to pull data out of the `value ANY` column, 
so that ItemAttribute is properly types as to what the data actually is

## Attribution
`geaves` itself has 0 dependencies outside of Go 1.24.4 and the Go standard library, these are licensed under MIT, read [here](https://go.dev/LICENSE)

### geaves-cli
`geaves-cli` depends on `geaves` as well as `modernc.org/sqlite` for it's sqlite driver, which has a license of BSD 3 clause, read [here](https://gitlab.com/cznic/sqlite/-/blob/master/LICENS)
