# readrows

[![GoDoc](https://godoc.org/github.com/motemen/go-readrows?status.svg)](https://godoc.org/github.com/motemen/go-readrows)
[![Build Status](https://travis-ci.org/motemen/go-readrows.svg?branch=master)](https://travis-ci.org/motemen/go-readrows)

Package readrows provides Scan method to read results of database/sql
APIs to structs.

## Examples

### Scan

```go
type record struct {
    ID       int       `db:"id"`
    Text     string    `db:"text"`
    Bool     bool      `db:"bool"`
    DateTime time.Time `db:"dt"`
}

var r []record

rows, _ := db.Query("SELECT * FROM foo")

err := Scan(&r, rows)
if err != nil {
    panic(err)
}
```

## Author

motemen <https://github.com/motemen>
