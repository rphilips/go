package index

import (
	"database/sql"
	"fmt"
	"time"

	"brocade.be/base/mumps"
	qtime "brocade.be/base/time"
	"brocade.be/iiiftool/lib/util"
)

// Write IIIF index information to SQLite index database
func SetSQLiteIndex(indexdata IndexData, db *sql.DB) error {

	insert, err := db.Prepare(`INSERT INTO indexes
	(key, loi, digest, iiifsys, location, metatime, sqlartime)
	Values($1,$2,$3,$4,$5,$6,$7)`)
	if err != nil {
		return fmt.Errorf("cannot prepare insert: %v", err)
	}
	defer insert.Close()

	lois := util.GetUniqueLOIs(indexdata.LOIs)

	for loi := range lois {
		_, err = insert.Exec(
			nil,
			loi,
			indexdata.Digest,
			indexdata.Iiifsys,
			indexdata.Location,
			indexdata.Metatime,
			indexdata.Sqlartime)
		if err != nil {
			// do not throw error, but allow to continue
			fmt.Printf("error executing insert: %v", err)
		}
	}

	return nil

}

// Log index info in MUMPS
func SetMIndex(indexdata IndexData) error {

	mpipe, err := mumps.Open("")
	if err != nil {
		return fmt.Errorf("mumps open error:\n%s", err)
	}
	defer mpipe.Close()

	lois := util.GetUniqueLOIs(indexdata.LOIs)

	for loi := range lois {
		digest := indexdata.Digest
		sortcode := indexdata.Sortcode
		iiifsys := indexdata.Iiifsys
		metaTimeObject, _ := time.Parse(time.RFC3339, indexdata.Metatime)
		sqlarTimeObject, _ := time.Parse(time.RFC3339, indexdata.Sqlartime)
		metatime := qtime.H(metaTimeObject)
		sqlartime := qtime.H(sqlarTimeObject)

		cmds := []string{
			`s ^BIIIF("index","` + loi + `","` + iiifsys + `","` + sortcode + `","` + digest + `")="` + metatime + `^` + sqlartime + `"`,
			`s ^BIIIF("index","` + digest + `","` + iiifsys + `","` + sortcode + `","` + loi + `")="` + metatime + `^` + sqlartime + `"`}

		for _, cmd := range cmds {
			err = mpipe.WriteExec(cmd)
			if err != nil {
				return fmt.Errorf("mumps exec error:\n%s", err)
			}
		}
	}

	return nil
}
