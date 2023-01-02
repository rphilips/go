package next

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	btime "brocade.be/base/time"
	pregistry "brocade.be/pbladng/lib/registry"
)

func Special(id string) (send *time.Time, status string) {

	year, week, _ := strings.Cut(id, "-")

	specials := pregistry.Registry["year"].(map[string]any)
	_, ok := specials[year]
	if !ok {
		return nil, ""
	}

	_, ok = specials[week]
	if !ok && week < "52" {
		return nil, "holiday"
	}
	if !ok {
		return nil, "bad"
	}
	i, _ := strconv.Atoi(week)
	next := fmt.Sprintf("%02d", i+1)
	_, ok = specials[next]
	if !ok && next < "52" {
		send = btime.DetectDate(specials[week].(string))
		return send, "holiday1"
	}
	send = btime.DetectDate(specials[week].(string))
	return send, ""

}

func NextToNew(id string) (nextid string, date string) {
	yr, _, _ := strings.Cut(id, "-")
	jaar, _ := strconv.Atoi(yr)
	yr1 := strconv.Itoa(jaar + 1)
	for _, year := range []string{yr, yr1} {
		specials := pregistry.Registry["year"].(map[string]any)
		data, ok := specials[year].(map[string]any)
		if !ok {
			return "", ""
		}
		for i := 1; i < 54; i++ {
			j := fmt.Sprintf("%02d", i)
			d, ok := data[j]
			if !ok {
				continue
			}
			thisid := year + "-" + j
			if thisid <= id {
				continue
			}
			date = d.(string)
			date = date[:10]
			return thisid, date
		}
	}
	return "", ""
}

func MailDate(id string) (date string) {
	year, week, _ := strings.Cut(id, "-")
	specials := pregistry.Registry["year"].(map[string]any)
	data, ok := specials[year].(map[string]any)
	if !ok {
		return ""
	}
	d, ok := data[week]
	if !ok {
		return ""
	}
	date = d.(string)
	date = date[:10]
	return date
}
