// Package request contains HTTP request helper functions for Cosmos contracts.
//
// The package focuses on safe extraction and decoding of request data:
// path/query parameters, headers/cookies, body parsing, lifecycle hooks,
// sessions, and correlation IDs.
//
// # Behavior notes
//
// Body helpers return typed errors for malformed payloads and enforce optional
// body size limits. Session helpers expose both panic-on-missing and
// non-panicking retrieval paths to match different middleware assumptions.
//
// Example
//
//	id, err := request.ParamInt(r, "id")
//	if err != nil {
//		return err
//	}
//
//	body, err := request.StrictLimitedJSON[CreateUserInput](r, 1<<20)
//	if err != nil {
//		return err
//	}
//
//	_ = id
//	_ = body
package request
