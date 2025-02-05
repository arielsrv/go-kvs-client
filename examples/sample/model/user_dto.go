package model

import "fmt"

type UserDTO struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	FullName  string `json:"full_name"`
	ID        int    `json:"id"`
}

func NewUserDTO(firstName string, lastName string) *UserDTO {
	return &UserDTO{
		FirstName: firstName,
		LastName:  lastName,
		FullName:  fmt.Sprintf("%s %s", firstName, lastName),
	}
}
