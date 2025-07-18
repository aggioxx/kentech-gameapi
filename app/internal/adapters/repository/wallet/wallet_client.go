package wallet

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type WalletClient struct {
	baseURL    string
	httpClient *http.Client
}

type WalletRequest struct {
	UserID uuid.UUID `json:"user_id"`
	Amount float64   `json:"amount"`
}

type WalletResponse struct {
	Reference string `json:"reference"`
	Success   bool   `json:"success"`
	Message   string `json:"message"`
}

func NewWalletClient(baseURL string) *WalletClient {
	return &WalletClient{
		baseURL:    baseURL,
		httpClient: &http.Client{},
	}
}

func (w *WalletClient) ProcessDeposit(ctx context.Context, userID uuid.UUID, amount float64) (string, error) {
	ctx, span := otel.Tracer("").Start(ctx, "WalletClient.ProcessDeposit", trace.WithAttributes(
		attribute.String("wallet.endpoint", "/deposit"),
		attribute.String("wallet.user_id", userID.String()),
		attribute.Float64("wallet.amount", amount),
	))
	defer span.End()
	return w.makeRequest(ctx, "/deposit", userID, amount)
}

func (w *WalletClient) ProcessWithdraw(ctx context.Context, userID uuid.UUID, amount float64) (string, error) {
	ctx, span := otel.Tracer("").Start(ctx, "WalletClient.ProcessWithdraw", trace.WithAttributes(
		attribute.String("wallet.endpoint", "/withdraw"),
		attribute.String("wallet.user_id", userID.String()),
		attribute.Float64("wallet.amount", amount),
	))
	defer span.End()
	return w.makeRequest(ctx, "/withdraw", userID, amount)
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

func (w *WalletClient) makeRequest(ctx context.Context, endpoint string, userID uuid.UUID, amount float64) (string, error) {
	request := WalletRequest{
		UserID: userID,
		Amount: amount,
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return "", err
	}

	url := fmt.Sprintf("%s%s", w.baseURL, endpoint)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := w.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var response WalletResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", err
	}

	if !response.Success {
		return "", fmt.Errorf("wallet service error: %s", response.Message)
	}

	return response.Reference, nil
}
