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
	// .env 파일 로더
)

//==============================================================
// 1. 모델 정의 (Model)
//==============================================================

// Document는 RAG 시스템에서 Context로 사용될 문서의 구조입니다.
// museumdb의 Item 테이블 구조에 맞춤
type Document struct {
	ThemeID   string  `json:"theme_id"`
	ItemSeq   int     `json:"item_seq"`
	ItemName  string  `json:"item_name"`
	ItemDesc  string  `json:"item_desc"`
	Latitude  float64 `json:"latitude,omitempty"`
	Longitude float64 `json:"longitude,omitempty"`
	TargetAge int     `json:"target_age,omitempty"`
}

// RAGRequest는 사용자로부터 받는 LLM 질문 요청 구조입니다.
type RAGRequest struct {
	Query string `json:"query"`
}

// RAGResponse는 LLM 응답을 클라이언트에게 반환하는 구조입니다.
type RAGResponse struct {
	Answer  string   `json:"answer"`
	Sources []string `json:"sources"` // RAG에 사용된 출처 (문서 제목 등)
}

//==============================================================
// 2. Repository 계층 (DB Access)
//==============================================================

// DBRepository는 데이터베이스 접근 메서드를 정의합니다.
type DBRepository struct {
	DB *sql.DB
}

// NewDBRepository는 DBRepository 인스턴스를 생성합니다.
func NewDBRepository(db *sql.DB) *DBRepository {
	return &DBRepository{DB: db}
}

// GetDocumentByID는 특정 ITEM_SEQ의 문서를 조회합니다. (Item 테이블 조회)
func (r *DBRepository) GetDocumentByID(ctx context.Context, itemSeq int) (*Document, error) {
	// museumdb의 Item 테이블에서 데이터를 조회하는 쿼리입니다.
	query := `SELECT THEME_ID, ITEM_SEQ, ITEM_NAME, ITEM_DESC, 
	                 COALESCE(LATITUDE, 0), COALESCE(LONGITUDE, 0), COALESCE(TARGET_AGE, 0)
	          FROM Item 
	          WHERE ITEM_SEQ = ?`
	doc := &Document{}

	// DB 연결 상태 및 오류 확인
	if r.DB == nil {
		return nil, fmt.Errorf("데이터베이스 연결이 초기화되지 않았습니다")
	}

	err := r.DB.QueryRowContext(ctx, query, itemSeq).Scan(
		&doc.ThemeID, &doc.ItemSeq, &doc.ItemName, &doc.ItemDesc,
		&doc.Latitude, &doc.Longitude, &doc.TargetAge)
	if err == sql.ErrNoRows {
		return nil, nil // 문서 없음
	}
	if err != nil {
		return nil, fmt.Errorf("문서 조회 오류: %w", err)
	}
	return doc, nil
}

// GetRelevantContext는 RAG에 필요한 Context를 조회합니다. (RAG Retrieval 단계)
// 이 예시에서는 DB에서 모든 문서를 조회하지만, 실제로는 벡터 검색 쿼리가 들어갑니다.
func (r *DBRepository) GetRelevantContext(ctx context.Context, query string) ([]*Document, error) {
	log.Printf("RAG Context Retrieval: 사용자 쿼리 '%s'에 대한 관련 문서 검색 중...", query)

	// TODO: 실제 RAG 시스템에서는 'query'의 임베딩을 생성하고,
	// 벡터 DB(OpenSearch, ChromaDB 등)에 유사도 검색(Vector Search)을 수행해야 합니다.

	// 임시 Placeholder: ID가 1과 2인 문서를 가져온다고 가정
	doc1, _ := r.GetDocumentByID(ctx, 1)
	doc2, _ := r.GetDocumentByID(ctx, 2)

	documents := make([]*Document, 0)
	if doc1 != nil {
		documents = append(documents, doc1)
	}
	if doc2 != nil {
		documents = append(documents, doc2)
	}

	return documents, nil
}

//==============================================================
// 3. Service 계층 (Business Logic & LLM/RAG)
//==============================================================

// LLMService는 비즈니스 로직과 외부 LLM API 통신을 담당합니다.
type LLMService struct {
	Repo *DBRepository
	// LLMClient는 실제 LLM API 클라이언트 구조체가 될 수 있습니다.
	// (예: Gemini Client, OpenAI Client 등)
}

