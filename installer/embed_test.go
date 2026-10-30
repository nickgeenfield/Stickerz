package main

import (
	"io/fs"
	"strings"
	"testing"
)

func TestPayloadContainsExpectedFiles(t *testing.T) {
	want := map[string]bool{
		"payload/install.sh":         false,
		"payload/uninstall.sh":       false,
		"payload/README.md":          false,
		"payload/src/stickerz.qmd":   false,
		"payload/library/index.json": false,
	}
	err := fs.WalkDir(payload, "payload", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			if _, ok := want[p]; ok {
				want[p] = true
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	for p, found := range want {
		if !found {
			t.Errorf("missing embedded file: %s", p)
		}
	}
	data, err := fs.ReadFile(payload, "payload/install.sh")
	if err != nil {
		t.Fatal(err)
	}
	s := string(data)
	for _, needle := range []string{"rebuild_hashtable", "stickerz.qmd", "NOT restart", "rm-stickers.qmd"} {
		if !strings.Contains(s, needle) {
			t.Errorf("install.sh missing %q", needle)
		}
	}
}
