terraform {
	required_version = ">= 1.0.0"

	required_providers {
		aws = {
			source  = "hashicorp/aws"
			version = ">= 4.0.0"
		}
	}
}

provider "aws" {
	access_key                  = "test"
	secret_key                  = "test"
	region                      = "eu-west-1"
	s3_use_path_style           = false
	skip_credentials_validation = true
	skip_metadata_api_check     = true
	skip_requesting_account_id  = false

	endpoints {
		config         = "http://localhost:4566"
		apigateway     = "http://localhost:4566"
		apigatewayv2   = "http://localhost:4566"
		cloudformation = "http://localhost:4566"
		cloudwatch     = "http://localhost:4566"
		dynamodb       = "http://localhost:4566"
		ec2            = "http://localhost:4566"
		es             = "http://localhost:4566"
		elasticache    = "http://localhost:4566"
		firehose       = "http://localhost:4566"
		iam            = "http://localhost:4566"
		kinesis        = "http://localhost:4566"
		lambda         = "http://localhost:4566"
		rds            = "http://localhost:4566"
		redshift       = "http://localhost:4566"
		route53        = "http://localhost:4566"
		s3             = "http://s3.localhost.localstack.cloud:4566"
		secretsmanager = "http://localhost:4566"
		ses            = "http://localhost:4566"
		sns            = "http://localhost:4566"
		sqs            = "http://localhost:4566"
		ssm            = "http://localhost:4566"
		stepfunctions  = "http://localhost:4566"
		sts            = "http://localhost:4566"
	}
}

// Create a DynamoDB table for KVS
// Please don't add anything like streams, triggers, secondary indexes, or things we have to be magicians to figure out.
// It's a key-value store, and even less things like streams where business logic is tied to a data repository.
resource "aws_dynamodb_table" "kvs" {
	name         = "__kvs-users-store"
	billing_mode = "PAY_PER_REQUEST"
	hash_key     = "key"
	point_in_time_recovery {
		enabled = false
	}

	attribute {
		name = "key"
		type = "S"
	}

	ttl {
		attribute_name = "ttl"
		enabled        = true
	}

	tags = {
		Name        = "kvs"
		Environment = "localstack"
	}
}
