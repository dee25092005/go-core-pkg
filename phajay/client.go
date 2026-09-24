package phajay

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/dee25092005/go-core-pkg/apperrors"
)

type Client struct {
	secretKey  string
	httpClient *http.Client
}

func getBankEndpoint(bankCode string) (string, error) {
	switch bankCode {
	case "bcel":
		return "https://payment-gateway.phajay.co/v1/api/payment/generate-bcel-qr", nil
	case "jdb":
		return "https://payment-gateway.phajay.co/v1/api/payment/generate-jdb-qr", nil
	case "ldb":
		return "https://payment-gateway.phajay.co/v1/api/payment/generate-ldb-qr", nil
	case "ib":
		return "https://payment-gateway.phajay.co/v1/api/payment/generate-ib-qr", nil
	case "stb":
		return "https://payment-gateway.phajay.co/v1/api/payment/generate-stb-qr", nil
	default:
		return "", fmt.Errorf("unsupported bank code: %s", bankCode)
	}
}

func NewClient(secretKey string) *Client {
	return &Client{
		secretKey:  secretKey,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *Client) GenerateQR(ctx context.Context, bankCode string, req GenerateQRRequest) (*GenerateQRResp, error) {
	endpoint, err := getBankEndpoint(bankCode)
	if err != nil {
		return nil, apperrors.BadRequest("invalid or unsupported bank code")
	}

	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, apperrors.Internal(err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, apperrors.BadRequest("failed to create http request")
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("secretKey", c.secretKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, apperrors.BadRequest("failed to connect to phajay gateway")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, apperrors.BadRequest("failed to read response body")
		}
		errorMessage := string(bodyBytes)

		return nil, apperrors.BadRequest(errorMessage)
	}

	var phajayResp GenerateQRResp
	if err := json.NewDecoder(resp.Body).Decode(&phajayResp); err != nil {
		return nil, apperrors.Internal(err)
	}

	return &phajayResp, nil

}
