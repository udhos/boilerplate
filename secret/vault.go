package secret

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/hashicorp/vault-client-go"
)

/*
export DB_URI=vault::http,localhost,8200,secret/foo:field
*/
func queryVault( /*unused*/ _ awsConfigSolver, vaultOptions string) (string, error) {
	const me = "queryVault"

	//
	// parse fields
	//

	const fields = 4

	options := strings.SplitN(vaultOptions, ",", fields)
	if len(options) < fields {
		return "", fmt.Errorf("%s: bad vault options, expecting %d fields - got: '%s'",
			me, fields, vaultOptions)
	}

	proto := options[0]
	host := options[1]
	port := options[2]
	path := options[3]

	if port != "" {
		host += ":" + port
	}

	//
	// build vault url
	//

	u, errJoin := url.JoinPath(proto+"://"+host, path)
	if errJoin != nil {
		return "", errJoin
	}

	log.Printf("%s: url: %s\n", me, u)

	//
	// resolve path
	//

	mountPath, secretPath, _ := strings.Cut(path, "/")
	mountPath = strings.TrimSpace(mountPath)
	if mountPath == "" {
		return "", fmt.Errorf("empty mount path is invalid: %s", path)
	}
	secretPath = strings.TrimSpace(secretPath)
	if secretPath == "" {
		return "", fmt.Errorf("empty secret path is invalid: %s", path)
	}

	//
	// login
	//

	client, err := vault.New(
		vault.WithAddress(u),
		vault.WithRequestTimeout(30*time.Second),
	)
	if err != nil {
		return "", err
	}

	if err := client.SetToken("dev-only-token"); err != nil {
		return "", err
	}

	//
	// query vault api
	//

	s, err := client.Secrets.KvV2Read(context.Background(), secretPath, vault.WithMountPath(mountPath))
	if err != nil {
		return "", err
	}

	log.Println("secret retrieved:", s.Data.Data)

	//
	// encode answer as json
	//

	data, err := json.Marshal(s.Data.Data)
	if err != nil {
		return "", err
	}

	return string(data), nil
}
