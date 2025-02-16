package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"EmployeeMerchStore/api"
	"EmployeeMerchStore/config"
	"EmployeeMerchStore/internal/database"
	"EmployeeMerchStore/internal/models"
	"EmployeeMerchStore/internal/repository"
	"EmployeeMerchStore/internal/service"
	"log"
)

// CreateTestHandler инициализирует тестовый хэндлер с !!!!реальной БД!!!!
// БД надо запустить и созать там базу данных из config/config.yml
func CreateTestHandler() *api.Handler {
	// Загружаем конфиг 
	cfg, err := config.LoadConfig("config/config.yml")
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// Инициализируем базу данных
	dbPool, err := database.InitDB(cfg)
	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}
	// Запускаем миграции
	err = database.RunMigrations(context.Background(), dbPool)
	if err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	// Создаем репозитории
	userRepo := repository.NewUserRepository(dbPool)
	purchasesRepo := repository.NewPurchasesRepository(dbPool)
	ledgerRepo := repository.NewLedgerRepository(dbPool)

	// Создаем сервисы
	userService := service.NewUserService(userRepo, cfg)
	purchasesService := service.NewPurchasesService(purchasesRepo, userRepo)
	ledgerService := service.NewLedgerService(ledgerRepo, userRepo)

	// Создаем и возвращаем хэндлер
	return api.NewHandler(userService, purchasesService, ledgerService)
}

func TestAuthEndpoint(t *testing.T) {
	handler := CreateTestHandler()
	server := httptest.NewServer(api.RegisterRoutes(handler))
	defer server.Close()

	payload := map[string]string{
		"username": "testuser",
		"password": "testpassword",
	}
	data, _ := json.Marshal(payload)
	resp, err := http.Post(server.URL+"/api/auth", "application/json", bytes.NewReader(data))
	if err != nil {
		t.Fatalf("Failed to make POST /api/auth request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200 OK, got %d", resp.StatusCode)
	}
	var res struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	if res.Token == "" {
		t.Fatalf("Expected token in response")
	}
}

func TestInfoEndpoint(t *testing.T) {
	handler := CreateTestHandler()
	server := httptest.NewServer(api.RegisterRoutes(handler))
	defer server.Close()

	payload := map[string]string{
		"username": "infouser",
		"password": "infopassword",
	}
	data, _ := json.Marshal(payload)
	resp, err := http.Post(server.URL+"/api/auth", "application/json", bytes.NewReader(data))
	if err != nil {
		t.Fatalf("Failed to get token: %v", err)
	}
	defer resp.Body.Close()
	var authRes struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&authRes); err != nil {
		t.Fatalf("Failed to decode auth response: %v", err)
	}
	if authRes.Token == "" {
		t.Fatalf("Token is empty")
	}

	req, err := http.NewRequest("GET", server.URL+"/api/info", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+authRes.Token)
	infoResp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("GET /api/info request failed: %v", err)
	}
	defer infoResp.Body.Close()

	if infoResp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(infoResp.Body)
		t.Fatalf("Expected status 200, got %d: %s", infoResp.StatusCode, string(body))
	}
	// Проверяем ответ
	var info struct {
		Coins       int `json:"coins"`
		Inventory   []models.UserMerch `json:"inventory"`
		CoinHistory struct {
			Received []models.Ledger `json:"received"`
			Sent     []models.Ledger `json:"sent"`
		} `json:"coinHistory"`
	}
	if err := json.NewDecoder(infoResp.Body).Decode(&info); err != nil {
		t.Fatalf("Failed to decode info response: %v", err)
	}
}

func TestSendCoinEndpoint(t *testing.T) {
	handler := CreateTestHandler()
	server := httptest.NewServer(api.RegisterRoutes(handler))
	defer server.Close()

	// Создаем двух пользователей для перевода
	// Пользователь-отправитель
	payloadSender := map[string]string{
		"username": "senderuser",
		"password": "senderpass",
	}
	dataSender, _ := json.Marshal(payloadSender)
	respSender, err := http.Post(server.URL+"/api/auth", "application/json", bytes.NewReader(dataSender))
	if err != nil {
		t.Fatalf("Failed to create sender user: %v", err)
	}
	defer respSender.Body.Close()
	var authSender struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(respSender.Body).Decode(&authSender); err != nil {
		t.Fatalf("Failed to decode sender auth response: %v", err)
	}
	if authSender.Token == "" {
		t.Fatalf("Sender token is empty")
	}

	// Пользователь-получатель
	payloadRecipient := map[string]string{
		"username": "recipientuser",
		"password": "recipientpass",
	}
	dataRecipient, _ := json.Marshal(payloadRecipient)
	respRecipient, err := http.Post(server.URL+"/api/auth", "application/json", bytes.NewReader(dataRecipient))
	if err != nil {
		t.Fatalf("Failed to create recipient user: %v", err)
	}
	defer respRecipient.Body.Close()
	var authRecipient struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(respRecipient.Body).Decode(&authRecipient); err != nil {
		t.Fatalf("Failed to decode recipient auth response: %v", err)
	}
	if authRecipient.Token == "" {
		t.Fatalf("Recipient token is empty")
	}

	// Отправляем перевод монет от senderuser к recipientuser
	reqPayload := map[string]interface{}{
		"toUser": "recipientuser",
		"amount": 50,
	}
	dataReq, _ := json.Marshal(reqPayload)
	req, err := http.NewRequest("POST", server.URL+"/api/sendCoin", bytes.NewReader(dataReq))
	if err != nil {
		t.Fatalf("Failed to create sendCoin request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+authSender.Token)

	sendResp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("SendCoin request failed: %v", err)
	}
	defer sendResp.Body.Close()
	if sendResp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(sendResp.Body)
		t.Fatalf("Expected status 200, got %d: %s", sendResp.StatusCode, string(body))
	}
}

func TestBuyMerchEndpoint(t *testing.T) {
	handler := CreateTestHandler()
	server := httptest.NewServer(api.RegisterRoutes(handler))
	defer server.Close()

	payload := map[string]string{
		"username": "buyeruser",
		"password": "buyerpass",
	}
	data, _ := json.Marshal(payload)
	resp, err := http.Post(server.URL+"/api/auth", "application/json", bytes.NewReader(data))
	if err != nil {
		t.Fatalf("Failed to create buyer user: %v", err)
	}
	defer resp.Body.Close()
	var authResp struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		t.Fatalf("Failed to decode buyer auth response: %v", err)
	}
	if authResp.Token == "" {
		t.Fatalf("Buyer token is empty")
	}

	req, err := http.NewRequest("GET", server.URL+"/api/buy/T-Shirt", nil)
	if err != nil {
		t.Fatalf("Failed to create buy request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+authResp.Token)

	buyResp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Buy request failed: %v", err)
	}
	defer buyResp.Body.Close()
	if buyResp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(buyResp.Body)
		t.Fatalf("Expected status 200, got %d: %s", buyResp.StatusCode, string(body))
	}
}