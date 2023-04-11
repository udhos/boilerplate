// Package envconfig loads configuration from env vars.
package envconfig

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// aws-dynamodb:region:table_name,key_name,key_value,value_attr[:field_name]
func queryDynamoDb(getAwsConfig awsConfigSolver, dynamoOptions string) (string, error) {
	const me = "queryDynamoDb"

	options := strings.SplitN(dynamoOptions, ",", 4)
	if len(options) < 4 {
		return "", fmt.Errorf("%s: bad dynamodb options, expecting 4 fields - got: '%s'",
			me, dynamoOptions)
	}

	table := options[0]
	keyName := options[1]
	keyValue := options[2]
	attrField := options[3]

	awsConfig, errAwsConfig := getAwsConfig.get()
	if errAwsConfig != nil {
		return "", errAwsConfig
	}

	dc := dynamodb.NewFromConfig(awsConfig)

	av, errMarshal := attributevalue.Marshal(keyValue)
	if errMarshal != nil {
		return "", errMarshal
	}

	key := map[string]types.AttributeValue{keyName: av}

	response, errGet := dc.GetItem(context.TODO(), &dynamodb.GetItemInput{
		Key: key, TableName: aws.String(table),
	})

	if errGet != nil {
		return "", errGet
	}

	if len(response.Item) == 0 {
		return "", fmt.Errorf("%s: item not found: '%s'",
			me, dynamoOptions)
	}

	body := map[string]string{}

	errUnmarshal := attributevalue.UnmarshalMap(response.Item, &body)
	if errUnmarshal != nil {
		return "", errUnmarshal
	}

	value, found := body[attrField]
	if !found {
		return "", fmt.Errorf("%s: item attribute '%s' not found: '%s'",
			me, attrField, dynamoOptions)
	}

	return value, nil
}
