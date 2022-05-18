package index

import (
	"database/sql"
	"fmt"
	"io/ioutil"

	"brocade.be/base/mumps"
)

// Wrapper for removing a IIIF digest and its associated lois from the index databases
func RemoveDigest(digest string) error {

	err := KillinSQLiteIndex(digest)
	if err != nil {
		return fmt.Errorf("error deleting digest from MUMPS index database: %v", err)
	}

	err = KillinMIndex(digest)
	if err != nil {
		return fmt.Errorf("error deleting digest from MUMPS index database: %v", err)
	}

	return nil
}

// Remove index info in SQLite
func KillinSQLiteIndex(digest string) error {
	index, err := sql.Open("sqlite", iiifIndexDb)
	if err != nil {
		return fmt.Errorf("cannot open index database: %v", err)
	}
	defer index.Close()

	_, err = index.Exec("DELETE FROM indexes where digest=?", digest)
	if err != nil {
		return fmt.Errorf("cannot delete digest from SQLite index database: %v", err)
	}

	return nil
}

// Remove index info in MUMPS
func KillinMIndex(digest string) error {

	payload := make(map[string]string)
	rou := `d %KillDigst^gbiiif(.RApayload,"` + digest + `",1)`
	oreader, _, err := mumps.Reader(rou, payload)
	if err != nil {
		return fmt.Errorf("mumps reader error:\n%s", err)
	}

	out, err := ioutil.ReadAll(oreader)
	if err != nil {
		return fmt.Errorf("error reading MUMPS response:\n%s\n%v", err, string(out))
	}

	return nil
}

// Remove M index completely
func KillMIndex() error {

	mpipe, err := mumps.Open("")
	if err != nil {
		return fmt.Errorf("mumps open error:\n%s", err)
	}
	defer mpipe.Close()

	cmd := `k ^BIIIF("index")`

	err = mpipe.WriteExec(cmd)
	if err != nil {
		return fmt.Errorf("mumps error:\n%s", err)
	}

	return nil
}
