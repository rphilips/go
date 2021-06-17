package cmd

import (
	"archive/tar"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	qfs "brocade.be/base/fs"
	qregistry "brocade.be/base/registry"
	"github.com/spf13/cobra"
)

var systemSupportCmd = &cobra.Command{
	Use:   "support",
	Short: "Update qtechng-support-dir",
	Long: `Update the directory containing support files for QtechNG.
With the flag --editor=vscode, the QtechNG extensions for VScode
are also downloaded and installed.`,
	Args: cobra.MaximumNArgs(1),
	Example: `qtechng system support
qtechng system support --editor=vscode`,
	RunE: systemSupport,
	Annotations: map[string]string{
		"remote-allowed": "no",
		"with-qtechtype": "W",
	},
}

func init() {
	systemSupportCmd.Flags().StringVar(&Feditor, "editor", "", "Editor used in development")
	systemCmd.AddCommand(systemSupportCmd)
}

// Untar untars a tarball and writes the files to a target path
// Source: https://golangdocs.com/tar-gzip-in-golang (modified)
func Untar(tarball, target string) error {
	reader, err := os.Open(tarball)
	if err != nil {
		return err
	}
	defer reader.Close()
	tarReader := tar.NewReader(reader)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		path := filepath.Join(target, header.Name)
		info := header.FileInfo()
		if info.IsDir() {
			if err = os.MkdirAll(path, info.Mode()); err != nil {
				return err
			}
			continue
		}

		file, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, info.Mode())
		if err != nil {
			return err
		}
		defer file.Close()
		_, err = io.Copy(file, tarReader)
		if err != nil {
			return err
		}
		// close file here to prevent lots of open files!
		file.Close()
	}
	return nil
}

func systemSupport(cmd *cobra.Command, args []string) error {

	qt := qregistry.Registry["qtechng-type"]
	onW := strings.ContainsRune(qt, 'W')
	onB := strings.ContainsRune(qt, 'B')
	onP := strings.ContainsRune(qt, 'P')
	if !onW || onB || onP {
		log.Fatal("cmd/system_support/1:\n", "command should be executed on workstation")
	}

	supportDir := qregistry.Registry["qtechng-support-dir"]
	if supportDir == "" {
		log.Fatal("cmd/system_support/2:\n", "registry qtechng-support-dir is empty or missing")
	}

	qtechngURL := qregistry.Registry["qtechng-url"]
	if qtechngURL == "" {
		log.Fatal("cmd/system_support/3:\n", "registry qtechng-url is empty or missing")
	}

	qSegments := strings.Split(qtechngURL, "/")
	binary := qSegments[len(qSegments)-1]
	qtechngSubDomain := strings.TrimRight(qtechngURL, binary)
	supportTar := "support.tar"
	supportLink := strings.Join([]string{qtechngSubDomain, supportTar}, "")
	supportPath := filepath.Join(supportDir, supportTar)
	err := qfs.GetURL(supportLink, supportPath, "tempfile")
	if err != nil {
		log.Fatal("cmd/system_support/5:\n", err)
	}

	err = Untar(supportPath, supportDir)
	if err != nil {
		log.Fatal("cmd/system_support/6:\n", err)
	}
	os.Remove(supportPath)

	if Feditor != "vscode" {
		return nil
	}

	vscode := qregistry.Registry["vscode-exe"]
	if vscode == "" {
		log.Fatal("cmd/system_support/7:\n", "registry qtechng-vscode-exe is empty or missing")
	}

	// workdir := qregistry.Registry["qtechng-work-dir"]
	workdir := qregistry.Registry["qtech-workstation-basedir"]
	dotvscode := filepath.Join(workdir, ".vscode")
	if !qfs.IsDir(dotvscode) {
		qfs.Mkdir(dotvscode, "process")
	}
	vscodeDir := filepath.Join(supportDir, "vscode")
	qfs.CopyFile(filepath.Join(vscodeDir, "tasks.json"), filepath.Join(dotvscode, "tasks.json"), "=", false)
	files, err := os.ReadDir(vscodeDir)
	if err != nil {
		log.Fatal("cmd/system_support/8:\n", err)
	}
	for _, file := range files {
		extensionPath := filepath.Join(vscodeDir, file.Name())
		if strings.HasSuffix(extensionPath, ".vsix") {
			extArgs := []string{"--install-extension", extensionPath, "--force"}
			extCmd := exec.Command(vscode, extArgs...)
			extCmd.Run()
		}
	}
	return nil
}
