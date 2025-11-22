package repository

import (
	"context"
	"database/sql"

	"hamo/internal/model"
)

// MySQLRepository MySQL 구현체
type MySQLRepository struct {
	DB *sql.DB
}

// NewMySQLRepository MySQL Repository 생성
func NewMySQLRepository(db *sql.DB) *MySQLRepository {
	return &MySQLRepository{DB: db}
}

// GetAllThemes 테마 전체 조회
func (r *MySQLRepository) GetAllThemes(ctx context.Context) ([]model.Theme, error) {
	query := `SELECT THEME_ID, THEME_NAME, COALESCE(THEME_DESC, '') FROM Theme`

	rows, err := r.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var themes []model.Theme
	for rows.Next() {
		var t model.Theme
		if err := rows.Scan(&t.ThemeID, &t.ThemeName, &t.ThemeDesc); err != nil {
			return nil, err
		}
		themes = append(themes, t)
	}
	return themes, nil
}

// GetItemsByTheme 테마별 아이템 조회
func (r *MySQLRepository) GetItemsByTheme(ctx context.Context, themeID string) ([]model.Item, error) {
	query := `SELECT 
				THEME_ID, ITEM_SEQ, ITEM_NAME, 
				COALESCE(ITEM_DESC, ''), 
				COALESCE(SCRIPT_CHILD, ''), 
				COALESCE(SCRIPT_GENERAL, ''), 
				COALESCE(LATITUDE, 0.0), 
				COALESCE(LONGITUDE, 0.0) 
			  FROM Item 
			  WHERE THEME_ID = ? 
			  ORDER BY ITEM_SEQ ASC`

	rows, err := r.DB.QueryContext(ctx, query, themeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []model.Item
	for rows.Next() {
		var i model.Item
		if err := rows.Scan(&i.ThemeID, &i.ItemSeq, &i.ItemName, &i.ItemDesc,
			&i.ScriptChild, &i.ScriptGeneral, &i.Latitude, &i.Longitude); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, nil
}

// GetQuizzesByTheme 테마별 퀴즈 조회
func (r *MySQLRepository) GetQuizzesByTheme(ctx context.Context, themeID string) ([]model.Quiz, error) {
	query := `SELECT 
				THEME_ID, QUIZ_NO, QUESTION, ANSWER, 
				COALESCE(OPTIONS, ''), 
				COALESCE(QUIZ_DESC, '') 
			  FROM Quiz 
			  WHERE THEME_ID = ? 
			  ORDER BY QUIZ_NO ASC`

	rows, err := r.DB.QueryContext(ctx, query, themeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var quizzes []model.Quiz
	for rows.Next() {
		var q model.Quiz
		if err := rows.Scan(&q.ThemeID, &q.QuizNo, &q.Question, &q.Answer, &q.Options, &q.QuizDesc); err != nil {
			return nil, err
		}
		quizzes = append(quizzes, q)
	}
	return quizzes, nil
}

// AddRecipient 수령자 등록
func (r *MySQLRepository) AddRecipient(ctx context.Context, req model.RecipientRequest) error {
	query := `INSERT INTO Recipient (THEME_ID, EMAIL, RECV_DATE, RECV_TIME) VALUES (?, ?, CURDATE(), CURTIME())`
	_, err := r.DB.ExecContext(ctx, query, req.ThemeID, req.Email)
	return err
}
