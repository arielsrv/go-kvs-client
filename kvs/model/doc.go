// Package model provides data transfer objects (DTOs) for use with the kvs package.
//
// This package contains example data models that can be used with the KVS client.
// These models demonstrate how to structure data for storage and retrieval using
// the key-value store client.
//
// The UserDTO is an example model that represents user information with fields for:
//   - FirstName: The user's first name
//   - LastName: The user's last name
//   - FullName: The user's full name (automatically generated)
//   - ID: A unique identifier for the user
//
// Usage:
//
//	// Create a new user DTO
//	user := model.NewUserDTO("John", "Doe")
//
//	// Set the ID
//	user.ID = 123
//
//	// Use with KVS client
//	client := kvs.NewAWSKVSClient[model.UserDTO](options...)
//	err := client.Save("user:123", user)
//
// This package is intended to serve as an example of how to create and use data models
// with the kvs package. In a real application, you would typically define your own
// domain-specific models in a separate package.
package model
