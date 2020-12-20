package source

import (
	"bytes"
	"path/filepath"
	"strings"

	qregistry "brocade.be/base/registry"
	qofile "brocade.be/qtechng/lib/file/ofile"
	qobject "brocade.be/qtechng/lib/object"
	qutil "brocade.be/qtechng/lib/util"
)

// Env rekent de environment uit voor een source
func (source *Source) Env() map[string]string {
	env := make(map[string]string)
	qpath := source.String()
	qdir, base := qutil.QPartition(qpath)
	ext := filepath.Ext(base)
	env["%qpath"] = qpath
	env["%qdir"] = qdir
	env["%ext"] = ext
	env["%basename"] = base
	env["%version"] = source.Release().String()
	env["%mostype"] = qregistry.Registry["m-os-type"]
	env["%systemname"] = qregistry.Registry["system-name"]
	env["%systemgroup"] = qregistry.Registry["system-group"]
	env["%systemroles"] = qregistry.Registry["system-roles"]
	env["%os"] = qregistry.Registry["os"]
	env["%mclib"] = qregistry.Registry["m-clib"]
	project := source.Project()
	if project != nil {
		p := project.String()
		env["%project"] = p
		env["%qrelpath"] = strings.TrimPrefix(qpath, p+"/")
	}
	return env
}

// NotReplace bereknet de macro's die niet moeten worden vervangen
func (source *Source) NotReplace() []string {
	qpath := source.String()
	project := source.Project()
	p := project.String()
	qrelpath := strings.TrimPrefix(qpath, p+"/")
	config, err := project.LoadConfig()
	if err != nil {
		return []string{}
	}
	return config.ObjectsNotReplaced[qrelpath]
}

// Resolve vervangt alle r4/i4/m4/l4/i4
func (source *Source) Resolve(what string, objectmap map[string]qobject.Object, textmap map[string]string, buffer *bytes.Buffer) (err error) {
	if objectmap == nil {
		objectmap = make(map[string]qobject.Object)
	}
	body, err := source.Fetch()
	if err != nil {
		return err
	}

	if !bytes.Contains(body, []byte("4_")) {
		buffer.Write(body)
		return nil
	}
	nature := source.Natures()
	if !nature["text"] {
		buffer.Write(body)
		return nil
	}
	if nature["objfile"] {
		buffer.Write(body)
		return nil
	}
	env := source.Env()
	notreplace := source.NotReplace()

	_, err = ResolveText(env, body, what, notreplace, objectmap, textmap, buffer, "")
	if err != nil {
		body, _ = source.Fetch()
		buffer.Write(body)
		return err
	}
	return nil
}

// ResolveText vervangt in een byte slice - geassocieerd met een bestand
func ResolveText(env map[string]string, body []byte, what string, notreplace []string, objectmap map[string]qobject.Object, textmap map[string]string, buffer *bytes.Buffer, lgalgo string) (lastlgalgo string, err error) {
	lastlgalgo = lgalgo
	if what == "" {
		what = "rlmit"
	}
	if !bytes.Contains(body, []byte("4_")) {
		buffer.Write(body)
		return
	}
	r := env["%version"]
	split := qutil.ObjectSplitter(body)
	ssplit := make([]string, 0)
	for _, x := range split {
		ssplit = append(ssplit, "'"+string(x)+"'")
	}
	check := true
	t4y := strings.Contains(what, "t")
	r4y := strings.Contains(what, "r")
	i4y := strings.Contains(what, "i")
	m4y := strings.Contains(what, "m")
	l4y := strings.Contains(what, "l")
	if !r4y && !i4y && !m4y && !l4y && !t4y {
		buffer.Write(body)
		return
	}
	objs := make(map[string]bool)
	skip := make(map[string]bool)
	for _, piece := range split {
		check = !check
		if !check {
			continue
		}
		spiece := string(piece)
		if strings.HasPrefix(spiece, "r") {
			continue
		}
		if strings.HasPrefix(spiece, "t") {
			continue
		}
		if skip[spiece] {
			continue
		}
		_, ok := objectmap[spiece]
		if ok {
			continue
		}
		first := spiece[0:1]
		if !strings.Contains(what, first) {
			skip[spiece] = true
			continue
		}
		ok = false
		for _, ob := range notreplace {
			ok = ob == spiece
			if ok {
				break
			}
		}
		if ok {
			skip[spiece] = true
			continue
		}
		prefix := spiece[0:1]
		switch prefix {
		case "m":
			if m4y {
				objs[spiece] = true
			}
		case "i":
			if i4y {
				objs[spiece] = true
			}
		case "l":
			if l4y {
				objs[spiece] = true
			}
		}
	}

	objectlist := make([]qobject.Object, len(objs))
	count := -1
	for obj := range objs {
		count++
		prefix := obj[0:1]
		switch prefix {
		case "m":
			object := new(qofile.Macro)
			object.SetRelease(r)
			object.SetName(obj[3:])
			objectlist[count] = object
		case "i":
			object := new(qofile.Include)
			object.SetRelease(r)
			object.SetName(obj[3:])
			objectlist[count] = object
		case "l":
			if strings.HasPrefix(obj, "l4") && strings.Count(obj, "_") == 2 {
				parts := strings.SplitN(obj, "_", 3)
				obj = "l4_" + parts[2]
			}
			object := new(qofile.Lgcode)
			object.SetRelease(r)
			object.SetName(obj[3:])
			objectlist[count] = object
		}
	}

	for k, v := range qobject.FetchList(objectlist) {
		objectmap[k] = v
	}

	written := false
	check = true
	for i, piece := range split {
		check = !check
		if !check {
			if !written {
				buffer.Write(piece)
			}
			written = false
			continue
		}
		spiece := string(piece)
		if skip[spiece] {
			buffer.Write(piece)
			continue
		}
		// registry values
		if r4y && strings.HasPrefix(spiece, "r4_") {
			x := strings.ReplaceAll(spiece[3:], "_", "-")
			ender := strings.HasSuffix(x, "-")
			if ender {
				x = strings.TrimSuffix(x, "-")
			}
			y, ok := qregistry.Registry[x]
			if ok {
				if ender {
					z := qregistry.Registry["web-base-url"]
					if !strings.HasPrefix(y, z) {
						y = z + y
					}
				}
				if len(lgalgo) > 1 {
					algo := lgalgo[1:]
					y = qutil.ApplyAlgo(y, algo)
				}
				buffer.WriteString(y)
				continue
			}
			buffer.Write(piece)
			continue
		}

		if t4y && strings.HasPrefix(spiece, "t4_") {
			textid := spiece[3:]
			_, ok := textmap[textid]
			if !ok {
				buffer.Write(piece)
				continue
			}
		}

		// not usable objects

		if strings.HasPrefix(spiece, "l4_") && strings.Count(spiece, "_") == 1 {
			parts := strings.SplitN(spiece, "_", 2)
			spiece = parts[0] + "_" + lgalgo + "_" + parts[1]
		}
		obj := spiece
		if strings.HasPrefix(obj, "l4_") && strings.Count(obj, "_") == 2 {
			parts := strings.SplitN(obj, "_", 3)
			lgalgo = parts[1]
			obj = "l4_" + parts[2]
		}

		object, ok := objectmap[obj]
		if !ok || object == nil {
			buffer.Write(piece)
			continue
		}

		// i4
		if i4y && strings.HasPrefix(spiece, "i4_") {
			err = i4ResolveText(env, spiece, what, notreplace, objectmap, textmap, buffer, "")
			if err != nil {
				return
			}
			continue
		}
		// t4
		if t4y && strings.HasPrefix(spiece, "t4_") {
			err = t4ResolveText(env, spiece, what, notreplace, objectmap, textmap, buffer, "")
			if err != nil {
				return
			}
			continue
		}
		// l4
		if l4y && strings.HasPrefix(spiece, "l4_") {
			lgalgo, err = l4ResolveText(env, spiece, what, notreplace, objectmap, textmap, buffer, lgalgo)
			if err != nil {
				return
			}
			continue
		}
		// m4
		if m4y && strings.HasPrefix(spiece, "m4_") {
			err = m4ResolveText(env, spiece, string(split[i+1]), what, notreplace, objectmap, textmap, buffer, "")
			if err != nil {
				return
			}
			written = true
			continue
		}

	}
	return
}

