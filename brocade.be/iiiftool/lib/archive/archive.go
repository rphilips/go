package archive

import (
	"fmt"
	"io"

	"brocade.be/base/docman"
	"brocade.be/base/fs"
	"brocade.be/iiiftool/lib/convert"
	"brocade.be/iiiftool/lib/iiif"
	qindex "brocade.be/iiiftool/lib/index"
	"brocade.be/iiiftool/lib/sqlite"
)

// Create (or update) IIIF archive from identifier and iiifsys
func Run(
	id string,
	iiifsys string,
	metaonly bool,
	index bool,
	cwd string,
	quality int,
	tile int,
	verbose bool) error {

	// Get IIIF metadata from MUMPS
	iiifMeta, err := iiif.Meta(id, iiifsys)
	if err != nil {
		return fmt.Errorf("iiiftool ERROR: %s", err)
	}

	// No digest could be created
	if iiifMeta.Info["digest"] == "" {
		return nil
	}

	sqlitefile := iiif.Digest2Location(iiifMeta.Info["digest"])

	iiifMeta.Iiifsys = iiifMeta.Info["iiifsys"]

	// Create SQLite contents

	if metaonly && fs.Exists(sqlitefile) {
		err = sqlite.ReplaceMeta(sqlitefile, iiifMeta)
		if err != nil {
			return fmt.Errorf("iiiftool ERROR: replace error:\n%s", err)
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
			docStream, errors := convert.ConvertDocmanIdsToJP2K(docIds, quality, tile, verbose)
			for _, e := range errors {
				if e != nil {
					return fmt.Errorf("iiiftool ERROR: conversion error:\n%s", e)
				}
			}
			convertedStream = docStream
		}

		// store file contents in SQLite archive
		err := sqlite.Create(sqlitefile, convertedStream, cwd, iiifMeta)
		if err != nil {
			return fmt.Errorf("iiiftool ERROR: store error:\n%s", err)
		}
	}

	// Update IIIF index
	if index {
		err = qindex.Update(sqlitefile)
		if err != nil {
			return fmt.Errorf("iiiftool ERROR: cannot update index:\n%s", err)
		}
	}

	return nil
}

// Update IIIF archive
func Update(digest string, id string, iiifsys string, verbose bool, dry bool) error {

	if verbose {
		fmt.Println(digest, id, iiifsys)
	}

	if dry {
		return nil
	}

	// image parameters can be 0 because there is never image conversion
	err := Run(id, iiifsys, true, true, "", 0, 0, false)
	if err != nil {
		return fmt.Errorf("iiiftool ERROR: cannot update archive: %v", err)
	}

	return nil
}
