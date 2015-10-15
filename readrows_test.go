package readrows

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"

	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
)

var db *sql.DB

func TestMain(m *testing.M) {
	tmpDir, err := ioutil.TempDir("", "go-readrows")
	if err != nil {
		log.Fatal(err)
	}

	db, err = sql.Open("sqlite3", filepath.Join(tmpDir, "test.db"))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec(`
CREATE TABLE foo (
	id INTEGER NOT NULL,
	text TEXT NOT NULL,
	bool BOOLEAN NOT NULL,
	dt DATETIME NOT NULL
)`)
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(`
INSERT INTO foo (id, text, bool, dt) VALUES
(1, "foo", 1, DATETIME()),
(2, "bar", 0, DATETIME())
`)
	if err != nil {
		log.Fatal(err)
	}

	os.Exit(m.Run())
}

var zeroTime time.Time

func TestScan(t *testing.T) {
	assert := assert.New(t)

	rows, err := db.Query("SELECT * FROM foo")
	if err != nil {
		t.Fatal(err)
	}
	if rows.Err() != nil {
		t.Fatal(rows.Err())
	}

	type record struct {
		ID       int       `db:"id"`
		Text     string    `db:"text"`
		Bool     bool      `db:"bool"`
		DateTime time.Time `db:"dt"`
	}

	var r []record

	err = Scan(&r, rows)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(2, len(r))

	assert.Equal(1, r[0].ID)
	assert.Equal("foo", r[0].Text)
	assert.Equal(true, r[0].Bool)
	assert.NotEqual(zeroTime, r[0].DateTime)

	assert.Equal(2, r[1].ID)
	assert.Equal("bar", r[1].Text)
	assert.Equal(false, r[1].Bool)
	assert.NotEqual(zeroTime, r[1].DateTime)

	t.Log(r)
}

func TestScan_Embedded(t *testing.T) {
	assert := assert.New(t)

	type TextBool struct {
		Text string `db:"text"`
		Bool bool   `db:"bool"`
	}
	type record struct {
		ID int `db:"id"`
		TextBool
		DateTime time.Time `db:"dt"`
	}

	rows, err := db.Query("SELECT * FROM foo WHERE id = 1")
	if err != nil {
		t.Fatal(err)
	}
	if rows.Err() != nil {
		t.Fatal(rows.Err())
	}

	var r []record

	err = Scan(&r, rows)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal("foo", r[0].Text)

	t.Logf("%+v", r)
}

func TestToSnakeCase(t *testing.T) {
	assert := assert.New(t)

	tests := map[string]string{
		"FooBar":    "foo_bar",
		"FooBarBaz": "foo_bar_baz",
		"URLPrefix": "url_prefix",
		"ID":        "id",
		"FooID":     "foo_id",
	}
	for f, t := range tests {
		assert.Equal(t, toSnakeCase(f))
	}
}

func ExampleScan() {
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
}
