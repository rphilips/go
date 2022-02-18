package sqlite

import (
	"bytes"
	"database/sql"
	"fmt"
	"time"
)

// Function that reads a single string data sql.Row
func ReadStringRow(row *sql.Row) (string, error) {
	var data string
	err := row.Scan(&data)
	if err != nil {
		return data, err
	}
	return data, nil
}

// Function that reads a single sqlar sql.Row
func ReadSqlarRow(row *sql.Row, sqlar *Sqlar) error {
	var data []byte
	var mtime int64

	err := row.Scan(
		&sqlar.Name,
		&sqlar.Mode,
		&mtime,
		&sqlar.Sz,
		&data)
	if err != nil {
		return err
	}

	sqlar.Reader = bytes.NewReader(data)
	sqlar.Mtime = time.Unix(mtime, 0)

	return nil
}

// Function that reads a single meta sql.Row
func ReadMetaRow(row *sql.Row, meta *Meta) error {
	err := row.Scan(
		&meta.Key,
		&meta.Digest,
		&meta.Identifier,
		&meta.Indexes,
		&meta.Imgloi,
		&meta.Iiifsys,
		&meta.Manifest)
	if err != nil {
		return err
	}
	return nil
}

// Read IIIF meta table from archive
func ReadMetaTable(path string) (Meta, error) {
	var meta Meta

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return meta, fmt.Errorf("cannot open archive: %v", err)
	}
	defer db.Close()

	row := db.QueryRow("SELECT * FROM meta")

	err = ReadMetaRow(row, &meta)
	if err != nil {
		return meta, fmt.Errorf("cannot read meta: %v for file %s", err, path)
	}

	return meta, nil
}

// Function that reads multiple Index sql.Rows
func ReadIndexRows(rows *sql.Rows) ([][]string, error) {

	result := make([][]string, 0)

	// key|id|digest|location|metatime|sqlartime
	columns, err := rows.Columns()
	if err != nil {
		return result, err
	}

	for rows.Next() {

		data := make([]string, len(columns))
		err := rows.Scan(&data[0], &data[1], &data[2], &data[3], &data[4], &data[5])
		if err != nil {
			return result, err
		}

		result = append(result, data)
	}
	if err := rows.Err(); err != nil {
		return result, err
	}
	return result, nil
}
