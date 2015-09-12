// Package readrows provides Scan method to read results of database/sql
// APIs to structs.
package readrows

import (
	"database/sql"
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

func visitStructFields(t reflect.Type, cb func(f reflect.StructField)) {
	numField := t.NumField()
	for i := 0; i < numField; i++ {
		f := t.Field(i)
		if f.Anonymous {
			visitStructFields(f.Type, cb)
		} else {
			cb(f)
		}
	}
}

// Scan reads rows to a given pointer to slice v.
// The "db" tag in the struct type definition is used to map
// database columns to struct fields.
//
// v must be a pointer to a slice of structs.
func Scan(v interface{}, rows *sql.Rows) error {
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return err
	}

	sliceValue, itemType, itemIsPtr, err := resolveReflection(v)

	// column name to field name
	fieldMap := map[string]string{}

	visitStructFields(itemType, func(f reflect.StructField) {
		colName := f.Tag.Get("db")
		if colName == "" {
			colName = toSnakeCase(f.Name)
		}

		fieldMap[colName] = f.Name
	})

	for rows.Next() {
		// type: *elem
		item := reflect.New(itemType)

		dests := make([]interface{}, 0, len(cols))
		for _, col := range cols {
			if fn, ok := fieldMap[col]; ok {
				dest := item.Elem().FieldByName(fn).Addr().Interface()
				dests = append(dests, dest)
			} else {
				dests = append(dests, emptyScanner{})
			}
		}

		err := rows.Scan(dests...)
		if err != nil {
			return err
		}

		if itemIsPtr == false {
			// type: elem
			item = reflect.Indirect(item)
		}
		sliceValue.Set(reflect.Append(sliceValue, item))
	}

	return nil
}

// Postconditions if err == nil:
//   - sliceValue.CanSet() == true
//   - itemType.Kind() == reflect.Struct
func resolveReflection(v interface{}) (sliceValue reflect.Value, itemType reflect.Type, itemIsPtr bool, err error) {
	// type: *[]*elem or *[]elem
	rv := reflect.ValueOf(v)
	rt := rv.Type()

	if rt.Kind() != reflect.Ptr {
		err = fmt.Errorf("must be a pointer to a slice of struct: %T", v)
		return
	}

	// type: []*elem or []elem
	sliceValue = reflect.Indirect(rv)
	rt = rt.Elem()

	if rt.Kind() != reflect.Slice {
		err = fmt.Errorf("must be a pointer to a slice of struct: %T", v)
		return
	}

	// type: *elem or elem
	itemType = rt.Elem()
	if itemType.Kind() == reflect.Ptr {
		// type: elem
		itemType = itemType.Elem()
		itemIsPtr = true
	}

	// elem must be struct
	if itemType.Kind() != reflect.Struct {
		err = fmt.Errorf("must be a pointer to a slice of struct: %T", v)
		return
	}

	return
}

var re = regexp.MustCompile(`^([[:upper:]]*)([[:upper:]])([^[:upper:]]+)`)

func toSnakeCase(s string) string {
	parts := []string{}

	for {
		pos := re.FindStringSubmatchIndex(s)
		if pos == nil {
			break
		}

		if pos[2]+1 < pos[3] {
			parts = append(parts, s[pos[2]:pos[3]])
		}

		parts = append(parts, s[pos[4]:pos[7]])

		s = s[pos[1]:]
	}

	if len(s) > 0 {
		parts = append(parts, s)
	}

	return strings.ToLower(strings.Join(parts, "_"))
}

type emptyScanner struct{}

func (emptyScanner) Scan(value interface{}) error {
	return nil
}
