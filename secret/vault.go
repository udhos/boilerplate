package secret

import (
	"context"
	"fmt"
	"strings"

	vault "github.com/hashicorp/vault/api"
	auth "github.com/hashicorp/vault/api/auth/aws"
	"github.com/udhos/boilerplate/boilerplate"
)

/*
export DB_URI=vault::token,dev-only-token,http,localhost,8200,secret/foo/key:field
*/
func queryVault(debug bool, printf boilerplate.FuncPrintf, _ /*unused*/ AwsConfigSolver, vaultOptions string) (string, error) {
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

	// drop spaces
	for i, s := range options {
		options[i] = strings.TrimSpace(s)
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

	u := proto + "://" + host

	if debug {
		printf("DEBUG %s: vault server URL: %s", me, u)
	}

	//
	// resolve path: secret/<secretPath>/<key>
	//

	mountPath, secretPath, key, errPath := parseSecretPath(path)
	if errPath != nil {
		return "", errPath
	}

	if debug {
		printf("DEBUG %s: raw_path=%s mount_path=%s secret_path=%s key=%s",
			me, path, mountPath, secretPath, key)
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
		client, err = vaultClientFromAwsRole(u, authOption)
		if err != nil {
			return "", err
		}
	default:
		return "", fmt.Errorf("unexpected auth type (token|aws-role): '%s': %s", authType, vaultOptions)
	}

	client.SetAddress(u)

	//
	// query vault api
	//

	s, err := client.KVv2(mountPath).Get(context.Background(), secretPath)
	if err != nil {
		return "", err
	}

	value := s.Data[key]

	if debug {
		printf("DEBUG %s: raw_path=%s mount_path=%s secret_path=%s key=%s raw_value=%v keyed_value=%v",
			me, path, mountPath, secretPath, key, s.Data, value)
	}

	str, isStr := value.(string)

	if !isStr {
		return "", fmt.Errorf("%s: not a string: %T: %v", me, value, value)
	}

	return str, nil
}

func vaultClientFromToken(u, token string) (*vault.Client, error) {
	config := vault.DefaultConfig()
	config.Address = u
	client, err := vault.NewClient(config)
	if err != nil {
		return nil, err
	}
	client.SetToken(token)
	return client, nil
}

func vaultClientFromAwsRole(u, role string) (*vault.Client, error) {

	config := vault.DefaultConfig() // modify for more granular configuration
	config.Address = u

	client, err := vault.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize Vault client: %w", err)
	}

	awsAuth, err := auth.NewAWSAuth(
		// if not provided, Vault will fall back on looking for
		// a role with the IAM role name if you're using the iam auth type,
		// or the EC2 instance's AMI id if using the ec2 auth type
		auth.WithRole(role),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize AWS auth method: %w", err)
	}

	authInfo, err := client.Auth().Login(context.Background(), awsAuth)
	if err != nil {
		return nil, fmt.Errorf("unable to login to AWS auth method: %w", err)
	}
	if authInfo == nil {
		return nil, fmt.Errorf("no auth info was returned after login")
	}

	return client, nil
}

func parseSecretPath(path string) (string, string, string, error) {
	const minSlash = 2

	path = strings.TrimSpace(path)
	path = strings.TrimPrefix(path, "/")

	if slashes := strings.Count(path, "/"); slashes < minSlash {
		return "", "", "", fmt.Errorf("bad vault path, expecting %d slashes - got %d: '%s'",
			minSlash, slashes, path)
	}

	secretIndex := strings.IndexByte(path, '/')
	if secretIndex < 0 {
		return "", "", "", fmt.Errorf("missing secret from path: %s", path)
	}

	mountPath := path[:secretIndex]

	keyIndex := strings.LastIndexByte(path, '/')
	if keyIndex < 0 {
		return "", "", "", fmt.Errorf("missing key from secret path: %s", path)
	}
	key := path[keyIndex+1:]

	secretPath := path[secretIndex+1 : keyIndex]

	mountPath = strings.TrimSpace(mountPath)
	if mountPath == "" {
		return "", "", "", fmt.Errorf("empty mount path is invalid: %s", path)
	}
	secretPath = strings.TrimSpace(secretPath)
	if secretPath == "" {
		return "", "", "", fmt.Errorf("empty secret path is invalid: %s", path)
	}
	key = strings.TrimSpace(key)
	if key == "" {
		return "", "", "", fmt.Errorf("empty key is invalid: %s", path)
	}

	return mountPath, secretPath, key, nil
}
