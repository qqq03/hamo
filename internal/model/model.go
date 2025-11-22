package model

// Theme: 테마 (대묶음)
type Theme struct {
	ThemeID   string `json:"theme_id"`
	ThemeName string `json:"theme_name"`
	ThemeDesc string `json:"theme_desc"`
}

// Item: 전시물 상세 정보 (스크립트 포함)
type Item struct {
	ThemeID       string  `json:"theme_id"`
	ItemSeq       int     `json:"item_seq"`
	ItemName      string  `json:"item_name"`
	ItemDesc      string  `json:"item_desc"`
	ScriptChild   string  `json:"script_child"`
	ScriptGeneral string  `json:"script_general"`
	Latitude      float64 `json:"latitude"`
	Longitude     float64 `json:"longitude"`
}

// Quiz: 퀴즈 정보
type Quiz struct {
	ThemeID  string `json:"theme_id"`
	QuizNo   int    `json:"quiz_no"`
	Question string `json:"question"`
	Answer   string `json:"answer"`
	Options  string `json:"options"`
	QuizDesc string `json:"quiz_desc"`
}

// RecipientRequest: 수령자 정보 (입력용 구조체)
type RecipientRequest struct {
	ThemeID string `json:"theme_id"`
	Email   string `json:"email"`
}
