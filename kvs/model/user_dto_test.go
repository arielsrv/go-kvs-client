package model_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/arielsrv/go-kvs-client/kvs/model"
)

func TestNewUserDTO(t *testing.T) {
	tests := []struct {
		name      string
		firstName string
		lastName  string
		want      *model.UserDTO
	}{
		{
			name:      "Create user with first and last name",
			firstName: "John",
			lastName:  "Doe",
			want: &model.UserDTO{
				FirstName: "John",
				LastName:  "Doe",
				FullName:  "John Doe",
			},
		},
		{
			name:      "Create user with empty first name",
			firstName: "",
			lastName:  "Doe",
			want: &model.UserDTO{
				FirstName: "",
				LastName:  "Doe",
				FullName:  " Doe",
			},
		},
		{
			name:      "Create user with empty last name",
			firstName: "John",
			lastName:  "",
			want: &model.UserDTO{
				FirstName: "John",
				LastName:  "",
				FullName:  "John ",
			},
		},
		{
			name:      "Create user with empty names",
			firstName: "",
			lastName:  "",
			want: &model.UserDTO{
				FirstName: "",
				LastName:  "",
				FullName:  " ",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := model.NewUserDTO(tt.firstName, tt.lastName)
			assert.Equal(t, tt.want.FirstName, got.FirstName)
			assert.Equal(t, tt.want.LastName, got.LastName)
			assert.Equal(t, tt.want.FullName, got.FullName)
			assert.Equal(t, 0, got.ID) // ID should be 0 by default
		})
	}
}
