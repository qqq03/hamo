package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"hamo/internal/model"
	"hamo/internal/repository"
)

// Handler HTTP 요청 처리
type Handler struct {
	Repo repository.Repository
}

// NewHandler Handler 생성
func NewHandler(repo repository.Repository) *Handler {
	return &Handler{Repo: repo}
}

// GetThemesHandler GET /api/themes : 테마 목록 조회
func (h *Handler) GetThemesHandler(w http.ResponseWriter, r *http.Request) {
	themes, err := h.Repo.GetAllThemes(r.Context())
	if err != nil {
		http.Error(w, "테마 조회 실패: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(themes)
}

// GetItemsHandler GET /api/items?theme_id=XXX : 테마별 아이템 조회
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

// GetQuizzesHandler GET /api/quizzes?theme_id=XXX : 테마별 퀴즈 조회
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

// AddRecipientHandler POST /api/recipient : 수령자 등록
func (h *Handler) AddRecipientHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST 메서드만 허용됩니다", http.StatusMethodNotAllowed)
		return
	}

	var req model.RecipientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON 파싱 오류", http.StatusBadRequest)
		return
	}

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
