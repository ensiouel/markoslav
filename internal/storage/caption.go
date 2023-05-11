package storage

import (
	"context"
	"errors"
	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"markoslav/internal/model"
	"markoslav/pkg/apperror"
	"markoslav/pkg/filter"
	"markoslav/pkg/postgres"
)

type CaptionStorage interface {
	Create(ctx context.Context, caption model.Caption) error

	GetByID(ctx context.Context, captionID uuid.UUID) (model.Caption, error)
	GetRandom(ctx context.Context) (model.Caption, error)

	ExistsByText(ctx context.Context, text string) (bool, error)

	Select(ctx context.Context, count int, offset int, options filter.Options) ([]model.Caption, error)

	Update(ctx context.Context, caption model.Caption) error

	Delete(ctx context.Context, captionID uuid.UUID) error
}

type captionStorage struct {
	client postgres.Client
}

func NewCaptionStorage(client postgres.Client) CaptionStorage {
	return &captionStorage{client: client}
}

func (storage *captionStorage) Create(ctx context.Context, caption model.Caption) error {
	builder := squirrel.Insert("caption").
		Columns("id", "text", "author_id", "approved", "created_at").
		Values(caption.ID, caption.Text, caption.AuthorID, caption.Approved, caption.CreatedAt).
		PlaceholderFormat(squirrel.Dollar)

	q, args, err := builder.ToSql()
	if err != nil {
		return apperror.Internal.WithError(err)
	}

	_, err = storage.client.Exec(ctx, q, args...)
	if err != nil {
		return apperror.Internal.WithError(err)
	}

	return nil
}

func (storage *captionStorage) GetByID(ctx context.Context, captionID uuid.UUID) (model.Caption, error) {
	builder := squirrel.Select("id", "text", "author_id", "approved", "created_at").
		From("caption").
		Where(squirrel.Eq{"id": captionID}).
		PlaceholderFormat(squirrel.Dollar)

	q, args, err := builder.ToSql()
	if err != nil {
		return model.Caption{}, apperror.Internal.WithError(err)
	}

	var caption model.Caption
	err = storage.client.Get(ctx, &caption, q, args...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Caption{}, apperror.NotFound.WithError(err)
		}

		return model.Caption{}, apperror.Internal.WithError(err)
	}

	return caption, nil
}

func (storage *captionStorage) GetRandom(ctx context.Context) (model.Caption, error) {
	builder := squirrel.Select("id", "text", "author_id", "approved", "created_at").
		From("caption").
		OrderBy("random()").
		Limit(1).
		PlaceholderFormat(squirrel.Dollar)

	q, args, err := builder.ToSql()
	if err != nil {
		return model.Caption{}, apperror.Internal.WithError(err)
	}

	var caption model.Caption
	err = storage.client.Get(ctx, &caption, q, args...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Caption{}, apperror.NotFound.WithError(err)
		}

		return model.Caption{}, apperror.Internal.WithError(err)
	}

	return caption, nil
}

func (storage *captionStorage) ExistsByText(ctx context.Context, text string) (bool, error) {
	q := `SELECT EXISTS (SELECT 1 FROM caption WHERE text = $1)`

	var exists bool
	err := storage.client.Get(ctx, &exists, q, text)
	if err != nil {
		return false, apperror.Internal.WithError(err)
	}

	return exists, nil
}

func (storage *captionStorage) Select(ctx context.Context, count int, offset int, options filter.Options) ([]model.Caption, error) {
	builder := squirrel.Select("id", "text", "author_id", "approved", "created_at").
		From("caption").
		Limit(uint64(count)).
		Offset(uint64(offset)).
		PlaceholderFormat(squirrel.Dollar)

	for _, field := range options.Fields() {
		switch field.Operator {
		case filter.OperatorEq:
			builder = builder.Where(squirrel.Eq{field.Name: field.Value})
		case filter.OperatorNotEq:
			builder = builder.Where(squirrel.NotEq{field.Name: field.Value})
		}
	}

	q, args, err := builder.ToSql()
	if err != nil {
		return nil, apperror.Internal.WithError(err)
	}

	var captions []model.Caption
	err = storage.client.Select(ctx, &captions, q, args...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperror.NotFound.WithError(err)
		}

		return nil, apperror.Internal.WithError(err)
	}

	return captions, nil
}

func (storage *captionStorage) Update(ctx context.Context, caption model.Caption) error {
	builder := squirrel.Update("caption").
		Set("text", caption.Text).
		Set("author_id", caption.AuthorID).
		Set("approved", caption.Approved).
		Set("created_at", caption.CreatedAt).
		Where(squirrel.Eq{"id": caption.ID}).
		PlaceholderFormat(squirrel.Dollar)

	q, args, err := builder.ToSql()
	if err != nil {
		return apperror.Internal.WithError(err)
	}

	_, err = storage.client.Exec(ctx, q, args...)
	if err != nil {
		return apperror.Internal.WithError(err)
	}

	return nil
}

func (storage *captionStorage) Delete(ctx context.Context, captionID uuid.UUID) error {
	builder := squirrel.Delete("caption").
		Where(squirrel.Eq{"id": captionID}).
		PlaceholderFormat(squirrel.Dollar)

	q, args, err := builder.ToSql()
	if err != nil {
		return apperror.Internal.WithError(err)
	}

	_, err = storage.client.Exec(ctx, q, args...)
	if err != nil {
		return apperror.Internal.WithError(err)
	}

	return nil
}
