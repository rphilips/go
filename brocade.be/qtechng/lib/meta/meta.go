package meta

import (
	"encoding/json"
	"sync"
	"time"

	qerror "brocade.be/qtechng/lib/error"
	qserver "brocade.be/qtechng/lib/server"
	qutil "brocade.be/qtechng/lib/util"
)

var metaCache = new(sync.Map)

//Meta modeleert een meta ingang
type Meta struct {
	Source string `json:"source"`
	Cu     string `json:"cu"`
	Mu     string `json:"mu"`
	Ct     string `json:"ct"`
	Mt     string `json:"mt"`
	It     string `json:"it"`
	Ft     string `json:"ft"`
	Digest string `json:"-"`
}

// New constructs a new Project
func (Meta) New(r string, s string) (meta *Meta, err error) {
	s = qutil.Canon(s)
	version, err := qserver.Release{}.New(r, true)
	if err != nil {
		e := &qerror.QError{
			Ref:     []string{"meta.new"},
			Version: r,
			File:    s,
			Msg:     []string{"Cannot instantiate version"},
		}
		err = qerror.QErrorTune(err, e)
		return
	}
	r = version.String()
	pid := r + " " + s
	met, _ := metaCache.Load(pid)
	if met != nil {
		return met.(*Meta), nil
	}

	fs, place := version.MetaPlace(s)
	blob, e := fs.ReadFile(place)
	meta = new(Meta)
	if e == nil {
		_ = json.Unmarshal(blob, meta)
	}
	met, _ = metaCache.LoadOrStore(pid, meta)
	return met.(*Meta), nil
}

// Store opslag van meta informatie
func (meta Meta) Store(r string, s string) (pmeta *Meta, err error) {
	pmeta = &meta
	s = qutil.Canon(s)
	version, err := qserver.Release{}.New(r, false)
	if err != nil {
		e := &qerror.QError{
			Ref:     []string{"meta.store.version"},
			Version: r,
			File:    s,
			Msg:     []string{"Cannot instantiate version"},
		}
		return pmeta, e
	}
	fs, place := version.MetaPlace(s)
	meta.Source = s
	changed, _, _, err := fs.Store(place, meta, "")
	if err != nil {
		e := &qerror.QError{
			Ref:     []string{"meta.store"},
			Version: r,
			File:    s,
			Msg:     []string{"Cannot store meta:`" + err.Error() + "'"},
		}
		return pmeta, e
	}
	if !changed {
		return pmeta, nil
	}
	pid := r + " " + s
	digest := meta.Digest
	meta.Digest = ""
	metaCache.Store(pid, &meta)
	meta.Digest = digest
	pmeta = &meta
	return
}

// Unlink opslag van meta informatie
func Unlink(r string, s string) (err error) {
	s = qutil.Canon(s)
	version, err := qserver.Release{}.New(r, false)
	if err != nil {
		e := &qerror.QError{
			Ref:     []string{"meta.unlink.version"},
			Version: r,
			File:    s,
			Msg:     []string{"Cannot instantiate version"},
		}
		return e
	}
	fs, place := version.MetaPlace(s)
	changed, err := fs.Waste(place)
	if err != nil {
		e := &qerror.QError{
			Ref:     []string{"meta.unlink"},
			Version: r,
			File:    s,
			Msg:     []string{"Cannot unlink meta:`" + err.Error() + "'"},
		}
		return e
	}
	if !changed {
		return
	}
	pid := r + " " + s
	metaCache.Delete(pid)
	return
}

// Update sets the meta information
func (meta *Meta) Update(met Meta) {
	h := time.Now()
	t := h.Format(time.RFC3339)

	if met.Ct != "" {
		meta.Ct = met.Ct
	}

	if met.Cu != "" {
		meta.Cu = met.Cu
	}

	if met.Mt != "" {
		meta.Mt = met.Mt
		if meta.Ct == "" {
			meta.Ct = meta.Mt
		}
	}

	if meta.Mt == "" {
		meta.Mt = t
	}
	if meta.Ct == "" {
		meta.Ct = meta.Mt
	}

	if met.Mu != "" {
		meta.Mu = met.Mu
		if meta.Cu == "" {
			meta.Cu = met.Mu
		}
	}

	if meta.Cu == "" {
		meta.Cu = "usystem"
	}

	if meta.Mu == "" {
		meta.Mu = meta.Cu
	}

}
