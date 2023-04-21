package pcre_test

import (
	"strings"
	"sync"
	"testing"

	"go.elara.ws/pcre"
)

func TestCompileError(t *testing.T) {
	r, err := pcre.Compile("(")
	if err == nil {
		t.Error("expected error, got nil")
	}
	defer r.Close()
}

func TestMatch(t *testing.T) {
	r := pcre.MustCompile(`\d+ (?=USD)`)
	defer r.Close()

	matches := r.MatchString("9000 USD")
	if !matches {
		t.Error("expected 9000 USD to match")
	}

	matches = r.MatchString("9000 RUB")
	if matches {
		t.Error("expected 9000 RUB not to match")
	}

	matches = r.MatchString("800 USD")
	if !matches {
		t.Error("expected 800 USD to match")
	}

	matches = r.MatchString("700 CAD")
	if matches {
		t.Error("expected 700 CAD not to match")
	}

	matches = r.Match([]byte("8 USD"))
	if !matches {
		t.Error("expected 8 USD to match")
	}
}

func TestMatchUngreedy(t *testing.T) {
	r := pcre.MustCompileOpts(`Hello, (.+)\.`, pcre.Ungreedy)
	defer r.Close()

	submatches := r.FindAllStringSubmatch("Hello, World. Hello, pcre2.", 1)
	if submatches[0][1] != "World" {
		t.Errorf("expected World, got %s", submatches[0][1])
	}

	matches := r.MatchString("hello, world.")
	if matches {
		t.Error("expected lowercase 'hello, world' not to match")
	}
}

func TestReplace(t *testing.T) {
	r := pcre.MustCompile(`(\d+)\.\d+`)
	defer r.Close()

	testStr := "123.54321 Test"

	newStr := r.ReplaceAllString(testStr, "${1}.12345")
	if newStr != "123.12345 Test" {
		t.Errorf(`expected "123.12345 Test", got "%s"`, newStr)
	}

	newStr = r.ReplaceAllString(testStr, "${9}.12345")
	if newStr != ".12345 Test" {
		t.Errorf(`expected ".12345 Test", got "%s"`, newStr)
	}

	newStr = r.ReplaceAllString(testStr, "${hi}.12345")
	if newStr != ".12345 Test" {
		t.Errorf(`expected ".12345 Test", got "%s"`, newStr)
	}

	newStr = r.ReplaceAllLiteralString(testStr, "${1}.12345")
	if newStr != "${1}.12345 Test" {
		t.Errorf(`expected "${1}.12345 Test", got "%s"`, newStr)
	}

	newStr = r.ReplaceAllStringFunc(testStr, func(s string) string {
		return strings.Replace(s, ".", ",", -1)
	})
	if newStr != "123,54321 Test" {
		t.Errorf(`expected "123,54321 Test", got "%s"`, newStr)
	}
}

func TestSplit(t *testing.T) {
	r := pcre.MustCompile("a*")
	defer r.Close()

	split := r.Split("abaabaccadaaae", 5)
	expected := [5]string{"", "b", "b", "c", "cadaaae"}

	if *(*[5]string)(split) != expected {
		t.Errorf("expected %v, got %v", expected, split)
	}

	split = r.Split("", 0)
	if split != nil {
		t.Errorf("expected nil, got %v", split)
	}

	split = r.Split("", 5)
	if split[0] != "" {
		t.Errorf(`expected []string{""}, got %v`, split)
	}
}

func TestFind(t *testing.T) {
	r := pcre.MustCompile(`(\d+)`)
	defer r.Close()

	testStr := "3 times 4 is 12"

	matches := r.FindAllString(testStr, -1)
	if len(matches) != 3 {
		t.Errorf("expected length 3, got %d", len(matches))
	}

	matches = r.FindAllString(testStr, 2)
	if len(matches) != 2 {
		t.Errorf("expected length 2, got %d", len(matches))
	}
	if matches[0] != "3" || matches[1] != "4" {
		t.Errorf("expected [3 4], got %v", matches)
	}

	index := r.FindStringIndex(testStr)
	if index[0] != 0 || index[1] != 1 {
		t.Errorf("expected [0 1], got %v", index)
	}

	submatch := r.FindStringSubmatch(testStr)
	if submatch[0] != "3" {
		t.Errorf("expected 3, got %s", submatch[0])
	}

	index = r.FindStringSubmatchIndex(testStr)
	if *(*[4]int)(index) != [4]int{0, 1, 0, 1} {
		t.Errorf("expected [0 1 0 1], got %v", index)
	}

	submatches := r.FindAllStringSubmatchIndex(testStr, 2)
	if len(submatches) != 2 {
		t.Errorf("expected length 2, got %d", len(submatches))
	}

	expected := [2][4]int{{0, 1, 0, 1}, {8, 9, 8, 9}}
	if *(*[4]int)(submatches[0]) != expected[0] {
		t.Errorf("expected %v, got %v", expected[0], submatches[0])
	}
	if *(*[4]int)(submatches[1]) != expected[1] {
		t.Errorf("expected %v, got %v", expected[0], submatches[0])
	}
}

func TestSubexpIndex(t *testing.T) {
	r := pcre.MustCompile(`(?P<number>\d)`)
	defer r.Close()

	index := r.SubexpIndex("number")
	if index != 1 {
		t.Errorf("expected 1, got %d", index)
	}

	index = r.SubexpIndex("string")
	if index != -1 {
		t.Errorf("expected -1, got %d", index)
	}

	num := r.NumSubexp()
	if num != 1 {
		t.Errorf("expected 1, got %d", num)
	}
}

func TestConcurrency(t *testing.T) {
	r := pcre.MustCompile(`\d*`)
	defer r.Close()

	wg := &sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()

		found := r.FindString("Test string 12345")
		if found != "12345" {
			t.Errorf("expected 12345, got %s", found)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		matched := r.MatchString("Hello")
		if matched {
			t.Errorf("Expected Hello not to match")
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		matched := r.MatchString("54321")
		if !matched {
			t.Errorf("expected 54321 to match")
		}
	}()

	wg.Wait()
}

func TestString(t *testing.T) {
	const expr = `()`

	r := pcre.MustCompile(expr)
	defer r.Close()

	if r.String() != expr {
		t.Errorf("expected %s, got %s", expr, r.String())
	}
}
