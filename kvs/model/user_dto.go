package model

import "fmt"

// UserDTO is an example value type stored through the KVS client.
// It shows the shape a domain model needs: exported fields with JSON tags,
// since values are marshalled to JSON before they reach the backend.
type UserDTO struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	FullName  string `json:"full_name"`
	ID        int    `json:"id"`
}

// NewUserDTO creates a UserDTO from a first and last name,
// deriving FullName from both. ID is left unset for the caller to assign.
func NewUserDTO(firstName string, lastName string) *UserDTO {
	return &UserDTO{
		FirstName: firstName,
		LastName:  lastName,
		FullName:  fmt.Sprintf("%s %s", firstName, lastName),
	}
}
