package model

import (
	"github.com/google/uuid"
	"time"
)

type Caption struct {
	ID        uuid.UUID `db:"id"`
	Text      string    `db:"text"`
	AuthorID  int64     `db:"author_id"`
	Approved  bool      `db:"approved"`
	CreatedAt time.Time `db:"created_at"`
}
