package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	// Go MySQL 드라이버
	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"

	// Go Cors 용
	"github.com/rs/cors" // 라이브러리 임포트

	// AWS SDK (Secrets Manager용)
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

//==============================================================
// 1. 모델 정의 (Model) - DB 스키마와 1:1 매핑
//==============================================================

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
	ItemDesc      string  `json:"item_desc"`      // 핵심 메시지
	ScriptChild   string  `json:"script_child"`   // 어린이용 해설
	ScriptGeneral string  `json:"script_general"` // 일반인용 해설
	Latitude      float64 `json:"latitude"`
	Longitude     float64 `json:"longitude"`
}

// Quiz: 퀴즈 정보
type Quiz struct {
	ThemeID  string `json:"theme_id"`
	QuizNo   int    `json:"quiz_no"`
	Question string `json:"question"`
	Answer   string `json:"answer"`
	Options  string `json:"options"` // 보기 (JSON 문자열 또는 콤마 구분)
	QuizDesc string `json:"quiz_desc"`
}

// Recipient: 수령자 정보 (입력용 구조체)
type RecipientRequest struct {
	ThemeID string `json:"theme_id"`
	Email   string `json:"email"`
	// RecvDate string `json:"recv_date"` // YYYY-MM-DD
	// RecvTime string `json:"recv_time"` // HH:MM:SS
}

//==============================================================
// 2. Repository 계층 (DB Access)
//==============================================================

type DBRepository struct {
	DB *sql.DB
}

func NewDBRepository(db *sql.DB) *DBRepository {
	return &DBRepository{DB: db}
}

// 1. 테마 전체 조회
func (r *DBRepository) GetAllThemes(ctx context.Context) ([]Theme, error) {
	query := `SELECT THEME_ID, THEME_NAME, COALESCE(THEME_DESC, ''), CREATED_AT FROM Theme`

	rows, err := r.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var themes []Theme
	for rows.Next() {
		var t Theme
		// CREATED_AT은 DB 설정에 따라 []uint8로 올 수 있어 string 변환이 필요할 수 있으나,
		// Scan 시 string 변수에 넣으면 드라이버가 자동 변환을 시도합니다.
		if err := rows.Scan(&t.ThemeID, &t.ThemeName, &t.ThemeDesc); err != nil {
			return nil, err
		}
		themes = append(themes, t)
	}
	return themes, nil
}