// NewLLMService는 LLMService 인스턴스를 생성합니다.
func NewLLMService(repo *DBRepository) *LLMService {
	return &LLMService{Repo: repo}
}

// ProcessRAG는 RAG 전체 로직을 수행합니다.
func (s *LLMService) ProcessRAG(ctx context.Context, query string) (*RAGResponse, error) {
	// 1. Context Retrieval (Repository 호출)
	documents, err := s.Repo.GetRelevantContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("RAG Context 조회 실패: %w", err)
	}

	if len(documents) == 0 {
		return &RAGResponse{Answer: "관련 정보를 찾을 수 없습니다.", Sources: []string{}}, nil
	}

	// 2. Prompt 구성
	var contextString string
	var sources []string
	for _, doc := range documents {
		contextString += fmt.Sprintf("장소명: %s\n설명: %s\n위도/경도: %.7f, %.7f\n---\n",
			doc.ItemName, doc.ItemDesc, doc.Latitude, doc.Longitude)
		sources = append(sources, doc.ItemName)
	}

	// LLM에 전달할 최종 프롬프트
	prompt := fmt.Sprintf("다음 정보를 참고하여 사용자 질문에 가장 적절하게 답변해주세요. 정보:\n%s\n\n사용자 질문: %s", contextString, query)
	log.Printf("LLM 호출을 위한 최종 프롬프트:\n%s", prompt)

	// 3. LLM API 호출 (Placeholder)
	// TODO: 실제로 외부 LLM API에 HTTP 요청을 보내고 응답을 받아야 합니다.
	// 예: Gemini API, OpenAI API 등을 사용하는 로직 구현
	llmAnswer := fmt.Sprintf("LLM 응답: 당신의 질문 '%s'은(는) [%s] 정보를 바탕으로 처리되었습니다.", query, sources[0])

	return &RAGResponse{
		Answer:  llmAnswer,
		Sources: sources,
	}, nil
}

//==============================================================
// 4. Handler 계층 (HTTP Request/Response)
//==============================================================

// Handler는 HTTP 요청 처리를 위한 구조체입니다.
type Handler struct {
	Service *LLMService
}

// NewHandler는 Handler 인스턴스를 생성합니다.
func NewHandler(service *LLMService) *Handler {
	return &Handler{Service: service}
}

