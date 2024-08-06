package model

import "fmt"

type UserDTO struct {
	ID        int    `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	FullName  string `json:"full_name"`
}

func NewUserDTO(firstName string, lastName string) *UserDTO {
	return &UserDTO{
		FirstName: firstName,
		LastName:  lastName,
		FullName:  fmt.Sprintf("%s %s", firstName, lastName),
	}
}
