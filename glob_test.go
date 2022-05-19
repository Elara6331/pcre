package pcre_test

import (
	"testing"

	"go.arsenm.dev/pcre"
)

func TestCompileGlob(t *testing.T) {
	r, err := pcre.CompileGlob("/**/bin")
	if err != nil {
		t.Fatal(err)
	}

	if !r.MatchString("/bin") {
		t.Error("expected /bin to match")
	}

	if !r.MatchString("/usr/bin") {
		t.Error("expected /usr/bin to match")
	}

	if !r.MatchString("/usr/local/bin") {
		t.Error("expected /usr/local/bin to match")
	}

	if r.MatchString("/usr") {
		t.Error("expected /usr not to match")
	}

	if r.MatchString("/usr/local") {
		t.Error("expected /usr/local not to match")
	}

	if r.MatchString("/home") {
		t.Error("expected /home not to match")
	}
}
