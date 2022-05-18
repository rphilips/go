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
	query := "SELECT * FROM indexes where key=40"
	rows, _ := db.Query(query)
	result, _ := ReadIndexRows(rows)
	expected := make([][]string, 0)

	expected = append(expected, []string{"40",
		"dg:ua:100",
		"5fd23dfc70d993af0da4e9b25c03766d45b66b32",
		"uact",
		"/library/database/iiif/23/b6/23b66b54d66730c52b9e4ad0fa399d07cfd32df5/db.sqlite",
		"2022-05-17T12:53:53+02:00",
		"2022-05-11T13:46:55+02:00"})
	util.Check(strings.Join(result[0], ""), strings.Join(expected[0], ""), t)
}

func TestReadMetaTable(t *testing.T) {
	var expected Meta
	expected.Key = "2"
	expected.Digest = "5fd23dfc70d993af0da4e9b25c03766d45b66b32"
	expected.Imgloi = "dg:ua:100"
	expected.Iiifsys = "uact"
	expected.Sortcode = " "
	expected.Indexes = "dg:ua:100^tg:uact:25"
	expected.Manifest = `{"@context":"http://iiif.io/api/presentation/3/context.json","homepage":[{"format":"text/html","id":"https://dev.anet.be/record/opacuactobj/tg:uact:25/N","label":{"en":["Catalogue record"],"nl":["Catalogus record"]},"type":"Text"}],"id":"https://dev.anet.be/iiif/5fd23dfc70d993af0da4e9b25c03766d45b66b32/manifest","items":[{"height":600,"id":"https://dev.anet.be/iiif/5fd23dfc70d993af0da4e9b25c03766d45b66b32/canvasbase/00000001","items":[{"id":"https://dev.anet.be/iiif/5fd23dfc70d993af0da4e9b25c03766d45b66b32/canvasbase/00000001/1","items":[{"body":{"format":"image/jpeg","id":"https://dev.anet.be/iiif/5fd23dfc70d993af0da4e9b25c03766d45b66b32/canvas/00000001","type":"Image"},"id":"https://dev.anet.be/iiif/5fd23dfc70d993af0da4e9b25c03766d45b66b32/canvasbase/00000001/1/image","motivation":"painting","target":"https://dev.anet.be/iiif/5fd23dfc70d993af0da4e9b25c03766d45b66b32/canvasbase/00000001","type":"Annotation"}],"type":"AnnotationPage"}],"label":{"en":["Image 00000001"],"fr":["Image 00000001"],"nl":["Image 00000001"]},"thumbnail":[{"format":"image/jpg","height":600,"id":"https://dev.anet.be/iiif/5fd23dfc70d993af0da4e9b25c03766d45b66b32/canvas/00000001","type":"Image","width":400}],"type":"Canvas","width":400},{"height":600,"id":"https://dev.anet.be/iiif/5fd23dfc70d993af0da4e9b25c03766d45b66b32/canvasbase/00000002","items":[{"id":"https://dev.anet.be/iiif/5fd23dfc70d993af0da4e9b25c03766d45b66b32/canvasbase/00000002/1","items":[{"body":{"format":"image/jpeg","id":"https://dev.anet.be/iiif/5fd23dfc70d993af0da4e9b25c03766d45b66b32/canvas/00000002","type":"Image"},"id":"https://dev.anet.be/iiif/5fd23dfc70d993af0da4e9b25c03766d45b66b32/canvasbase/00000002/1/image","motivation":"painting","target":"https://dev.anet.be/iiif/5fd23dfc70d993af0da4e9b25c03766d45b66b32/canvasbase/00000002","type":"Annotation"}],"type":"AnnotationPage"}],"label":{"en":["Image 00000002"],"fr":["Image 00000002"],"nl":["Image 00000002"]},"thumbnail":[{"format":"image/jpg","height":600,"id":"https://dev.anet.be/iiif/5fd23dfc70d993af0da4e9b25c03766d45b66b32/canvas/00000002","type":"Image","width":400}],"type":"Canvas","width":400}],"label":{"en":["[Anna van Sint-Bartholomeus]"],"nl":["[Anna van Sint-Bartholomeus]"]},"logo":[{"format":"image/svg+xml","height":50,"id":"https://dev.anet.be/test/uantwerpen.png","type":"Image","width":100}],"metadata":[{"label":{"en":["Title"],"nl":["Titel"]},"value":{"en":["[Anna van Sint-Bartholomeus]"],"nl":["[Anna van Sint-Bartholomeus]"]}},{"label":{"en":["Place"],"nl":["Plaats"]},"value":{"en":["Antwerpen [BE]"],"nl":["Antwerpen [BE]"]}},{"label":{"en":["Date"],"nl":["Datering"]},"value":{"en":["1610–1850"],"nl":["1610–1850"]}},{"label":{"en":["Workshop of"],"nl":["Atelier van"]},"value":{"en":["Cornelis I Galle (1576 - 1650)"],"nl":["Cornelis I Galle (1576 - 1650)"]}},{"label":{"en":["Printer"],"nl":["Drukker"]},"value":{"en":["Anoniem"],"nl":["Anoniem"]}},{"label":{"en":["Type of object"],"nl":["Object type"]},"value":{"en":[""],"nl":[""]}},{"label":{"en":["Technique"],"nl":["Techniek"]},"value":{"en":[""],"nl":[""]}},{"label":{"en":["Carrier"],"nl":["Drager"]},"value":{"en":[""],"nl":[""]}},{"label":{"en":["Genre"],"nl":["Genre"]},"value":{"en":[""],"nl":[""]}},{"label":{"en":["Collection"],"nl":["Collectie"]},"value":{"en":["UAntwerpen print collection"],"nl":["Prentenkabinet UAntwerpen"]}},{"label":{"en":["Call number"],"nl":["Plaatskenmerk"]},"value":{"en":["RG PK: Thijs KP 1.26"],"nl":["RG PK: Thijs KP 1.26"]}},{"label":{"en":["Barcode"],"nl":["Streepjescode"]},"value":{"en":["A030209194148B"],"nl":["A030209194148B"]}}],"provider":[{"homepage":[{"format":"text/html","id":"https://uantwerpen.be/library","label":{"en":["Homepage - University of Antwerp Library"]},"type":"Text"}],"id":"https://uantwerpen.be/ ","label":{"en":["University of Antwerp Library"]},"logo":[{"format":"image/svg+xml","height":50,"id":"https://dev.anet.be/test/uantwerpen.png","type":"Image","width":50}],"type":"Agent"}],"rendering":[{"format":"application/pdf","id":"https://dev.anet.be/docman/uact/9e67e7/1.pdf","label":{"en":["Download as PDF file"],"fr":["Télécharger en fichier pdf"],"nl":["Download als PDF bestand"]},"type":"Dataset"},{"id":"https://dev.anet.be/iiif/5fd23dfc70d993af0da4e9b25c03766d45b66b32/sqlite","label":{"en":["SQLite"],"fr":["SQLite"],"nl":["SQLite"]},"type":"Dataset"}],"requiredStatement":{"label":{"en":["Attribution"],"nl":["Attributie"]},"value":{"en":["Provided by University of Antwerp Libraries"],"nl":["Beschikbaar gesteld door de bibliotheek van UAntwerpen"]}},"rights":"http://creativecommons.org/licenses/by/4.0/","seeAlso":[{"format":"text/xml","id":"https://dev.anet.be/oai/thing/server.phtml?verb=GetRecord\u0026metadataPrefix=thingdc\u0026identifier=tg:uact:25","label":{"en":["Description in Dublin Core XML"],"nl":["Beschrijving in Dublin Core XML"]},"type":"Dataset"}],"summary":{"en":["[Anna van Sint-Bartholomeus]"],"nl":["[Anna van Sint-Bartholomeus]"]},"type":"Manifest"}`
	result, _ := ReadMetaTable(testDB)
	if expected != result {
		t.Errorf(fmt.Sprintf("\nResult: \n[%v]\nExpected: \n[%v]\n", result, expected))
	}
}