// GetDataHandler는 일반 데이터 조회를 처리하는 HTTP 핸들러입니다.
func (h *Handler) GetDataHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "GET 메서드만 허용됩니다", http.StatusMethodNotAllowed)
		return
	}

	// 쿼리 파라미터에서 ITEM_SEQ를 가져옴 (예: ?id=1 또는 ?item_seq=1)
	itemSeqStr := r.URL.Query().Get("id")
	if itemSeqStr == "" {
		itemSeqStr = r.URL.Query().Get("item_seq")
	}
	var itemSeq int
	if itemSeqStr == "" {
		http.Error(w, "item_seq 파라미터가 필요합니다 (예: ?id=1)", http.StatusBadRequest)
		return
	}
	fmt.Sscanf(itemSeqStr, "%d", &itemSeq)

	doc, err := h.Service.Repo.GetDocumentByID(r.Context(), itemSeq)
	if err != nil {
		log.Printf("DB 조회 오류: %v", err)
		http.Error(w, "데이터 조회 중 서버 오류 발생", http.StatusInternalServerError)
		return
	}

	if doc == nil {
		http.Error(w, "문서를 찾을 수 없습니다", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(doc)
}

// RAGHandler는 LLM을 이용한 RAG 처리를 담당하는 HTTP 핸들러입니다.
func (h *Handler) RAGHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "POST 메서드만 허용됩니다", http.StatusMethodNotAllowed)
		return
	}

	var req RAGRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "요청 본문(JSON) 파싱 오류", http.StatusBadRequest)
		return
	}

	if req.Query == "" {
		http.Error(w, "질문(Query) 내용이 비어있습니다", http.StatusBadRequest)
		return
	}

	// Service 계층의 RAG 처리 로직 호출
	response, err := h.Service.ProcessRAG(r.Context(), req.Query)
	if err != nil {
		log.Printf("RAG 처리 중 오류: %v", err)
		http.Error(w, "LLM 처리 중 서버 오류 발생", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(response)
}

//==============================================================
// 5. Main 함수 (Initialization & Routing)
//==============================================================

func main() {
	// .env 파일 로드 (현재 실행 파일이 있는 디렉토리에서 찾음)
	// 파일이 없어도 에러 무시 (환경 변수를 직접 사용 가능)
	err := godotenv.Load("hamo/.env")
	if err != nil {
		// hamo/.env가 없으면 현재 디렉토리의 .env 시도
		err = godotenv.Load(".env")
		if err != nil {
			log.Println("Note: .env 파일을 찾을 수 없습니다. 환경 변수를 직접 사용합니다.")
		}
	}

	// 1. 환경 변수 설정 (AWS RDS 연결 정보)
	// 실제 환경에서는 AWS Secrets Manager 또는 환경 변수를 통해 안전하게 관리해야 합니다.
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASS")
	dbHost := os.Getenv("DB_HOST") // VPC 내부 RDS Endpoint
	// DB 포트는 환경변수 DB_PORT로 오버라이드 가능합니다 (예: SSH 터널 포트 포워딩 시 사용)
	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" {
		dbPort = "3306"
	}
	dbName := os.Getenv("DB_NAME")

	// SKIP_DB_CHECK: if set to "1" or "true", skip pinging the DB (useful when using SSH/SSM port forwarding or testing without DB)
	// 개발/디버그 편의를 위해 DB 연결 확인을 건너뛸 수 있는 옵션
	// SKIP_DB_CHECK=1 또는 SKIP_DB_CHECK=true 로 설정하면 PingContext를 수행하지 않습니다.
	skipDBCheck := false
	skipEnv := os.Getenv("SKIP_DB_CHECK")
	if skipEnv == "1" || skipEnv == "true" || skipEnv == "TRUE" {
		skipDBCheck = true
	}

	// DB 설정이 없는 경우 임시로 기본값 사용 (실제 환경에서는 오류 처리 필요)
	if dbUser == "" {
		dbUser = "user"
		dbPass = ""
		dbHost = "localhost" // 테스트를 위해 localhost로 설정
		dbName = "ragdb"
		log.Println("경고: 환경 변수가 설정되지 않아 임시 DB 연결 정보를 사용합니다. 실제 AWS 환경에서 DB_USER/PASS/HOST/NAME을 설정해야 합니다.")
	}

	// 2. AWS RDS MySQL 연결 (선택적으로 건너뛰기 가능)
	var db *sql.DB
	if skipDBCheck {
		log.Println("SKIP_DB_CHECK=true: DB 연결 확인을 건너뜁니다. (로컬 포트 포워딩 사용 중일 수 있음)")
		db = nil
	} else {
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			dbUser, dbPass, dbHost, dbPort, dbName)

		var err error
		db, err = sql.Open("mysql", dsn)
		if err != nil {
			log.Fatalf("MySQL 드라이버 초기화 실패: %v", err)
		}

		// DB 연결 확인
		db.SetConnMaxLifetime(time.Minute * 3)
		db.SetMaxOpenConns(10)
		db.SetMaxIdleConns(10)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err = db.PingContext(ctx); err != nil {
			log.Fatalf("AWS RDS MySQL 연결 실패 (VPC 설정 확인 필요): %v", err)
		}
		log.Println("AWS RDS MySQL (VPC 내부) 연결 성공.")
		defer db.Close()
	}

	// 3. 계층 구조 초기화 및 의존성 주입
	repo := NewDBRepository(db)
	service := NewLLMService(repo)
	handler := NewHandler(service)

	// 4. 라우팅 설정 (Go 표준 라이브러리 사용)
	http.HandleFunc("/api/data", handler.GetDataHandler)
	http.HandleFunc("/api/rag", handler.RAGHandler)

	// 5. 서버 시작
	port := "8080"
	log.Printf("Go 서버가 포트 %s에서 시작됩니다. (EC2 배포 환경)", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("서버 실행 오류: %v", err)
	}
}

// 참고: Postman 등을 사용하여 RAGHandler 테스트 시 JSON 본문 형식
// POST /api/rag
// Body:
// {
//     "query": "Go 서버 개발 시 AWS VPC를 어떻게 구성해야 하나요?"
// }
