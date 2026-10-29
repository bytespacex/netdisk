package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseVolumeSpec(t *testing.T) {
	cases := []struct {
		in, path, name, permsUsers string
		ok                         bool
	}{
		{`C:\data:docs:rwdm,admin`, `C:\data`, "docs", "rwdm,admin", true},
		{`/tmp/a:b:docs:rwdm,admin`, `/tmp/a:b`, "docs", "rwdm,admin", true},
		{`\\server\share:net:r,user`, `\\server\share`, "net", "r,user", true},
		{`~/documents:docs:rwdm,admin`, `~/documents`, "docs", "rwdm,admin", true},
		{`bad`, "", "", "", false},
		{`:name:r`, "", "", "", false},
		{`/tmp/vol:name:`, "", "", "", false},
	}
	for _, c := range cases {
		path, name, perms, ok := parseVolumeSpec(c.in)
		if path != c.path || name != c.name || perms != c.permsUsers || ok != c.ok {
			t.Errorf("parseVolumeSpec(%q) = (%q, %q, %q, %v), want (%q, %q, %q, %v)",
				c.in, path, name, perms, ok, c.path, c.name, c.permsUsers, c.ok)
		}
	}
}

func TestGetSafeName(t *testing.T) {
	cases := map[string]string{
		`../../CON.txt`: "_CON.txt",
		"foo.  ":        "foo",
		"a\x00b":        "a_b",
		"..":            "unnamed_file",
		"a#b?c.txt":     "a#b_c.txt",
	}
	for in, want := range cases {
		if got := getSafeName(in); got != want {
			t.Errorf("getSafeName(%q) = %q, want %q", in, got, want)
		}
	}
	if got := getSafeName(strings.Repeat("a", 300)); len(got) > 220 {
		t.Errorf("long name was not truncated: %d bytes", len(got))
	}
}

func TestURLPathEscape(t *testing.T) {
	got := urlPathEscape("vol/a#b?c d%中文.txt")
	want := "vol/a%23b%3Fc%20d%25%E4%B8%AD%E6%96%87.txt"
	if got != want {
		t.Errorf("urlPathEscape = %q, want %q", got, want)
	}
}

func TestPathInsideVolume(t *testing.T) {
	root := t.TempDir()
	inside := filepath.Join(root, "inside")
	outside := filepath.Join(root, "outside")
	if err := os.MkdirAll(inside, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(outside, 0755); err != nil {
		t.Fatal(err)
	}

	if !pathInsideVolume(inside, filepath.Join(inside, "new.txt")) {
		t.Error("new path inside volume was rejected")
	}
	if pathInsideVolume(inside, filepath.Join(inside, "..", "outside")) {
		t.Error("lexical escape was accepted")
	}

	link := filepath.Join(inside, "link")
	if err := os.Symlink(outside, link); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	if pathInsideVolume(inside, link) || pathInsideVolume(inside, filepath.Join(link, "secret")) {
		t.Error("symlink escape was accepted")
	}

	target := filepath.Join(inside, "target.txt")
	if err := os.WriteFile(target, []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}
	insideLink := filepath.Join(inside, "inside-link")
	if err := os.Symlink(target, insideLink); err != nil {
		t.Fatal(err)
	}
	if !pathInsideVolume(inside, insideLink) {
		t.Error("symlink to a file inside the volume was rejected")
	}
}

func TestGetFullAbsPath(t *testing.T) {
	root := t.TempDir()
	vol := filepath.Join(root, "vol")
	other := filepath.Join(root, "vol2")
	if err := os.MkdirAll(vol, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(other, 0755); err != nil {
		t.Fatal(err)
	}

	const name = "netdisk-test-volume"
	old, hadOld := volumes[name]
	volumes[name] = &Volume{Path: vol, Perms: "r", Users: []string{"u"}}
	t.Cleanup(func() {
		if hadOld {
			volumes[name] = old
		} else {
			delete(volumes, name)
		}
	})

	if got := getFullAbsPath(name + "/../vol2/secret.txt"); got != "" {
		t.Errorf("sibling escape accepted: %q", got)
	}
	if got := getFullAbsPath(name + "/sub/new.txt"); got != filepath.Join(vol, "sub", "new.txt") {
		t.Errorf("inside path = %q", got)
	}
}

func TestAttachmentDisposition(t *testing.T) {
	for _, name := range []string{"a b.zip", "中文.zip", "a;b.zip"} {
		if got := attachmentDisposition(name); !strings.HasPrefix(got, "attachment;") {
			t.Errorf("attachmentDisposition(%q) = %q", name, got)
		}
	}
}
