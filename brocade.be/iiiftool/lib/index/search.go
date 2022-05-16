package index

import (
	"database/sql"
	"fmt"

	"brocade.be/iiiftool/lib/sqlite"
)

// Search the index database for a search string
func Search(search string) ([][]string, error) {
	result := make([][]string, 0)

	index, err := sql.Open("sqlite", iiifIndexDb)
	if err != nil {
		return result, fmt.Errorf("error opening index database: %v", err)
	}
	defer index.Close()

	query := "SELECT * FROM indexes where loi='" + search + "' or digest='" + search + "'"
	rows, err := index.Query(query)
	if err != nil {
		return result, fmt.Errorf("error querying index database: %v", err)
	}
	result, err = sqlite.ReadIndexRows(rows)
	if err != nil {
		return result, fmt.Errorf("error reading result: %v", err)
	}

	return result, nil
}

// Given a IIIF identifier, lookup its digest
// in the index database
func LookupId(identifier string) (string, error) {
	index, err := sql.Open("sqlite", iiifIndexDb)
	if err != nil {
		return "", fmt.Errorf("cannot open index database: %v", err)
	}
	defer index.Close()

	row := index.QueryRow("SELECT digest FROM indexes where loi=?", identifier)
	digest, err := sqlite.ReadStringRow(row)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if (err != nil) && (err != sql.ErrNoRows) {
		return "", fmt.Errorf("error selecting digest: %v", err)
	}

	return digest, nil
}

// Lookup all IIIF identifiers in the index database
func LookupAll() ([][]string, error) {
	result := make([][]string, 0)

	index, err := sql.Open("sqlite", iiifIndexDb)
	if err != nil {
		return result, fmt.Errorf("error opening index database: %v", err)
	}
	defer index.Close()

	query := "SELECT * FROM indexes"
	rows, err := index.Query(query)
	if err != nil {
		return result, fmt.Errorf("error querying index database: %v", err)
	}
	result, err = sqlite.ReadIndexRows(rows)
	if err != nil {
		return result, fmt.Errorf("error reading result: %v", err)
	}

	return result, nil
}
