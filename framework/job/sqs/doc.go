// Package sqs provides an Amazon SQS-backed job driver.
//
// [New] creates its AWS SDK client from [SQSDriverConfig.AWSConfig]. [NewFrom]
// accepts a caller-owned SDK client.
package sqs
