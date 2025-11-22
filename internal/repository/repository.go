package repository

import (
	"context"

	"hamo/internal/model"
)

// Repository 인터페이스 정의 (향후 다른 DB로 교체 가능)
type Repository interface {
	// Theme 관련
	GetAllThemes(ctx context.Context) ([]model.Theme, error)

	// Item 관련
	GetItemsByTheme(ctx context.Context, themeID string) ([]model.Item, error)

	// Quiz 관련
	GetQuizzesByTheme(ctx context.Context, themeID string) ([]model.Quiz, error)

	// Recipient 관련
	AddRecipient(ctx context.Context, req model.RecipientRequest) error
}
