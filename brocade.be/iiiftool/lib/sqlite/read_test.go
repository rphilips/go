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
	query := "SELECT * FROM indexes where key=1"
	rows, _ := db.Query(query)
	result, _ := ReadIndexRows(rows)
	expected := make([][]string, 0)

	expected = append(expected, []string{"1",
		"tg:uapr:1307",
		"14b785df369f240c6ebc157d9d8fcb131acc0910",
		"/library/database/iiif/01/90/0190cca131bcf8d9d751cbe6c042f963fd587b41/db.sqlite",
		"2022-02-16T16:55:06+01:00",
		"2022-02-16T16:55:06+01:00"})
	util.Check(strings.Join(result[0], ""), strings.Join(expected[0], ""), t)
}

func TestReadMetaTable(t *testing.T) {
	var expected Meta
	expected.Key = "4"
	expected.Identifier = "c:stcv:12915850/iiifsys=stcv/urlty=stcv"
	expected.Digest = "178f1d80dc33c4cea76cfdfb4c1b3598f88b55dc"
	expected.Imgloi = "c:stcv:12915850"
	expected.Indexes = "c:stcv:12915850^c:stcv:12915850,iiifsys:stcv,urlty:stcv"
	expected.Manifest = `{"@context":"http://iiif.io/api/presentation/3/context.json","behaviour":["paged"],"id":"https://dev.anet.be/iiif/178f1d80dc33c4cea76cfdfb4c1b3598f88b55dc/manifest","items":[{"height":"600","id":"https://dev.anet.be/iiif/178f1d80dc33c4cea76cfdfb4c1b3598f88b55dc/canvasbase/00000001","items":[{"id":"https://dev.anet.be/iiif/178f1d80dc33c4cea76cfdfb4c1b3598f88b55dc/canvasbase/00000001/1","items":[{"body":{"format":"image/jpeg","id":"https://dev.anet.be/iiif/178f1d80dc33c4cea76cfdfb4c1b3598f88b55dc/canvas/00000001","type":"Image"},"id":"https://dev.anet.be/iiif/178f1d80dc33c4cea76cfdfb4c1b3598f88b55dc/canvasbase/00000001/1/image","motivation":"painting","target":"https://dev.anet.be/iiif/178f1d80dc33c4cea76cfdfb4c1b3598f88b55dc/canvasbase/00000001","type":"Annotation"}],"type":"AnnotationPage"}],"label":{"en":["Image 00000001"],"fr":["Image 00000001"],"nl":["Image 00000001"]},"type":"Canvas","width":"400"},{"height":"600","id":"https://dev.anet.be/iiif/178f1d80dc33c4cea76cfdfb4c1b3598f88b55dc/canvasbase/00000002","items":[{"id":"https://dev.anet.be/iiif/178f1d80dc33c4cea76cfdfb4c1b3598f88b55dc/canvasbase/00000002/1","items":[{"body":{"format":"image/jpeg","id":"https://dev.anet.be/iiif/178f1d80dc33c4cea76cfdfb4c1b3598f88b55dc/canvas/00000002","type":"Image"},"id":"https://dev.anet.be/iiif/178f1d80dc33c4cea76cfdfb4c1b3598f88b55dc/canvasbase/00000002/1/image","motivation":"painting","target":"https://dev.anet.be/iiif/178f1d80dc33c4cea76cfdfb4c1b3598f88b55dc/canvasbase/00000002","type":"Annotation"}],"type":"AnnotationPage"}],"label":{"en":["Image 00000002"],"fr":["Image 00000002"],"nl":["Image 00000002"]},"type":"Canvas","width":"400"},{"height":"600","id":"https://dev.anet.be/iiif/178f1d80dc33c4cea76cfdfb4c1b3598f88b55dc/canvasbase/00000003","items":[{"id":"https://dev.anet.be/iiif/178f1d80dc33c4cea76cfdfb4c1b3598f88b55dc/canvasbase/00000003/1","items":[{"body":{"format":"image/jpeg","id":"https://dev.anet.be/iiif/178f1d80dc33c4cea76cfdfb4c1b3598f88b55dc/canvas/00000003","type":"Image"},"id":"https://dev.anet.be/iiif/178f1d80dc33c4cea76cfdfb4c1b3598f88b55dc/canvasbase/00000003/1/image","motivation":"painting","target":"https://dev.anet.be/iiif/178f1d80dc33c4cea76cfdfb4c1b3598f88b55dc/canvasbase/00000003","type":"Annotation"}],"type":"AnnotationPage"}],"label":{"en":["Image 00000003"],"fr":["Image 00000003"],"nl":["Image 00000003"]},"type":"Canvas","width":"400"},{"height":"600","id":"https://dev.anet.be/iiif/178f1d80dc33c4cea76cfdfb4c1b3598f88b55dc/canvasbase/00000004","items":[{"id":"https://dev.anet.be/iiif/178f1d80dc33c4cea76cfdfb4c1b3598f88b55dc/canvasbase/00000004/1","items":[{"body":{"format":"image/jpeg","id":"https://dev.anet.be/iiif/178f1d80dc33c4cea76cfdfb4c1b3598f88b55dc/canvas/00000004","type":"Image"},"id":"https://dev.anet.be/iiif/178f1d80dc33c4cea76cfdfb4c1b3598f88b55dc/canvasbase/00000004/1/image","motivation":"painting","target":"https://dev.anet.be/iiif/178f1d80dc33c4cea76cfdfb4c1b3598f88b55dc/canvasbase/00000004","type":"Annotation"}],"type":"AnnotationPage"}],"label":{"en":["Image 00000004"],"fr":["Image 00000004"],"nl":["Image 00000004"]},"type":"Canvas","width":"400"}],"label":{"label":{"en":["Cijfer bouck, inhoudende vele nieuwe, fraye, ende gherieuighe practijcken va[n] arithmetica / Adriaen vander Gucht. - Brugghe : Pieter de Clerck, 1569. - [4], 123 f. ; quarto ; +\u003csup\u003e4\u003c/sup\u003e A-2H\u003csup\u003e4\u003c/sup\u003e (lacks 2H4, blank?). - Fingerprint 156904 - # b1 A te : # b2 2H2 bijden$. -  Bibliographic reference: Cockx-Indestege, E. Belgica typographica 4623. - Sheet count: 31.75"],"fr":["Cijfer bouck, inhoudende vele nieuwe, fraye, ende gherieuighe practijcken va[n] arithmetica / Adriaen vander Gucht. - Brugghe : Pieter de Clerck, 1569. - [4], 123 f. ; quarto ; +\u003csup\u003e4\u003c/sup\u003e A-2H\u003csup\u003e4\u003c/sup\u003e (lacks 2H4, blank?). - Fingerprint 156904 - # b1 A te : # b2 2H2 bijden$. -  Référence bibliographique: Cockx-Indestege, E. Belgica typographica 4623. - Nombre de feuilles: 31.75"],"nl":["Cijfer bouck, inhoudende vele nieuwe, fraye, ende gherieuighe practijcken va[n] arithmetica / Adriaen vander Gucht. - Brugghe : Pieter de Clerck, 1569. - [4], 123 f. ; quarto ; +\u003csup\u003e4\u003c/sup\u003e A-2H\u003csup\u003e4\u003c/sup\u003e (lacks 2H4, blank?). - Fingerprint 156904 - # b1 A te : # b2 2H2 bijden$. -  Bibliografische referentie: Cockx-Indestege, E. Belgica typographica 4623. - Aantal vellen: 31.75"]}},"logo":[{"format":"image/svg+xml","height":50,"id":"https://anet.be/desktop/uantwerpennew/static/Banner_website_UAntwerpen_Bibliotheek_01.svg","type":"Image"}],"provider":[{"homepage":[{"format":"text/html","id":"https://uantwerpen.be/library","label":{"en":["Homepage - University of Antwerp Library"]},"type":"Text"}],"id":"https://uantwerpen.be/ ","label":{"en":["University of Antwerp Library"]},"logo":[{"format":"image/svg+xml","height":50,"id":"https://anet.be/desktop/uantwerpennew/static/Banner_website_UAntwerpen_Bibliotheek_01.svg","type":"Image"}],"type":"Agent"}],"requiredStatement":{"label":{"en":["Attribution"]},"value":{"en":["Provided courtesy of University of Antwerp Library"]}},"rights":"http://creativecommons.org/licenses/by/4.0/","type":"Manifest"}`
	result, _ := ReadMetaTable(testDB)
	if expected != result {
		t.Errorf(fmt.Sprintf("\nResult: \n[%v]\nExpected: \n[%v]\n", result.Manifest, expected.Manifest))
	}
}
