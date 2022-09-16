package tools

import (
	"crypto/sha512"
	"encoding/hex"
	"os"
)

func ImgProps(fname string, lineno int, mtime string, digest string) (mti string, dig string, err error) {

	fi, err := os.Stat(fname)
	if err != nil {
		return
	}
	mti = fi.ModTime().String()
	if mtime == mti {
		dig = digest
		return
	}

	blob, err := os.ReadFile(fname)
	if err != nil {
		return
	}
	sum := sha512.Sum512(blob)
	dig = hex.EncodeToString(sum[:28])
	return
}
