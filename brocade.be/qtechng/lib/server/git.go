package server

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	qfs "brocade.be/base/fs"
	qregistry "brocade.be/base/registry"
)

func (release Release) InitGit() {
	qtechType := qregistry.Registry["qtechng-type"]
	if !strings.ContainsRune(qtechType, 'B') {
		return

	}

	if qregistry.Registry["qtechng-git-enable"] != "1" {
		return
	}
	if release.String() != "0.00" {
		return
	}

	// initialises mercurial repository
	cmd := exec.Command("git", "init", "--quiet")
	sourcedir, _ := release.FS("").RealPath("/source")
	cmd.Dir = sourcedir
	cmd.Run()

	cmd = exec.Command("git", "add", "--all")
	cmd.Dir = sourcedir
	cmd.Run()

	cmd = exec.Command("git", "commit", "--quiet", "--message", "Init")
	cmd.Dir = sourcedir
	cmd.Run()

	if qregistry.Registry["qtechng-backup-url"] == "" {
		return
	}

	url := qregistry.Registry["qtechng-backup-url"]

	format := `[core]
	repositoryformatversion = 0
	filemode = true
	bare = false
	logallrefupdates = true
[remote "backup"]
	url = %s
	fetch = +refs/heads/*:refs/remotes/backup/*
`
	configfile := filepath.Join(sourcedir, ".git", "config")
	qfs.Store(configfile, fmt.Sprintf(format, url), "qtech")

}
