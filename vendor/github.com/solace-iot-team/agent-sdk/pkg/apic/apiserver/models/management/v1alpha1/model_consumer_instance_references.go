/*
 * API Server specification.
 *
 * No description provided (generated by Openapi Generator https://github.com/openapitools/openapi-generator)
 *
 * API version: SNAPSHOT
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package v1alpha1

// ConsumerInstanceReferences struct for ConsumerInstanceReferences
type ConsumerInstanceReferences struct {
	// Reference to Amplify Central APIService
	ApiService string `json:"apiService,omitempty"`
	// Reference to Amplify Central APIServiceRevision
	ApiServiceRevision string `json:"apiServiceRevision,omitempty"`
}
