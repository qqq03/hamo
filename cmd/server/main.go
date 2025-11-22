package main

import (
	"log"
	"net/http"
	"time"

	"hamo/internal/config"
	"hamo/internal/handler"
	"hamo/internal/repository"
	"hamo/pkg/database"

	"github.com/rs/cors"
)

// loggingMiddleware HTTP 요청 로깅 미들웨어
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// IP 주소 추출 (X-Forwarded-For 헤더 우선, 없으면 RemoteAddr)
		clientIP := r.Header.Get("X-Forwarded-For")
		if clientIP == "" {
			clientIP = r.RemoteAddr
		}

		// 응답 기록을 위한 ResponseWriter 래퍼
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		// 다음 핸들러 실행
		next.ServeHTTP(wrapped, r)

		// 로그 출력
		log.Printf("[%s] %s %s - Status: %d - Duration: %v",
			clientIP,
			r.Method,
			r.URL.Path,
			wrapped.statusCode,
			time.Since(start),
		)
	})
}

// responseWriter HTTP 응답 상태 코드를 기록하기 위한 래퍼
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func main() {
	// 1. 설정 로드
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Config load error: %v", err)
	}

	// 2. DB 연결
	db, err := database.NewMySQLConnection(
		cfg.DBUser,
		cfg.DBPass,
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBName,
	)
	if err != nil {
		log.Fatalf("DB connection error: %v", err)
	}
	defer db.Close()

	// 3. Repository & Handler 초기화
	repo := repository.NewMySQLRepository(db)
	h := handler.NewHandler(repo)

	// 4. 라우팅 설정
	mux := http.NewServeMux()
	mux.HandleFunc("/api/themes", h.GetThemesHandler)
	mux.HandleFunc("/api/items", h.GetItemsHandler)
	mux.HandleFunc("/api/quizzes", h.GetQuizzesHandler)
	mux.HandleFunc("/api/recipient", h.AddRecipientHandler)

	// 5. CORS 설정
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
		Debug:            false, // CORS 디버그 로그 비활성화 (IP 로그만 표시)
	})

	// 6. 미들웨어 체인: 로깅 -> CORS -> 라우터
	handlerWithCors := c.Handler(mux)
	handlerWithLogging := loggingMiddleware(handlerWithCors)

	// 7. 서버 시작
	log.Printf("서버가 포트 %s에서 시작됩니다.", cfg.ServerPort)
	if err := http.ListenAndServe(":"+cfg.ServerPort, handlerWithLogging); err != nil {
		log.Fatalf("서버 실행 실패: %v", err)
	}
}
