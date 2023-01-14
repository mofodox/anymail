package data

import (
	"context"
	"time"

	"github.com/mofodox/anymail/internal/database"
)

type MTPPCard struct {
	ID         int       `json:"id"`
	MTPPNumber string    `json:"mtpp_number"`
	MTPPUrl    string    `json:"mtpp_url"`
	CreatedAt  time.Time `json:"created_at"`
	Version    int       `json:"version"`
}

type MTPPCardModel struct {
	DB *database.DB
}

func (m MTPPCardModel) Insert(mtppCard *MTPPCard) error {
	qry := `INSERT INTO mtpp_table (mtpp_number, mtpp_url) VALUES ($1, $2) RETURNING id, created_at, version`

	args := []interface{}{mtppCard.MTPPNumber, mtppCard.MTPPUrl}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return m.DB.QueryRowContext(ctx, qry, args...).Scan(&mtppCard.ID, &mtppCard.CreatedAt, &mtppCard.Version)
}
