package id

import "github.com/google/uuid"

func New() uuid.UUID {
	v, err := uuid.NewV7()
	if err != nil {
		panic(err)
	}
	return v
}
