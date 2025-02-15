package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"EmployeeMerchStore/internal/models"
	"EmployeeMerchStore/internal/service"
	"github.com/gorilla/mux"
)

// Handler объединяет ссылки на сервисы, необходимые для обработки запросов.
type Handler struct {
	UserService      *service.UserService
	PurchasesService *service.PurchasesService
	LedgerService    *service.LedgerService
}

type Req struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// CreateUser обрабатывает POST /api/createUser.
// Тело запроса (JSON) должно содержать:
//    - username (string)
//    - password (string)
// Создает нового пользователя и возвращает JWT-токен.
func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req Req
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if req.Username == "" || req.Password == "" {
		http.Error(w, "username and password are required", http.StatusBadRequest)
		return
	}

	token, err := h.UserService.CreateUser(r.Context(), req.Username, req.Password)
	if err != nil {
		http.Error(w, "failed to create user: "+err.Error(), http.StatusInternalServerError)
		return
	}

	resp := struct {
		Token string `json:"token"`
	}{
		Token: token,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// Auth обрабатывает POST /api/auth.
// Если пользователь существует – проверяет пароль
// Возвращает JWT-токен.
func (h *Handler) Auth(w http.ResponseWriter, r *http.Request) {
	var req Req
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if req.Username == "" || req.Password == "" {
		http.Error(w, "username and password required", http.StatusBadRequest)
		return
	}

	// Пробуем аутентифицировать пользователя
	token, err := h.UserService.Auth(r.Context(), req.Username, req.Password)
	if err != nil {
		// Если ошибка содержит информацию о том, что пользователь не найден, создаем его автоматически.
		if strings.Contains(err.Error(), "no rows") {
			token, err = h.UserService.CreateUser(r.Context(), req.Username, req.Password)
			if err != nil {
				http.Error(w, "failed to create user", http.StatusInternalServerError)
				return
			}
		} else {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
	}

	resp := struct {
		Token string `json:"token"`
	}{Token: token}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// Info обрабатывает GET /api/info.
// Возвращает баланс, инвентарь и историю транзакций.
func (h *Handler) Info(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "missing authorization header", http.StatusUnauthorized)
		return
	}
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		http.Error(w, "invalid authorization header", http.StatusUnauthorized)
		return
	}
	token := parts[1]

	// Получаем информацию пользователя по токену
	balance, inventory, received, sent, err := h.UserService.GetInfo(r.Context(), token, h.PurchasesService, h.LedgerService)
	if err != nil {
		http.Error(w, "failed to get info: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var inv []models.UserMerch
	for _, item := range inventory {
		inv = append(inv, *item)
	}
	var rec []models.Ledger
	for _, tx := range received {
		rec = append(rec, *tx)
	}
	var snt []models.Ledger
	for _, tx := range sent {
		snt = append(snt, *tx)
	}

	resp := struct {
		Coins       int                 `json:"coins"`
		Inventory   []models.UserMerch  `json:"inventory"`
		CoinHistory struct {
			Received []models.Ledger `json:"received"`
			Sent     []models.Ledger `json:"sent"`
		} `json:"coinHistory"`
	}{
		Coins:     balance,
		Inventory: inv,
		CoinHistory: struct {
			Received []models.Ledger `json:"received"`
			Sent     []models.Ledger `json:"sent"`
		}{
			Received: rec,
			Sent:     snt,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// SendCoin обрабатывает POST /api/sendCoin.
// Ожидает JSON с полями toUser (имя получателя) и amount (количество монет).
func (h *Handler) SendCoin(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "missing authorization header", http.StatusUnauthorized)
		return
	}
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		http.Error(w, "invalid authorization header", http.StatusUnauthorized)
		return
	}
	token := parts[1]
	senderID, err := h.UserService.DecodeToken(token)
	if err != nil {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}

	var req struct {
		ToUser string  `json:"toUser"`
		Amount int `json:"amount"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if req.ToUser == "" || req.Amount <= 0 {
		http.Error(w, "toUser and positive amount required", http.StatusBadRequest)
		return
	}

	// Выполняем перевод монет
	if err := h.LedgerService.SendMoney(r.Context(), senderID, req.ToUser, req.Amount); err != nil {
		http.Error(w, "failed to send coin: "+err.Error(), http.StatusBadRequest)
		return
	}

	resp := struct {
		Message string `json:"message"`
	}{Message: "Coin transfer successful"}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// BuyMerch обрабатывает GET /api/buy/{item}.
// Выполняется покупка мерча за монеты.
func (h *Handler) BuyMerch(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "missing authorization header", http.StatusUnauthorized)
		return
	}
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		http.Error(w, "invalid authorization header", http.StatusUnauthorized)
		return
	}
	token := parts[1]
	userID, err := h.UserService.DecodeToken(token)
	if err != nil {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	item := vars["item"]
	if item == "" {
		http.Error(w, "item parameter is required", http.StatusBadRequest)
		return
	}

	// Выполняем покупку мерча
	if err := h.PurchasesService.BuyMerch(r.Context(), userID, item); err != nil {
		http.Error(w, "failed to buy merch: "+err.Error(), http.StatusBadRequest)
		return
	}

	resp := struct {
		Message string `json:"message"`
	}{Message: "Purchase successful"}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}