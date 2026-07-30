package fcm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const messagingScope = "https://www.googleapis.com/auth/firebase.messaging"

// Client envía push notifications vía la API HTTP v1 de Firebase Cloud Messaging,
// usando una cuenta de servicio para autenticarse (sin depender del SDK completo
// firebase-admin-go, mucho más pesado en dependencias).
type Client struct {
	projectID   string
	tokenSource oauth2.TokenSource
	httpClient  *http.Client
}

// NewClientFromFile crea un Client leyendo las credenciales de una cuenta de servicio
// de Firebase/GCP desde un archivo JSON (variable de entorno FCM_CREDENTIALS_FILE).
func NewClientFromFile(ctx context.Context, credentialsPath string) (*Client, error) {
	data, err := os.ReadFile(credentialsPath)
	if err != nil {
		return nil, fmt.Errorf("fcm: no se pudo leer %s: %w", credentialsPath, err)
	}
	return NewClientFromJSON(ctx, data)
}

// NewClientFromJSON crea un Client a partir del contenido JSON de la cuenta de servicio
// (útil cuando las credenciales vienen en una variable de entorno en vez de un archivo).
func NewClientFromJSON(ctx context.Context, credentialsJSON []byte) (*Client, error) {
	var raw struct {
		ProjectID string `json:"project_id"`
	}
	if err := json.Unmarshal(credentialsJSON, &raw); err != nil {
		return nil, fmt.Errorf("fcm: credenciales inválidas: %w", err)
	}
	if raw.ProjectID == "" {
		return nil, fmt.Errorf("fcm: las credenciales no traen project_id")
	}

	creds, err := google.CredentialsFromJSON(ctx, credentialsJSON, messagingScope)
	if err != nil {
		return nil, fmt.Errorf("fcm: error preparando credenciales: %w", err)
	}

	return &Client{
		projectID:   raw.ProjectID,
		tokenSource: creds.TokenSource,
		httpClient:  &http.Client{Timeout: 10 * time.Second},
	}, nil
}

// Message es una notificación push dirigida a un solo token de dispositivo.
type Message struct {
	Token string
	Title string
	Body  string
	Data  map[string]string
}

// Send envía msg vía la API HTTP v1 de FCM. Si FCM responde con un error (p.ej. token
// inválido/expirado), devuelve un *SendError que el llamador puede inspeccionar para
// decidir si debe desactivar ese token.
func (c *Client) Send(ctx context.Context, msg Message) error {
	token, err := c.tokenSource.Token()
	if err != nil {
		return fmt.Errorf("fcm: error obteniendo access token: %w", err)
	}

	payload := map[string]interface{}{
		"message": map[string]interface{}{
			"token": msg.Token,
			"notification": map[string]string{
				"title": msg.Title,
				"body":  msg.Body,
			},
			"data": msg.Data,
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("fcm: error serializando mensaje: %w", err)
	}

	url := fmt.Sprintf("https://fcm.googleapis.com/v1/projects/%s/messages:send", c.projectID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("fcm: error de red enviando push: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return &SendError{StatusCode: resp.StatusCode, Body: string(respBody)}
	}

	return nil
}

// SendError representa una respuesta de error de la API de FCM (p.ej. token inválido).
type SendError struct {
	StatusCode int
	Body       string
}

func (e *SendError) Error() string {
	return fmt.Sprintf("fcm: respuesta %d: %s", e.StatusCode, e.Body)
}

// IsInvalidToken indica si el error implica que el token de dispositivo ya no es válido
// (desinstalado, expirado) y debería eliminarse para no seguir intentando enviarle.
func (e *SendError) IsInvalidToken() bool {
	return e.StatusCode == http.StatusNotFound || e.StatusCode == http.StatusBadRequest
}
