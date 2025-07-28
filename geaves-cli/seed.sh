#!/bin/sh
go build -o geaves ./...

table="$(pwd)/test.sql"
export GEAVES_CONNECTION="$table?_pragma=foreign_keys(1)"
./geaves generate | sqlite3 $table

./geaves entity create --name="Test" --slug=test

./geaves attribute create --name="B2" --slug=b2 --type bool
./geaves attribute create --name="B" --slug=b --type bool
./geaves attribute create --name="S" --slug=s --type string
./geaves attribute create --name="I" --slug=i --type int
./geaves attribute create --name="I8" --slug=i8 --type int8
./geaves attribute create --name="I16" --slug=i16 --type int16
./geaves attribute create --name="I32" --slug=i32 --type int32
./geaves attribute create --name="I64" --slug=i64 --type int64
./geaves attribute create --name="UI" --slug=ui --type uint
./geaves attribute create --name="UI8" --slug=ui8 --type uint8
./geaves attribute create --name="UI16" --slug=ui16 --type uint16
./geaves attribute create --name="UI32" --slug=ui32 --type uint32
./geaves attribute create --name="UI64" --slug=ui64 --type uint64
./geaves attribute create --name="0x" --slug=0x --type byte
./geaves attribute create --name="R" --slug=r --type rune
./geaves attribute create --name="F32" --slug=f32 --type float32
./geaves attribute create --name="F64" --slug=f64 --type float64
./geaves attribute create --name="Blob" --slug=blob --type blob
./geaves attribute create --name="date" --slug=date --type date
./geaves attribute create --name="time" --slug=time --type time
./geaves attribute create --name="datetime" --slug=datetime --type datetime

./geaves link test b2
./geaves linkreq test b
./geaves linkreq test s
./geaves linkreq test i
./geaves linkreq test i8
./geaves linkreq test i16
./geaves linkreq test i32
./geaves linkreq test i64
./geaves linkreq test ui
./geaves linkreq test ui8
./geaves linkreq test ui16
./geaves linkreq test ui32
./geaves linkreq test ui64
./geaves linkreq test 0x
./geaves linkreq test r
./geaves linkreq test f32
./geaves linkreq test f64
./geaves linkreq test blob
./geaves linkreq test date
./geaves linkreq test time
./geaves linkreq test datetime

./geaves item create test

./geaves item add 1 b 1
./geaves item add 1 s "testing string"
./geaves item add 1 i 2
./geaves item add 1 i8 3
./geaves item add 1 i16 4
./geaves item add 1 i32 5
./geaves item add 1 i64 6
./geaves item add 1 ui 7
./geaves item add 1 ui8 8
./geaves item add 1 ui16 9
./geaves item add 1 ui32 10
./geaves item add 1 ui64 11
./geaves item add 1 0x "a"
./geaves item add 1 r "b"
./geaves item add 1 f32 "1.2"
./geaves item add 1 f64 "1.3"
./geaves item add 1 blob "asjdkf"
./geaves item add 1 date "2025-01-01"
./geaves item add 1 time "01:02:03"
./geaves item add 1 datetime "2025-01-01 11:02:03"

./geaves entity list
./geaves entity info test
./geaves attribute list -e
./geaves item info 1
