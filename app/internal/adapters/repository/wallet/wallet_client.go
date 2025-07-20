package wallet

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"kentech-project/pkg/logger"
	"net/http"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type WalletClient struct {
	baseURL    string
	httpClient *http.Client
	logger     *logger.Logger
	apiKey     string
}

type DepositRequest struct {
	Currency     string                      `json:"currency"`
	Transactions []DepositRequestTransaction `json:"transactions"`
	UserID       int                         `json:"userId"`
}

type DepositRequestTransaction struct {
	Amount    float64 `json:"amount"`
	BetID     int     `json:"betId"`
	Reference string  `json:"reference"`
}

type OperationResponse struct {
	Balance      string                         `json:"balance"`
	Transactions []OperationResponseTransaction `json:"transactions"`
}

type OperationResponseTransaction struct {
	ID        int    `json:"id"`
	Reference string `json:"reference"`
}

func NewWalletClient(baseURL string, log *logger.Logger, apiKey string) *WalletClient {
	return &WalletClient{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		logger:     log,
		apiKey:     apiKey,
	}
}

func (w *WalletClient) ProcessDeposit(ctx context.Context, userID int, amount float64, currency string, betID int, reference string) (OperationResponse, error) {
	ctx, span := otel.Tracer("").Start(ctx, "WalletClient.ProcessDeposit", trace.WithAttributes(
		attribute.String("wallet.endpoint", "/api/v1/deposit"),
		attribute.Int("wallet.user_id", userID),
		attribute.Float64("wallet.amount", amount),
	))
	defer span.End()
	return w.makeRequest(ctx, "/api/v1/deposit", userID, amount, currency, betID, reference)
}

func (w *WalletClient) ProcessWithdraw(ctx context.Context, userID int, amount float64, currency string, betID int, reference string) (OperationResponse, error) {
	ctx, span := otel.Tracer("").Start(ctx, "WalletClient.ProcessWithdraw", trace.WithAttributes(
		attribute.String("wallet.endpoint", "/api/v1/withdraw"),
		attribute.Int("wallet.user_id", userID),
		attribute.Float64("wallet.amount", amount),
	))
	defer span.End()
	return w.makeRequest(ctx, "/api/v1/withdraw", userID, amount, currency, betID, reference)
}

func (w *WalletClient) CancelTransaction(ctx context.Context, reference string) error {
	ctx, span := otel.Tracer("").Start(ctx, "WalletClient.CancelTransaction", trace.WithAttributes(
		attribute.String("wallet.endpoint", "/cancel/"+reference),
		attribute.String("wallet.reference", reference),
	))
	defer span.End()
	url := fmt.Sprintf("%s/cancel/%s", w.baseURL, reference)

	req, err := http.NewRequestWithContext(ctx, "POST", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("X-API-KEY", w.apiKey)

	resp, err := w.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("wallet service returned status: %d", resp.StatusCode)
	}

	return nil
}

func (w *WalletClient) makeRequest(ctx context.Context, endpoint string, userID int, amount float64, currency string, betID int, reference string) (OperationResponse, error) {
	request := DepositRequest{
		Currency: currency,
		UserID:   userID,
		Transactions: []DepositRequestTransaction{
			{
				Amount:    amount,
				BetID:     betID,
				Reference: reference,
			},
		},
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return OperationResponse{}, err
	}

	url := fmt.Sprintf("%s%s", w.baseURL, endpoint)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return OperationResponse{}, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-KEY", w.apiKey)

	resp, err := w.httpClient.Do(req)
	if err != nil {
		w.logger.Errorf("Wallet service request failed", "error", err)
		return OperationResponse{}, err
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	bodyString := string(bodyBytes)

	if resp.StatusCode != http.StatusOK {
		w.logger.Errorf("Wallet service returned error",
			"status", resp.StatusCode,
			"body", bodyString,
		)
		return OperationResponse{}, fmt.Errorf("wallet service returned status: %d", resp.StatusCode)
	}

	var response OperationResponse
	if err := json.Unmarshal(bodyBytes, &response); err != nil {
		w.logger.Errorf("Failed to decode wallet response", "error", err, "body", bodyString)
		return OperationResponse{}, err
	}

	return response, nil
}
