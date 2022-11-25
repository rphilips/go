package structure

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"

	vregistry "brocade.be/vchess/lib/registry"
	vstrings "brocade.be/vchess/lib/strings"
)

type Club struct {
	Short   string
	Full    string
	Stamno  string
	Url     string
	Site    string
	Address []string
	Contact []string
	Keys    []string
}

var AllClubs = make(map[string]*Club)

func (club Club) String() string {
	return club.Short
}

func (club *Club) Find(season *Season, prefix string) (err error) {
	if len(AllClubs) == 0 {
		fname := season.FName("clubs")
		f, err := os.Open(fname)
		if err != nil {
			return err
		}
		r := csv.NewReader(f)
		r.Comma = ';'
		r.FieldsPerRecord = -1
		oldlines, err := r.ReadAll()
		if err != nil {
			return err
		}
		f.Close()
		for _, line := range oldlines {
			if len(line) == 0 {
				continue
			}
			stamno := line[0]
			if stamno == "" {
				continue
			}
			_, err = strconv.Atoi(stamno)
			if err != nil {
				return err
			}
			keys := make([]string, 0)
			if len(line) > 3 {
				keys = line[3:]
			}
			club := Club{
				Short:  line[1],
				Full:   line[2],
				Stamno: stamno,
				Keys:   keys,
			}
			AllClubs[stamno] = &club
			short := strings.TrimRight(strings.ToUpper(strings.TrimSpace(club.Short)), "1234567890 -\t")
			if short != "" {
				AllClubs[short] = &club
			}
			full := strings.TrimRight(strings.ToUpper(strings.TrimSpace(club.Full)), "1234567890 -\t")
			if full != "" {
				AllClubs[full] = &club
			}
			for _, key := range club.Keys {
				key := strings.TrimRight(strings.ToUpper(strings.TrimSpace(key)), "1234567890 -\t")
				if key != "" {
					AllClubs[key] = &club
				}
			}
		}
	}

	_, e := strconv.Atoi(prefix)
	if e != nil {
		s := strings.TrimRight(strings.ToUpper(strings.TrimSpace(prefix)), "1234567890 -\t")
		if s == "" {
			return fmt.Errorf("unusable prefix `%s`", prefix)
		}
		prefix = s
	}

	if len(AllClubs) != 0 && AllClubs[prefix] == nil {
		for key, c := range AllClubs {
			if strings.HasPrefix(prefix, key) {
				AllClubs[prefix] = c
				break
			}
		}
	}

	x := AllClubs[prefix]

	if x == nil {
		return fmt.Errorf("club with prefix `%s` is not known", prefix)
	}
	if x != nil {
		club.Short = x.Short
		club.Full = x.Full
		club.Stamno = x.Stamno
		club.Url = x.Url
		club.Site = x.Site
		club.Address = append([]string{}, x.Address...)
		club.Contact = append([]string{}, x.Contact...)
		club.Keys = append([]string{}, x.Keys...)
		return
	}
	return

}

func (club *Club) URL() (url string) {
	url = vregistry.Registry["kbsb"].(map[string]any)["url"].(string)
	url = strings.ReplaceAll(url, "{stamno}", club.Stamno)
	club.Url = url
	return
}

