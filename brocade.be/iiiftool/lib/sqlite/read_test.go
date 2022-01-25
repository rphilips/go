package sqlite

import (
	"database/sql"
	"fmt"
	"strings"
	"testing"
)

const testDB = "test_archive.sqlite"
const testindexDB = "test_index.sqlite"

func check(result interface{}, expected interface{}, t *testing.T) {
	if result != expected {
		t.Errorf(fmt.Sprintf("\nResult: \n[%v]\nExpected: \n[%v]\n", result, expected))
	}
}

func TestReadStringRow(t *testing.T) {
	db, _ := sql.Open("sqlite", testDB)
	defer db.Close()
	row := db.QueryRow("SELECT key FROM files where name=?", "00000001.jp2")
	result, _ := ReadStringRow(row)
	expected := "1"
	check(result, expected, t)
}

func TestReadIndexRows(t *testing.T) {
	db, _ := sql.Open("sqlite", testindexDB)
	defer db.Close()
	query := "SELECT * FROM indexes where key=1"
	rows, _ := db.Query(query)
	result, _ := ReadIndexRows(rows)
	expected := make([][]string, 0)
	expected = append(expected, []string{"1", "dg:ua:9", "e1e53b3d6b74c2e7ed0615ec687e68fdb61de242", "/library/database/iiif/24/2e/242ed16bdf86e786ce5160de7e2c47b6d3b35e1e/db.sqlite"})
	check(strings.Join(result[0], ""), strings.Join(expected[0], ""), t)
}
