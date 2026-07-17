package repository

import (
	"context"

	"grocerics-backend/internal/domain"

	"gorm.io/gorm"
)

type QCRawResponseRepository struct{ db *gorm.DB }

func NewQCRawResponseRepository(db *gorm.DB) *QCRawResponseRepository {
	return &QCRawResponseRepository{db: db}
}
func (r *QCRawResponseRepository) Create(row *domain.QCRawResponse) error {
	return gorm.G[domain.QCRawResponse](r.db).Create(context.Background(), row)
}
