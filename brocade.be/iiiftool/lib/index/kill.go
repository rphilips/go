package index

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	"brocade.be/base/mumps"
)

// Remove a IIIF digest and its associated identifiers from the index database
func RemoveDigest(digest string) error {
	index, err := sql.Open("sqlite", iiifIndexDb)
	if err != nil {
		return fmt.Errorf("cannot open index database: %v", err)
	}
	defer index.Close()

	_, err = index.Exec("DELETE FROM indexes where digest=?", digest)
	if err != nil {
		return fmt.Errorf("cannot delete digest from SQLite index database: %v", err)
	}

	payload := make(map[string]string)
	rou := `d %GetIds^gbiiif(.RApayload,"` + digest + `",1)`
	oreader, _, err := mumps.Reader(rou, payload)
	if err != nil {
		return fmt.Errorf("mumps reader error:\n%s", err)
	}

	out, err := ioutil.ReadAll(oreader)
	if err != nil {
		return fmt.Errorf("error reading MUMPS response:\n%s", err)
	}

	var result map[string]string
	var identifiers []string

	err = json.Unmarshal(out, &result)
	if err != nil {
		return fmt.Errorf("json error:\n%s", err)
	}

	for identifier := range result {
		identifiers = append(identifiers, identifier)
	}

	err = killinMIndex(digest, identifiers)
	if err != nil {
		return fmt.Errorf("error deleting digest from MUMPS index database: %v", err)
	}

	return nil
}

// Remove index info in MUMPS
func killinMIndex(digest string, identifiers []string) error {
	mpipe, err := mumps.Open("")
	if err != nil {
		return fmt.Errorf("mumps open error:\n%s", err)
	}
	defer mpipe.Close()

	cmds := []string{
		`k ^BIIIF("index",1,"digest2id","` + digest + `")`}

	for _, identifier := range identifiers {
		loi := strings.Split(identifier, ",")[0]
		loi = strings.TrimRight(loi, ",")
		cmds = append(cmds, `k ^BIIIF("index",1,"id2digest","`+loi+`","`+loi+`","`+digest+`")`)
		cmds = append(cmds, `k ^BIIIF("index",1,"id2digest","`+loi+`","`+identifier+`","`+digest+`")`)
	}

	for _, cmd := range cmds {
		err = mpipe.WriteExec(cmd)
		if err != nil {
			return fmt.Errorf("mumps exec error:\n%s", err)
		}
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
