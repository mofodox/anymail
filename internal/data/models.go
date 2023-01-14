package data

import (
	"github.com/mofodox/anymail/internal/database"
)

type Models struct {
	MTPPCards MTPPCardModel
}

func NewModels(db *database.DB) Models {
	return Models{
		MTPPCards: MTPPCardModel{DB: db},
	}
}
