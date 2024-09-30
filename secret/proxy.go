package secret

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/udhos/boilerplate/boilerplate"
)

/*
proxy||proto,host,port,secret_name[|field_name]

export DB_URI=proxy||http,localhost,8080,vault::token,dev-only-token,http,localhost,8200,secret/myapp1/mongodb:uri
*/
func queryProxy(debug bool, printf boilerplate.FuncPrintf,
	_ /*unused*/ awsConfigSolver, proxyOptions string) (string, error) {
	const me = "queryProxy"

	const fields = 4

	options := strings.SplitN(proxyOptions, ",", fields)
	if len(options) < fields {
		return "", fmt.Errorf("%s: bad proxy options, expecting %d fields - got: '%s'",
			me, fields, proxyOptions)
	}

	// remove spaces
	for i, s := range options {
		options[i] = strings.TrimSpace(s)
	}

	proto := options[0]
	host := options[1]
	port := options[2]
	secretName := options[3]

	if port != "" {
		host += ":" + port
	}

	u, errJoin := url.JoinPath(proto+"://"+host, "/secret")
	if errJoin != nil {
		return "", errJoin
	}

	requestBody := proxyPayload{
		SecretName: secretName,
	}

	body, errBody := json.Marshal(requestBody)
	if errBody != nil {
		return "", errBody
	}

	req, errReq := http.NewRequest("POST", u, bytes.NewBuffer(body))
	if errReq != nil {
		return "", errReq
	}

	req.Header.Set("Content-Type", "application/json")

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

	var responseBody proxyPayload

	errJSON := json.Unmarshal(respBody, &responseBody)

	if debug {
		printf("DEBUG %s: secret_name=%s secret_value=%s body=%s error=%v",
			me, secretName, responseBody.SecretValue, str, errJSON)
	}

	return responseBody.SecretValue, errJSON
}

type proxyPayload struct {
	SecretName  string `json:"secret_name,omitempty"`
	SecretValue string `json:"secret_value,omitempty"`
}
