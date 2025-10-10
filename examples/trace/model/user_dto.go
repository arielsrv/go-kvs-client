package model

type UserDTO struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	FullName  string `json:"full_name"`
	ID        int    `json:"id"`
}
