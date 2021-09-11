package util

import (
	"os"
	"strings"
)

func SaveEnvvars(envvarsave *map[string]string, envvars ...string) {
	// Process list of envvars specified
	for _, envvar := range envvars {
		_, exists := (*envvarsave)[envvar]
		if exists {
			continue
		}
		(*envvarsave)[envvar] = os.Getenv(envvar)
	}
}

func RestoreEnvvars(envvarsave *map[string]string, envvars ...string) {
	// Process list of envvars specified
	for _, envvar := range envvars {
		envvarval, exists := (*envvarsave)[envvar]
		if exists { // If doesn't exist in the map (i.e. not saved), ignore
			os.Setenv(envvar, envvarval)
			delete((*envvarsave), envvar) // Remove entry now that it is restored
		}
	}
}

// includeInEnvvar is a function that modifies a given envvar to contain the given element if it doesn't already have it. Returns
// true if it modified the envvar and false if the envvar already contained the element.
func IncludeInEnvvar(envvar, valueadd string) bool {
	var retval bool

	curval := os.Getenv(envvar)
	// Some special processing for certain envvars (only 1 now, may add others)
	switch envvar {
	case "ydb_routines":
		if curval == "" {
			curval = os.Getenv("gtmroutines")
		}
	}
	// Now see if value add is already part of the envvar value. If so, bypass modifying it.
	if !strings.Contains(curval, valueadd) {
		if curval != "" {
			curval = curval + " "
		}
		os.Setenv("ydb_routines", curval+valueadd)
		retval = true
	}
	return retval
}