// handles i4
func i4ResolveText(env map[string]string, include string, what string, notreplace []string, objectmap map[string]qobject.Object, textmap map[string]string, buffer *bytes.Buffer, lgalgo string) (err error) {
	object := objectmap[include].(*qofile.Include)
	content := []byte(object.Content)

	_, err = ResolveText(env, content, what, notreplace, objectmap, textmap, buffer, lgalgo)
	return err
}

// handles t4
func t4ResolveText(env map[string]string, text string, what string, notreplace []string, objectmap map[string]qobject.Object, textmap map[string]string, buffer *bytes.Buffer, lgalgo string) (err error) {

	content := []byte(textmap[text])
	_, err = ResolveText(env, content, what, notreplace, objectmap, textmap, buffer, lgalgo)
	return err
}

// handles l4
func l4ResolveText(env map[string]string, lgcode string, what string, notreplace []string, objectmap map[string]qobject.Object, textmap map[string]string, buffer *bytes.Buffer, lastlgalgo string) (lgalgo string, err error) {
	obj := lgcode
	if strings.HasPrefix(obj, "l4_") && strings.Count(obj, "_") == 1 {
		parts := strings.SplitN(obj, "_", 2)
		obj = parts[0] + lastlgalgo + "_" + parts[1]
	}
	parts := strings.SplitN(obj, "_", 3)
	obj = "l4_" + parts[2]
	lgalgo = parts[1]
	object := objectmap[obj].(*qofile.Lgcode)
	content := object.Replacer(nil, lgcode)
	if content == lgcode {
		buffer.WriteString(lgcode)
		return
	}
	what = strings.ReplaceAll(what, "m", "")
	what = strings.ReplaceAll(what, "i", "")
	what = strings.ReplaceAll(what, "t", "")
	lgalgo, err = ResolveText(env, []byte(content), what, notreplace, objectmap, textmap, buffer, lgalgo)
	return
}

// handles m4
func m4ResolveText(env map[string]string, macro string, extra string, what string, notreplace []string, objectmap map[string]qobject.Object, textmap map[string]string, buffer *bytes.Buffer, lgalgo string) (err error) {
	obj := macro
	object := objectmap[obj].(*qofile.Macro)
	args, rest, err := object.Args(extra)
	// if err != nil {
	// 	buffer.WriteString(macro)
	// 	buffer.WriteString(extra)
	// 	return err
	// }

	envex := make(map[string]string)

	for k, v := range env {
		envex[k] = v
	}

	for k, v := range args {
		envex[k] = v
	}

	calc := object.Replacer(envex, "")
	what = strings.ReplaceAll(what, "i", "")

	_, err = ResolveText(env, []byte(calc+rest), what, notreplace, objectmap, textmap, buffer, lgalgo)

	return err
}
