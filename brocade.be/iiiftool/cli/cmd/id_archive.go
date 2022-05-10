package cmd

import (
	"io"
	"log"
	"strings"

	"brocade.be/base/docman"
	"brocade.be/base/fs"
	"brocade.be/iiiftool/lib/convert"
	"brocade.be/iiiftool/lib/iiif"
	"brocade.be/iiiftool/lib/index"
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
	Example: `iiiftool id archive c:stcv:12915850 --iiifsys=stcv --urlty=stcv`,
	RunE:    idArchive,
}

var Furlty string
var Fimgty string
var Faccess string
var Fmime string
var Fiiifsys string
var Findex bool
var Fmetaonly bool

func init() {
	idCmd.AddCommand(idArchiveCmd)
	idArchiveCmd.PersistentFlags().StringVar(&Furlty, "urlty", "", "URL type")
	idArchiveCmd.PersistentFlags().StringVar(&Fimgty, "imgty", "", "Image type")
	idArchiveCmd.PersistentFlags().StringVar(&Faccess, "access", "", "Access type")
	idArchiveCmd.PersistentFlags().StringVar(&Fmime, "mime", "", "Mime type")
	idArchiveCmd.PersistentFlags().StringVar(&Fiiifsys, "iiifsys", "test", "IIIF system")
	idArchiveCmd.PersistentFlags().IntVar(&Fquality, "quality", 70, "Quality parameter")
	idArchiveCmd.PersistentFlags().IntVar(&Ftile, "tile", 256, "Tile parameter")
	idArchiveCmd.PersistentFlags().BoolVar(&Findex, "index", true, "Rebuild IIIF index")
	idArchiveCmd.PersistentFlags().BoolVar(&Fverbose, "verbose", false, "Display information")
	idArchiveCmd.PersistentFlags().BoolVar(&Fmetaonly, "metaonly", false,
		`If images are present, only the meta information (including manifest) is replaced.
	If there are no images present, the usual archiving routine is performed.`)
}

func idArchive(cmd *cobra.Command, args []string) error {
	// Verify input
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

	// Get IIIF metadata from MUMPS
	iiifMeta, err := iiif.Meta(id, loiType, Furlty, Fimgty, Faccess, Fmime, Fiiifsys)
	if err != nil {
		log.Fatalf("iiiftool ERROR: %s", err)
	}

	iiifMeta.Info["olddigest"] = "" // to do: Richard past de MUMPS code aan eens de index er is
	var sqlitefile string
	if iiifMeta.Info["olddigest"] == "" {
		sqlitefile = iiif.TempLocation()
	} else {
		sqlitefile = iiif.Digest2Location(iiifMeta.Info["olddigest"])
	}

	iiifMeta.Iiifsys = Fiiifsys // to do: vroeger kwam dit uit MUMPS, nu niet meer?!

	// Create SQLite contents

	if Fmetaonly && fs.Exists(sqlitefile) {
		err = sqlite.ReplaceMeta(sqlitefile, iiifMeta)
		if err != nil {
			log.Fatalf("iiiftool ERROR: replace error:\n%s", err)
		}
	} else {

		imgLen := len(iiifMeta.Images)
		var convertedStream []io.Reader

		if imgLen > 0 {
			// get file contents from docman ids
			docIds := make([]docman.DocmanID, imgLen)
			for i, image := range iiifMeta.Images {
				docIds[i] = docman.DocmanID(image["loc"])
			}
			// convert docman files from TIFF/JPG to JP2K
			docStream, errors := convert.ConvertDocmanIdsToJP2K(docIds, Fquality, Ftile, Fverbose)
			for _, e := range errors {
				if e != nil {
					log.Fatalf("iiiftool ERROR: conversion error:\n%s", e)
				}
			}
			convertedStream = docStream
		}

		// store file contents in SQLite archive
		digest, err := sqlite.Create(sqlitefile, convertedStream, Fcwd, iiifMeta)
		if err != nil {
			log.Fatalf("iiiftool ERROR: store error:\n%s", err)
		}

		// move file if necessary
		if iiifMeta.Info["olddigest"] == "" {
			new := iiif.Digest2Location(digest)
			_, err := util.CreateFile(new)
			if err != nil {
				log.Fatalf("cannot create file: %v", err)
			}
			err = fs.CopyFile(sqlitefile, new, "=", true)
			if err != nil {
				log.Fatalf("cannot copy file: %v", err)
			}
			err = fs.Rmpath(sqlitefile)
			if err != nil {
				log.Fatalf("cannot remove tempfile: %v", err)
			}
			sqlitefile = new
		}
	}

	// Update IIIF index
	if Findex {
		err = index.Update(sqlitefile)
		if err != nil {
			log.Fatalf("iiiftool ERROR: cannot update index:\n%s", err)
		}
	}

	return nil
}
