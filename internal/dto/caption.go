package dto

import "github.com/google/uuid"

type CreateCaption struct {
	Text     string
	AuthorID int64
}

type UpdateCaption struct {
	ID       uuid.UUID
	Approved bool
}
