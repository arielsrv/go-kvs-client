package dynamodb

type Item struct {
	Key   string `dynamodbav:"key"`
	Value string `dynamodbav:"value"`
	TTL   int64  `dynamodbav:"ttl"`
}
