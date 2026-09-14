package phajay

type GenerateQRResp struct {
	Message       string `json:"message"`
	TransactionID string `json:"transactionId"`
	QRCode        string `json:"qrCode"`
	Link          string `json:"link"`
}
