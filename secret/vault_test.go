package secret

import "testing"

const (
	expectError = true
	expectOk    = false
)

type testCase struct {
	name               string
	path               string
	expectedError      bool
	expectedMountPath  string
	expectedSecretPath string
	expectedKey        string
}

var testTable = []testCase{
	{"empty", "", expectError, "", "", ""},
	{"blank", "   ", expectError, "", "", ""},
	{"one slash", "/", expectError, "", "", ""},
	{"two slashes", "//", expectError, "", "", ""},
	{"three slashes", "///", expectError, "", "", ""},
	{"one slash with spaces", " / ", expectError, "", "", ""},
	{"two slashes with spaces", " / / ", expectError, "", "", ""},
	{"three slashes with spaces", " / / / ", expectError, "", "", ""},
	{"one word", "a", expectError, "", "", ""},
	{"two words", "a/b", expectError, "", "", ""},
	{"three words", "a/b/c", expectOk, "a", "b", "c"},
	{"four words", "a/b/c/d", expectOk, "a", "b/c", "d"},
	{"one word with spaces", " a ", expectError, "", "", ""},
	{"two words with spaces", " a / b ", expectError, "", "", ""},
	{"three words with spaces", " a / b / c ", expectOk, "a", "b", "c"},
	{"four words with spaces", " a / b / c / d", expectOk, "a", "b / c", "d"},
	{"/ + one word", "/a", expectError, "", "", ""},
	{"/ + two words", "/a/b", expectError, "", "", ""},
	{"/ + three words", "/a/b/c", expectOk, "a", "b", "c"},
	{"/ + four words", "/a/b/c/d", expectOk, "a", "b/c", "d"},
	{"/ + one word with spaces", " / a ", expectError, "", "", ""},
	{"/ + two words with spaces", " / a / b ", expectError, "", "", ""},
	{"/ + three words with spaces", " / a / b / c ", expectOk, "a", "b", "c"},
	{"/ + four words with spaces", " / a / b / c / d ", expectOk, "a", "b / c", "d"},
}

func TestParseSecretPath(t *testing.T) {

	for i, data := range testTable {
		mountPath, secretPath, key, errParse := parseSecretPath(data.path)

		errored := errParse != nil

		if errored != data.expectedError {
			t.Errorf("%d/%d: %s: path='%s' unexpected error: %v",
				i+1, len(testTable), data.name, data.path, errParse)
			continue
		}

		if data.expectedMountPath != mountPath {
			t.Errorf("%d/%d: %s: path='%s' wrong mount path: expected=%s got=%s",
				i+1, len(testTable), data.name, data.path, data.expectedMountPath, mountPath)
		}

		if data.expectedSecretPath != secretPath {
			t.Errorf("%d/%d: %s: path='%s' wrong secret path: expected=%s got=%s",
				i+1, len(testTable), data.name, data.path, data.expectedSecretPath, secretPath)
		}

		if data.expectedKey != key {
			t.Errorf("%d/%d: %s: path='%s' key: expected=%s got=%s",
				i+1, len(testTable), data.name, data.path, data.expectedKey, key)
		}

	}
}
