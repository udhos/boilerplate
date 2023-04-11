// Package envconfig loads configuration from env vars.
package envconfig

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

/*
export DB_URI=#http::GET,https,tttt.lambda-url.us-east-1.on.aws,/,eyJwYXJhbWV0ZXIiOiJtb25nb2RiIn0=,Bearer secret:uri
#   Method: GET
# Protocol: https
#     Host: tttt.lambda-url.us-east-1.on.aws
#     Path: /
#     Body: {"parameter":"mongodb"} (base64 encoded as eyJwYXJhbWV0ZXIiOiJtb25nb2RiIn0=)
#    Token: Bearer secret
# Response: {"uri":"mongodb://127.0.0.1:27001/?retryWrites=false"}
*/
func queryHTTP(unused awsConfigSolver, httpOptions string) (string, error) {
	const me = "queryHTTP"

	options := strings.SplitN(httpOptions, ",", 6)
	if len(options) < 6 {
		return "", fmt.Errorf("%s: bad http options, expecting 6 fields - got: '%s'",
			me, httpOptions)
	}

	method := options[0]
	proto := options[1]
	host := options[2]
	path := options[3]
	body := options[4]
	token := options[5]

	u, errJoin := url.JoinPath(proto+"://"+host, path)
	if errJoin != nil {
		return "", errJoin
	}

	bodyPlain, errBody := base64.StdEncoding.DecodeString(body)
	if errBody != nil {
		return "", errBody
	}

	req, errReq := http.NewRequest(method, u, bytes.NewBuffer(bodyPlain))
	if errReq != nil {
		return "", errReq
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", token)

	client := http.DefaultClient
	resp, errDo := client.Do(req)
	if errDo != nil {
		return "", errDo
	}

	defer resp.Body.Close()

	respBody, errRead := io.ReadAll(resp.Body)
	if errRead != nil {
		return "", errRead
	}

	str := string(respBody)

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("%s: URL=%s bad status=%d: %v",
			me, u, resp.StatusCode, str)
	}

	return str, nil
}
