package util

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"strconv"
)

// Function that takes a string as argument
// and returns the reverse of string.
func StrReverse(str string) (result string) {
	for _, v := range str {
		result = string(v) + result
	}
	return result
}

// Function that prepares gm conversion command arguments
func GmConvertArgs(quality int, tile int) []string {
	squality := strconv.Itoa(quality)
	stile := strconv.Itoa(tile)
	args := []string{"convert", "-flatten", "-quality", squality}
	args = append(args, "-define", "jp2:prg=rlcp", "-define", "jp2:numrlvls=7")
	args = append(args, "-define", "jp2:tilewidth="+stile, "-define", "jp2:tileheight="+stile)
	return args
}

// Function that reads a single string data sql.Row
func ReadStringRow(row *sql.Row) string {
	var data string
	err := row.Scan(&data)
	if err != nil {
		return data
	}
	return data
}

// Function that reads multiple sql.Rows
func ReadRows(rows *sql.Rows) ([]interface{}, error) {

	result := make([]interface{}, 0)

	columns, err := rows.Columns()
	if err != nil {
		return result, err
	}

	for rows.Next() {

		data := make([]*interface{}, len(columns))
		err := rows.Scan(&data[0], &data[1], &data[2])
		if err != nil {
			return result, err
		}

		values := make([]interface{}, 0)

		for _, item := range data {
			values = append(values, *item)
		}

		result = append(result, values)
	}
	if err := rows.Err(); err != nil {
		return result, err
	}
	return result, nil
}

// Function that standardizes image filenames
func ImageName(name string, index int) string {
	ext := filepath.Ext(name)
	base := fmt.Sprintf("%08d", index)
	return base + ext
}
