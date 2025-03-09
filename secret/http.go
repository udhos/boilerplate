package secret

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/udhos/boilerplate/boilerplate"
)

/*
export DB_URI=#http::GET,https,tttt.lambda-url.us-east-1.on.aws,443,/,text/plain,eyJwYXJhbWV0ZXIiOiJtb25nb2RiIn0=,Bearer secret:uri
#       Method: GET
#     Protocol: https
#         Host: tttt.lambda-url.us-east-1.on.aws
#         Port: 443
#         Path: /
# Content-Type: text/plain
#         Body: {"parameter":"mongodb"} (base64 encoded as eyJwYXJhbWV0ZXIiOiJtb25nb2RiIn0=)
#        Token: Bearer secret
#     Response: {"uri":"mongodb://127.0.0.1:27001/?retryWrites=false"}
*/
func queryHTTP(_ /*debug*/ bool, _ /*printf*/ boilerplate.FuncPrintf, _ /*unused*/ AwsConfigSolver, httpOptions string) (string, error) {
	const me = "queryHTTP"

	const minFields = 8
	options := strings.SplitN(httpOptions, ",", minFields)
	if len(options) < minFields {
		return "", fmt.Errorf("%s: bad http options, expecting %d fields - got: '%s'",
			me, minFields, httpOptions)
	}

	for i, o := range options {
		options[i] = strings.TrimSpace(o)
	}

	method := options[0]
	proto := options[1]
	host := options[2]
	port := options[3]
	path := options[4]
	contentType := options[5]
	body := options[6]
	token := options[7]

	if port != "" {
		host += ":" + port
	}

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

	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	if token != "" {
		req.Header.Set("Authorization", token)
	}

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
