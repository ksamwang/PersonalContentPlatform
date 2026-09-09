package domain

import "github.com/google/uuid"

type Collection struct {
	ID          uuid.UUID `json:"id"`
	WorkspaceID uuid.UUID `json:"workspace_id"`
	Title       string    `json:"title"`
	Slug        string    `json:"slug"`
	Visibility  string    `json:"visibility"`
	Sections    []Section `json:"sections"`
}

type Section struct {
	ID      uuid.UUID `json:"id"`
	Title   string    `json:"title"`
	SortKey float64   `json:"sort_key"`
	Items   []Item    `json:"items"`
}

type Item struct {
	ID         uuid.UUID `json:"id"`
	ObjectID   uuid.UUID `json:"object_id"`
	Title      string    `json:"title"`
	SortKey    float64   `json:"sort_key"`
	Annotation string    `json:"annotation"`
}
