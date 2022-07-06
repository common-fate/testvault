package testvault

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/common-fate/ddb"
)

type Membership struct {
	Vault  string `json:"vault" dynamodbav:"vault"`
	User   string `json:"user" dynamodbav:"user"`
	Active bool   `json:"active"`
}

func (m *Membership) DDBKeys() (ddb.Keys, error) {
	keys := ddb.Keys{
		PK: "VAULT#" + m.Vault,
		SK: m.User,
	}
	return keys, nil
}

// GetMembership is the access pattern to
// fetch a vault membership from DynamoDB.
//
// It returns ddb.ErrNoItems if the vault
// membership doesn't exist.
type GetMembership struct {
	Vault  string
	User   string
	Result *Membership
}

func (g *GetMembership) BuildQuery() (*dynamodb.QueryInput, error) {
	qi := &dynamodb.QueryInput{
		Limit:                  aws.Int32(1),
		KeyConditionExpression: aws.String("PK = :pk1 and SK = :sk1"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk1": &types.AttributeValueMemberS{Value: "VAULT#" + g.Vault},
			":sk1": &types.AttributeValueMemberS{Value: g.User},
		},
	}
	return qi, nil
}

func (g *GetMembership) UnmarshalQueryOutput(out *dynamodb.QueryOutput) error {
	if len(out.Items) != 1 {
		return ddb.ErrNoItems
	}

	return attributevalue.UnmarshalMap(out.Items[0], &g.Result)
}
