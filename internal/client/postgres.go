package client

import (
	"context"
	"database/sql"
)

type clientRepository struct {
	db *sql.DB
}

func NewPostgresRepo(db *sql.DB) Repository {
	return &clientRepository{db: db}
}

func (r *clientRepository) Create(ctx context.Context, client Client) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO clients (id, capacity, refill_rate)
		VALUES ($1, $2, $3)
		ON CONFLICT (id) DO NOTHING`, client.ID, client.Capacity, client.RefillRate)

	if err != nil {
		return err
	}

	return nil
}

func (r *clientRepository) Get(ctx context.Context, id string) (*Client, error) {
	var client Client
	err := r.db.QueryRowContext(ctx, `
		SELECT id, capacity, refill_rate FROM clients WHERE id = $1`, id).
		Scan(&client.ID, &client.Capacity, &client.RefillRate)
	if err != nil {
		return nil, err
	}

	return &client, nil
}

func (r *clientRepository) Update(ctx context.Context, client Client) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE clients SET capacity = $2, refill_rate = $3 WHERE id = $1`,
		client.ID, client.Capacity, client.RefillRate)
	if err != nil {
		return err
	}

	return nil
}

func (r *clientRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM clients WHERE id = $1`, id)
	if err != nil {
		return err
	}

	return nil
}

func (r *clientRepository) Close() error {
	if r.db != nil {
		return r.db.Close()
	}

	return nil
}
