package secret

import "testing"

type secretNameTest struct {
	testName          string
	prefix            string
	secretName        string
	expectRegion      string
	expectSecretName  string
	expectJsonField   string
	expectErrorResult bool
}

var secretNameTestTable = []secretNameTest{
	{"empty", "", "", "", "", "", expectError},
	{"bad prefix", "ttt", "aws-secretsmanager:region:name:json_field", "", "", "", expectError},
	{"secret1", "aws-secretsmanager", "aws-secretsmanager:region:name:json_field", "region", "name", "json_field", expectOk},
	{"secret2", "aws-secretsmanager", "aws-secretsmanager:region:name", "region", "name", "", expectOk},
	{"secret3", "aws-secretsmanager", "aws-secretsmanager|region|name|json_field", "region", "name", "json_field", expectOk},
}

func TestParseSecretName(t *testing.T) {

	for i, data := range secretNameTestTable {
		region, secretName, jsonField, errParse := parseSecretName(data.prefix, data.secretName)

		isError := errParse != nil

		if isError != data.expectErrorResult {
			t.Errorf("%d/%d: %s: unexpected error: got=%t expected=%t: %v",
				i+1, len(secretNameTestTable), data.testName, isError, data.expectErrorResult, errParse)
			continue
		}

		if region != data.expectRegion {
			t.Errorf("%d/%d: %s: region error: got=%s expected=%s",
				i+1, len(secretNameTestTable), data.testName, region, data.expectRegion)
		}

		if secretName != data.expectSecretName {
			t.Errorf("%d/%d: %s: secret name error: got=%s expected=%s",
				i+1, len(secretNameTestTable), data.testName, secretName, data.expectSecretName)
		}

		if jsonField != data.expectJsonField {
			t.Errorf("%d/%d: %s: json field error: got=%s expected=%s",
				i+1, len(secretNameTestTable), data.testName, jsonField, data.expectJsonField)
		}

	}
}
