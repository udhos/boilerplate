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
	//auth "github.com/hashicorp/vault-client-go/api/auth/aws"
	auth "github.com/hashicorp/vault/api/auth/aws"
)

/*
export DB_URI=vault::token,dev-only-token,http,localhost,8200,secret/foo:field
*/
func queryVault( /*unused*/ _ awsConfigSolver, vaultOptions string) (string, error) {
	const me = "queryVault"

	//
	// parse fields
	//

	const fields = 6

	options := strings.SplitN(vaultOptions, ",", fields)
	if len(options) < fields {
		return "", fmt.Errorf("%s: bad vault options, expecting %d fields - got: '%s'",
			me, fields, vaultOptions)
	}

	authType := options[0]
	authOption := options[1]
	proto := options[2]
	host := options[3]
	port := options[4]
	path := options[5]

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

	var client *vault.Client

	switch authType {
	case "token":
		var err error
		client, err = vaultClientFromToken(u, authOption)
		if err != nil {
			return "", err
		}
	case "aws-role", "":
		var err error
		client, err = vaultClientFromAwsRole(u)
		if err != nil {
			return "", err
		}
	default:
		return "", fmt.Errorf("unexpected auth type (token|aws-role): '%s': %s", authType, vaultOptions)
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

func vaultClientFromToken(u, token string) (*vault.Client, error) {
	client, err := vault.New(
		vault.WithAddress(u),
		vault.WithRequestTimeout(30*time.Second),
	)
	if err != nil {
		return nil, err
	}
	if err := client.SetToken(token); err != nil {
		return nil, err
	}
	return client, nil
}

func vaultClientFromAwsRole(u string) (*vault.Client, error) {

	client, err := vault.New(
		vault.WithAddress(u),
		vault.WithRequestTimeout(30*time.Second),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize Vault client: %w", err)
	}

	_, err = auth.NewAWSAuth(
		// if not provided, Vault will fall back on looking for a role with the IAM role name if you're using the iam auth type, or the EC2 instance's AMI id if using the ec2 auth type
		// dev-role-iam:
		//   $ vault write auth/aws/role/dev-role-iam auth_type=iam bound_iam_principal_arn=arn:aws:iam::123456789012:role/MyRole policies=prod,dev max_ttl=500h
		// see https://developer.hashicorp.com/vault/docs/auth/aws
		auth.WithRole("dev-role-iam"),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize AWS auth method: %w", err)
	}

	return client, nil
}
