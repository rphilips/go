package fnmatch

import (
    "testing"
)

func TestMatch(t *testing.T) {

    type mytest struct {
        path     string
        pattern  string
        expected bool
        id       string
    }

    test := []mytest{{"/a/b/c/verb.exe", "*.exe", true, "1"}, {"/a/b/c/verb.exe", "/*/*.exe", true, "2"}}
    for _, tst := range test {
        if tst.expected {
            if !Match(tst.pattern, tst.path) {
                t.Errorf("Problem in %s", tst.id)
            }
        } else {
            if Match(tst.pattern, tst.path) {
                t.Errorf("Problem in %s", tst.id)
            }
        }

    }

}
