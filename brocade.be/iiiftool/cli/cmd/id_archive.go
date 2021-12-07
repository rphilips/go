package cmd

import (
	"io"
	"log"
	"os/exec"
	"strings"

	"brocade.be/base/docman"
	"brocade.be/base/parallel"
	"brocade.be/iiiftool/lib/iiif"
	"brocade.be/iiiftool/lib/sqlite"
	"brocade.be/iiiftool/lib/util"

	"github.com/spf13/cobra"
)

var idArchiveCmd = &cobra.Command{
	Use:   "archive",
	Short: "Create archive for a IIIF identifier",
	Long: `Given a IIIF identifier, convert the appropriate image files
and store them in an SQLite archive.
Various additional parameters are in use and sometimes required:
--urlty:	url type (required for c-loi/o-loi)
--imgty:	image type (required for tg-loi)
--access:	access type (space separated)
--mime:		mime type (space separated)`,
	Args:    cobra.ExactArgs(1),
	Example: `iiiftool id archive dg:ua:1`,
	RunE:    idArchive,
}

var Furlty string
var Fimgty string
var Faccess string
var Fmime string
var Fiiifsys string

func init() {
	idCmd.AddCommand(idArchiveCmd)
	idArchiveCmd.PersistentFlags().StringVar(&Furlty, "urlty", "", "URL type")
	idArchiveCmd.PersistentFlags().StringVar(&Fimgty, "imgty", "", "Image type")
	idArchiveCmd.PersistentFlags().StringVar(&Faccess, "access", "", "Access type")
	idArchiveCmd.PersistentFlags().StringVar(&Fmime, "mime", "", "Mime type")
	idArchiveCmd.PersistentFlags().StringVar(&Fiiifsys, "iiif", "test", "IIIF system")
	idArchiveCmd.PersistentFlags().IntVar(&Fquality, "quality", 70, "quality parameter")
	idArchiveCmd.PersistentFlags().IntVar(&Ftile, "tile", 256, "tile parameter")
}

func idArchive(cmd *cobra.Command, args []string) error {
	// verify input
	id := args[0]
	if id == "" {
		log.Fatalf("iiiftool ERROR: argument is empty")
	}

	loiType := strings.Split(id, ":")[0]
	switch loiType {
	case "c", "o":
		if Furlty == "" {
			log.Fatalf("iiiftool ERROR: c-loi requires --urlty flag")
		}
	case "tg":
		if Fimgty == "" {
			log.Fatalf("iiiftool ERROR: tg-loi requires --imgty flag")
		}
	}

	// get IIIF metadata from MUMPS
	mResponse, err := iiif.Meta(id, loiType, Furlty, Fimgty, Faccess, Fmime, Fiiifsys)
	if err != nil {
		log.Fatalf("iiiftool ERROR: %s", err)
	}

	// get file contents from docman ids
	imgLen := len(mResponse.Images)
	originalStream := make([]io.Reader, imgLen)

	empty := true
	for i, image := range mResponse.Images {
		docid := docman.DocmanID(image["loc"])
		reader, err := docid.Reader()
		if err != nil {
			log.Fatalf("iiiftool ERROR: docman error:\n%s", err)
		}
		empty = false
		originalStream[i] = reader
	}
	if empty {
		log.Fatalf("iiiftool ERROR: no docman images found")
	}

	// convert file contents from TIFF/JPG to JP2K
	convertedStream := make([]io.Reader, imgLen)

	fn := func(n int) (interface{}, error) {
		old := originalStream[n]
		args := util.GmConvertArgs(Fquality, Ftile)
		// "Specify input_file as - for standard input, output_file as - for standard output",
		// so says http://www.graphicsmagick.org/GraphicsMagick.html#files,
		// but it needs to be "- jp2:-"!
		args = append(args, "-", "jp2:-")
		cmd := exec.Command("gm", args...)
		stdin, err := cmd.StdinPipe()
		if err != nil {
			return nil, err
		}
		out, err := cmd.StdoutPipe()
		if err != nil {
			return nil, err
		}
		_, err = cmd.StderrPipe()
		if err != nil {
			return nil, err
		}
		cmd.Start()

		go func() {
			defer stdin.Close()
			io.Copy(stdin, old)
		}()
		convertedStream[n] = out
		return out, nil
	}

	_, mapErr := parallel.NMap(len(originalStream), -1, fn)
	for _, e := range mapErr {
		if e != nil {
			log.Fatalf("iiiftool ERROR: conversion error:\n%s", e)
		}
	}

	// store file contents in SQLite archive
	sqlitefile := iiif.Digest2Location(mResponse.Digest)
	err = sqlite.Store(sqlitefile, convertedStream, Fcwd, mResponse)
	if err != nil {
		log.Fatalf("iiiftool ERROR: store error:\n%s", err)
	}

	return nil
}