// 2. 테마별 아이템 전체 조회
func (r *DBRepository) GetItemsByTheme(ctx context.Context, themeID string) ([]Item, error) {
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

	var items []Item
	for rows.Next() {
		var i Item
		if err := rows.Scan(&i.ThemeID, &i.ItemSeq, &i.ItemName, &i.ItemDesc,
			&i.ScriptChild, &i.ScriptGeneral, &i.Latitude, &i.Longitude); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, nil
}

// 3. 테마별 퀴즈 조회
func (r *DBRepository) GetQuizzesByTheme(ctx context.Context, themeID string) ([]Quiz, error) {
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

	var quizzes []Quiz
	for rows.Next() {
		var q Quiz
		if err := rows.Scan(&q.ThemeID, &q.QuizNo, &q.Question, &q.Answer, &q.Options, &q.QuizDesc); err != nil {
			return nil, err
		}
		quizzes = append(quizzes, q)
	}
	return quizzes, nil
}

// 4. 수령자 등록 (INSERT)
func (r *DBRepository) AddRecipient(ctx context.Context, req RecipientRequest) error {
	query := `INSERT INTO Recipient (THEME_ID, EMAIL, RECV_DATE, RECV_TIME) VALUES (?, ?, ?, ?)`

	_, err := r.DB.ExecContext(ctx, query, req.ThemeID, req.Email)
	return err
}

//==============================================================
// 3. Handler 계층 (HTTP Request/Response)
//==============================================================

type Handler struct {
	Repo *DBRepository
}

func NewHandler(repo *DBRepository) *Handler {
	return &Handler{Repo: repo}
}

// GET /api/themes : 테마 목록 조회
func (h *Handler) GetThemesHandler(w http.ResponseWriter, r *http.Request) {
	themes, err := h.Repo.GetAllThemes(r.Context())
	if err != nil {
		http.Error(w, "테마 조회 실패: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(themes)
}

// GET /api/items?theme_id=XXX : 테마별 아이템 조회
func (h *Handler) GetItemsHandler(w http.ResponseWriter, r *http.Request) {
	themeID := r.URL.Query().Get("theme_id")
	if themeID == "" {
		http.Error(w, "theme_id 파라미터가 필요합니다", http.StatusBadRequest)
		return
	}

	items, err := h.Repo.GetItemsByTheme(r.Context(), themeID)
	if err != nil {
		http.Error(w, "아이템 조회 실패: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(items)
}

// GET /api/quizzes?theme_id=XXX : 테마별 퀴즈 조회
func (h *Handler) GetQuizzesHandler(w http.ResponseWriter, r *http.Request) {
	themeID := r.URL.Query().Get("theme_id")
	if themeID == "" {
		http.Error(w, "theme_id 파라미터가 필요합니다", http.StatusBadRequest)
		return
	}

	quizzes, err := h.Repo.GetQuizzesByTheme(r.Context(), themeID)
	if err != nil {
		http.Error(w, "퀴즈 조회 실패: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(quizzes)
}

// POST /api/recipient : 수령자 등록
func (h *Handler) AddRecipientHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST 메서드만 허용됩니다", http.StatusMethodNotAllowed)
		return
	}

	var req RecipientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON 파싱 오류", http.StatusBadRequest)
		return
	}

	// 간단한 유효성 검사
	if req.ThemeID == "" || req.Email == "" {
		http.Error(w, "theme_id와 email은 필수값입니다", http.StatusBadRequest)
		return
	}

	err := h.Repo.AddRecipient(r.Context(), req)
	if err != nil {
		log.Printf("수령자 등록 오류: %v", err)
		http.Error(w, "수령자 등록 중 서버 오류", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(`{"message": "success"}`))
}

//==============================================================
// 4. Main (설정 및 실행)
//==============================================================

// DBCredentials 등 기존 AWS Secrets Manager 관련 구조체는 동일하게 유지
type DBCredentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func getSecretFromAWS(secretName string, region string) (*DBCredentials, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("AWS config error: %w", err)
	}
	client := secretsmanager.NewFromConfig(cfg)
	result, err := client.GetSecretValue(context.TODO(), &secretsmanager.GetSecretValueInput{SecretId: &secretName})
	if err != nil {
		return nil, err
	}
	var creds DBCredentials
	json.Unmarshal([]byte(*result.SecretString), &creds)
	return &creds, nil
}

func main() {
	_ = godotenv.Load(".env") // 로컬 개발용

	// DB 연결 설정 (기존 로직 유지)
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASS")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")

	// AWS Secrets Manager 로직 (환경변수에 따라 실행)
	if os.Getenv("USE_SECRETS_MANAGER") == "true" {
		creds, err := getSecretFromAWS(os.Getenv("SECRET_NAME"), os.Getenv("AWS_REGION"))
		if err == nil {
			dbUser = creds.Username
			dbPass = creds.Password
			log.Println("AWS Secrets Manager credentials loaded.")
		}
	}

	if dbPort == "" {
		dbPort = "3306"
	}
	// 로컬 테스트 시 기본값
	if dbHost == "" {
		dbHost = "localhost"
		dbName = "museumdb"
		dbUser = "root"
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dbUser, dbPass, dbHost, dbPort, dbName)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("DB Open Error: %v", err)
	}

	// [추가] DB 연결 풀 설정 (time 패키지 사용으로 에러 해결 및 DB 연결 끊김 방지)
	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)

	// DB 연결 테스트 (Ping)
	if os.Getenv("SKIP_DB_CHECK") != "true" {
		if err := db.Ping(); err != nil {
			log.Printf("DB 연결 실패 (설정 확인 필요): %v", err)
		} else {
			log.Println("DB 연결 성공")
		}
	}
	defer db.Close()

	// Repository & Handler 초기화
	repo := NewDBRepository(db)
	handler := NewHandler(repo)

	mux := http.NewServeMux()

	// 라우팅
	// mux에 핸들러를 등록합니다.
	mux.HandleFunc("/api/themes", handler.GetThemesHandler)       // 1. 테마 전체 조회
	mux.HandleFunc("/api/items", handler.GetItemsHandler)         // 2. 테마별 아이템 조회
	mux.HandleFunc("/api/quizzes", handler.GetQuizzesHandler)     // 3. 테마별 퀴즈 조회
	mux.HandleFunc("/api/recipient", handler.AddRecipientHandler) // 4. 수령자 등록

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"}, // 모든 출처 허용 (친구 접속 OK)
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
		Debug:            true, // true로 하면 서버 로그에 CORS 요청 내역이 찍혀서 디버깅하기 좋습니다.
	})

	handlerWithCors := c.Handler(mux)

	port := "8080"
	if err := http.ListenAndServe(":"+port, handlerWithCors); err != nil {
		log.Fatalf("서버 실행 실패: %v", err)
	}
}
