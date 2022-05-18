package sqlite

import (
	"database/sql"
	"fmt"
	"strings"
	"testing"

	"brocade.be/iiiftool/lib/util"
)

const testDB = "test_archive.sqlite"
const testindexDB = "test_index.sqlite"

func TestReadStringRow1(t *testing.T) {
	db, _ := sql.Open("sqlite", testDB)
	defer db.Close()
	row := db.QueryRow("SELECT key FROM files where name=?", "00000001.jp2")
	result, _ := ReadStringRow(row)
	expected := "1"
	util.Check(result, expected, t)
}

func TestReadStringRow2(t *testing.T) {
	db, _ := sql.Open("sqlite", testDB)
	defer db.Close()
	row := db.QueryRow("SELECT key FROM files where name=?", "helloworld")
	result, _ := ReadStringRow(row)
	expected := ""
	util.Check(result, expected, t)
}

func TestReadIndexRows(t *testing.T) {
	db, _ := sql.Open("sqlite", testindexDB)
	defer db.Close()
	query := "SELECT * FROM indexes where key=29"
	rows, _ := db.Query(query)
	result, _ := ReadIndexRows(rows)
	expected := make([][]string, 0)

	expected = append(expected, []string{"29",
		"dg:ua:100",
		"8ac2f1e4589df3741992331cb28baf7767fc1a44",
		"uapr",
		"/library/database/iiif/44/a1/44a1cf7677fab82bc1332991473fd9854e1f2ca8/db.sqlite",
		"2022-05-10T14:19:14+02:00",
		"2022-05-10T14:19:14+02:00"})
	util.Check(strings.Join(result[0], ""), strings.Join(expected[0], ""), t)
}

func TestReadMetaTable(t *testing.T) {
	var expected Meta
	expected.Key = "1"
	expected.Digest = "8ac2f1e4589df3741992331cb28baf7767fc1a44"
	expected.Imgloi = "dg:ua:100"
	expected.Iiifsys = "uapr"
	expected.Sortcode = " "
	expected.Indexes = "dg:ua:100^tg:uact:25"
	expected.Manifest = `{"@context":"http://iiif.io/api/presentation/3/context.json","id":"https://dev.anet.be/iiif/8ac2f1e4589df3741992331cb28baf7767fc1a44/manifest","items":[{"height":600,"id":"https://dev.anet.be/iiif/8ac2f1e4589df3741992331cb28baf7767fc1a44/canvasbase/00000001","items":[{"id":"https://dev.anet.be/iiif/8ac2f1e4589df3741992331cb28baf7767fc1a44/canvasbase/00000001/1","items":[{"body":{"format":"image/jpeg","id":"https://dev.anet.be/iiif/8ac2f1e4589df3741992331cb28baf7767fc1a44/canvas/00000001","type":"Image"},"id":"https://dev.anet.be/iiif/8ac2f1e4589df3741992331cb28baf7767fc1a44/canvasbase/00000001/1/image","motivation":"painting","target":"https://dev.anet.be/iiif/8ac2f1e4589df3741992331cb28baf7767fc1a44/canvasbase/00000001","type":"Annotation"}],"type":"AnnotationPage"}],"label":{"en":["Image 00000001"],"fr":["Image 00000001"],"nl":["Image 00000001"]},"thumbnail":[{"format":"image/jpg","height":600,"id":"https://dev.anet.be/iiif/8ac2f1e4589df3741992331cb28baf7767fc1a44/canvas/00000001","type":"Image","width":400}],"type":"Canvas","width":400},{"height":600,"id":"https://dev.anet.be/iiif/8ac2f1e4589df3741992331cb28baf7767fc1a44/canvasbase/00000002","items":[{"id":"https://dev.anet.be/iiif/8ac2f1e4589df3741992331cb28baf7767fc1a44/canvasbase/00000002/1","items":[{"body":{"format":"image/jpeg","id":"https://dev.anet.be/iiif/8ac2f1e4589df3741992331cb28baf7767fc1a44/canvas/00000002","type":"Image"},"id":"https://dev.anet.be/iiif/8ac2f1e4589df3741992331cb28baf7767fc1a44/canvasbase/00000002/1/image","motivation":"painting","target":"https://dev.anet.be/iiif/8ac2f1e4589df3741992331cb28baf7767fc1a44/canvasbase/00000002","type":"Annotation"}],"type":"AnnotationPage"}],"label":{"en":["Image 00000002"],"fr":["Image 00000002"],"nl":["Image 00000002"]},"thumbnail":[{"format":"image/jpg","height":600,"id":"https://dev.anet.be/iiif/8ac2f1e4589df3741992331cb28baf7767fc1a44/canvas/00000002","type":"Image","width":400}],"type":"Canvas","width":400}],"logo":[{"format":"image/svg+xml","height":50,"id":"https://dev.anet.be/test/uantwerpen.png","type":"Image","width":100}],"provider":[{"homepage":[{"format":"text/html","id":"https://uantwerpen.be/library","label":{"en":["Homepage - University of Antwerp Library"]},"type":"Text"}],"id":"https://uantwerpen.be/ ","label":{"en":["University of Antwerp Library"]},"logo":[{"format":"image/svg+xml","height":50,"id":"https://dev.anet.be/test/uantwerpen.png","type":"Image","width":50}],"type":"Agent"}],"rendering":[{"format":"application/pdf","id":"https://dev.anet.be/docman/uact/9e67e7/1.pdf","label":{"en":["Download as PDF file"],"fr":["Télécharger en fichier pdf"],"nl":["Download als PDF bestand"]},"type":"Dataset"},{"id":"https://dev.anet.be/iiif/8ac2f1e4589df3741992331cb28baf7767fc1a44/sqlite","label":{"en":["SQLite"],"fr":["SQLite"],"nl":["SQLite"]},"type":"Dataset"}],"requiredStatement":{"label":{"en":["Attribution"],"nl":["Attributie"]},"value":{"en":["Provided by University of Antwerp Libraries"],"nl":["Beschikbaar gesteld door de bibliotheek van UAntwerpen"]}},"rights":"http://creativecommons.org/licenses/by/4.0/","type":"Manifest"}`
	result, _ := ReadMetaTable(testDB)
	if expected != result {
		t.Errorf(fmt.Sprintf("\nResult: \n[%v]\nExpected: \n[%v]\n", result, expected))
	}
}