func (club *Club) Load(round string) (err error) {
	if round != "" {
		season := new(Season)
		season.Init(nil)
		clubfile := season.ClubFile(club.Stamno)
		data, e := os.ReadFile(clubfile)
		if e == nil {
			info := make(map[string]any)
			json.Unmarshal(data, &info)
			r := "R" + strings.ReplaceAll(round, "R", "")
			rinfo, ok := info[r]
			if !ok {
				rinfo, ok = info["default"]
			}
			if ok {
				xinfo := rinfo.(map[string]any)
				site, ok := xinfo["site"]
				if ok {
					club.Site = site.(string)
				}
				address, ok := xinfo["address"]
				if ok {
					for _, line := range address.([]any) {
						club.Address = append(club.Address, line.(string))
					}
				}
				contact, ok := xinfo["contact"]
				if ok {
					for _, line := range contact.([]any) {
						club.Contact = append(club.Contact, line.(string))
					}
				}
				return nil
			}
		}
	}

	url := club.URL()
	res, err := http.Get(url)
	if err != nil {
		return err
	}
	content, err := io.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return err
	}
	page := vstrings.UTF8(content)
	site := ""

	nr := club.Stamno
	rex := regexp.MustCompile(`>` + nr + `:\s*([^<]+)</font><br>([^<]+)`)
	subs := rex.FindAllStringSubmatch(page, -1)
	if len(subs) == 0 || len(subs[0]) < 3 {
		return fmt.Errorf("names of club `%s` do not follow regexp", nr)
	}
	club.Short = subs[0][1]
	club.Full = subs[0][2]

	for _, x := range []string{"<td><b>Lokaal</b></td>", "<td><b>Local</b></td>"} {
		if strings.Contains(page, x) {
			_, y, _ := strings.Cut(page, x)
			y, _, _ = strings.Cut(y, "</tr>")
			y, _, _ = strings.Cut(y, "<tr>")
			y = strings.ReplaceAll(y, "<td>", "")
			y = strings.ReplaceAll(y, "</td>", "")
			y = strings.ReplaceAll(y, "<th>", "")
			y = strings.ReplaceAll(y, "</th>", "")
			y = strings.ReplaceAll(y, "<br>", "\n")
			y = strings.ReplaceAll(y, "<br />", "\n")
			y = strings.ReplaceAll(y, "<br/>", "\n")
			y = strings.TrimSpace(y)
			y = strings.ReplaceAll(y, "\n\n", "\n")
			site = strings.TrimSpace(y)
		}
		if site != "" {
			club.Site = site
			break
		}
	}
	address := ""
	for _, x := range []string{"<td><b>Adresse</b></td>", "<td><b>Adres</b></td>"} {
		if strings.Contains(page, x) {
			_, y, _ := strings.Cut(page, x)
			y, _, _ = strings.Cut(y, "</tr>")
			y, _, _ = strings.Cut(y, "<tr>")
			y = strings.ReplaceAll(y, "<td>", "")
			y = strings.ReplaceAll(y, "</td>", "")
			y = strings.ReplaceAll(y, "<th>", "")
			y = strings.ReplaceAll(y, "</th>", "")
			y = strings.ReplaceAll(y, "<br>", "\n")
			y = strings.ReplaceAll(y, "<br />", "\n")
			y = strings.ReplaceAll(y, "<br/>", "\n")
			y = strings.TrimSpace(y)
			y = strings.ReplaceAll(y, "\n\n", "\n")
			address = strings.TrimSpace(y)
		}
		if address != "" {
			for _, x := range strings.SplitN(address, "\n", -1) {
				x = strings.TrimSpace(x)
				if x == "" {
					continue
				}
				club.Address = append(club.Address, x)
			}
			if len(club.Address) != 0 {
				break
			}
		}
	}
	contact := ""
	for _, x := range []string{"<td><b>Verantwoordelijke<br>Interclubs KBSB</b></td>", "<td><b>Responsable des<br>Interclubs FRBE</b></td>"} {
		if strings.Contains(page, x) {
			_, y, _ := strings.Cut(page, x)
			y, _, _ = strings.Cut(y, "</tr>")
			y, _, _ = strings.Cut(y, "<tr>")
			y = strings.ReplaceAll(y, "<td>", "")
			y = strings.ReplaceAll(y, "</td>", "")
			y = strings.ReplaceAll(y, "<th>", "")
			y = strings.ReplaceAll(y, "</th>", "")
			y = strings.ReplaceAll(y, "<br>", "\n")
			y = strings.ReplaceAll(y, "<b>", "")
			y = strings.ReplaceAll(y, "</b>", "")
			y = strings.ReplaceAll(y, `<font color='red'>`, "")
			y = strings.ReplaceAll(y, `</font>`, "")
			y = strings.ReplaceAll(y, "<br />", "\n")
			y = strings.ReplaceAll(y, "<br/>", "\n")
			y = strings.TrimSpace(y)
			y = strings.ReplaceAll(y, "\n\n", "\n")
			contact = strings.TrimSpace(y)
		}
		if contact != "" {
			for _, x := range strings.SplitN(contact, "\n", -1) {
				x = strings.TrimSpace(x)
				if x == "" {
					continue
				}
				club.Contact = append(club.Contact, x)
			}
			if len(club.Contact) != 0 {
				break
			}
		}
	}
	return nil
}

func Update(season *Season) (err error) {

	fname := season.FName("clubs")
	f, err := os.Open(fname)
	if err != nil {
		return err
	}
	r := csv.NewReader(f)
	r.Comma = ';'
	r.FieldsPerRecord = -1
	oldlines, err := r.ReadAll()
	if err != nil {
		return err
	}
	f.Close()
	csvlines := make([][]string, 0)
	for _, line := range oldlines {
		if len(line) == 0 {
			continue
		}
		stamno := line[0]
		if stamno == "" {
			continue
		}
		_, err = strconv.Atoi(stamno)
		if err != nil {
			return err
		}
		keys := make([]string, 0)
		if len(line) > 3 {
			keys = line[3:]
		}
		club := Club{
			Stamno: stamno,
			Keys:   keys,
		}
		err = (&club).Load("")
		if err != nil {
			return
		}

		csvlines = append(csvlines, append([]string{club.Stamno, club.Short, club.Full}, keys...))
	}

	f, err = os.Create(fname)

	if err != nil {
		return fmt.Errorf("failed creating file: %s", err)
	}

	csvwriter := csv.NewWriter(f)
	csvwriter.Comma = ';'

	for _, row := range csvlines {
		csvwriter.Write(row)
	}
	csvwriter.Flush()
	f.Close()

	return
}
