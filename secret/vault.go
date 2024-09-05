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

	u, errJoin := url.JoinPath(proto+"://"+host, path)
	if errJoin != nil {
		return "", errJoin
	}

	log.Printf("%s: url: %s", me, u)

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

	s, err := client.Secrets.KvV1Read(context.Background(), path)
	if err != nil {
		return "", err
	}

	log.Println("secret retrieved:", s.Data)

	data, err := json.Marshal(s.Data)
	if err != nil {
		return "", err
	}

	return string(data), nil
}
