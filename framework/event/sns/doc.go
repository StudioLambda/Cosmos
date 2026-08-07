// Package sns provides an Amazon SNS-backed event publisher.
//
// Event names are mapped to configured SNS topic ARNs. The publisher does not
// create subscriptions or provision SNS resources.
//
// [New] creates its AWS SDK client from [SNSPublisherConfig.AWSConfig].
// [NewFrom] accepts a caller-owned SDK client.
package sns
