package cmd

import (
	"archive/tar"
	"io"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	qfs "brocade.be/base/fs"
	qregistry "brocade.be/base/registry"
	qutil "brocade.be/qtechng/lib/util"
)

var systemSupportCmd = &cobra.Command{
	Use:     "support",
	Short:   "Update qtechng-support-dir",
	Long:    `Update the directory containing support files for QtechNG`,
	Args:    cobra.MaximumNArgs(1),
	Example: `qtechng system support`,
	RunE:    systemSupport,
	Annotations: map[string]string{
		"remote-allowed": "no",
		"with-qtechtype": "W",
	},
}

func init() {
	systemCmd.AddCommand(systemSupportCmd)
}

// Untar untars a tarball and writes the files to a target path
// Source: https://golangdocs.com/tar-gzip-in-golang
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
	tarURL := strings.Join([]string{qtechngSubDomain, "support.tar"}, "")

	fileURL, err := url.Parse(tarURL)
	if err != nil {
		log.Fatal("cmd/system_support/4:\n", err)
	}
	path := fileURL.Path
	fileSegments := strings.Split(path, "/")
	tarPath := filepath.Join(supportDir, fileSegments[len(fileSegments)-1])
	err = qfs.GetURL(tarURL, tarPath, "tempfile")
	if err != nil {
		log.Fatal("cmd/system_support/5:\n", err)
	}

	// Untar file
	err = Untar(tarPath, supportDir)
	if err != nil {
		log.Fatal("cmd/system_support/6:\n", err)
	}
	os.Remove(tarPath)

	// install vsix for vscode

	dir := filepath.Join(supportDir, "vscode")
	err = qutil.VSCode(dir)
	if err != nil {
		log.Fatal("cmd/system_support/7:\n", err)
	}

	return nil
}
