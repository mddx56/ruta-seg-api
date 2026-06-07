package service

import (
	"context"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"strings"

	"github.com/samber/do"
)

type BackupService interface {
	TriggerBackup(ctx context.Context) (string, error)
	TriggerSchemaBackup(ctx context.Context) (string, error)
	DownloadBackup(ctx context.Context) (io.ReadCloser, string, int64, error)
	DownloadSchemaBackup(ctx context.Context) (io.ReadCloser, string, int64, error)
}

type backupService struct {
	endpoint string
	apiKey   string
	client   *http.Client
}

func NewBackupService(injector *do.Injector) (BackupService, error) {
	endpoint := os.Getenv("ENDPOINT_BACKUP")
	apiKey := os.Getenv("API_KEY")

	// Ensure endpoint doesn't end with slash for consistent joining
	endpoint = strings.TrimRight(endpoint, "/")

	return &backupService{
		endpoint: endpoint,
		apiKey:   apiKey,
		client:   &http.Client{},
	}, nil
}

// TriggerBackup initiates a full backup
// Matches JSON: "backup trigger" (POST /backup/trigger)
func (s *backupService) TriggerBackup(ctx context.Context) (string, error) {
	url := fmt.Sprintf("%s/backup/trigger", s.endpoint)
	return s.doTriggerRequest(ctx, url)
}

// TriggerSchemaBackup initiates a schema-only backup
// Matches JSON: "schema trigger" (POST /backup/schema/trigger)
func (s *backupService) TriggerSchemaBackup(ctx context.Context) (string, error) {
	url := fmt.Sprintf("%s/backup/schema/trigger", s.endpoint)
	return s.doTriggerRequest(ctx, url)
}

func (s *backupService) doTriggerRequest(ctx context.Context, url string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "POST", url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("X-API-KEY", s.apiKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("%s", string(bodyBytes))
	}

	return string(bodyBytes), nil
}

// DownloadBackup retrieves the latest full backup SQL
// Matches JSON: "backup" (GET /backup/download)
func (s *backupService) DownloadBackup(ctx context.Context) (io.ReadCloser, string, int64, error) {
	url := fmt.Sprintf("%s/backup/download", s.endpoint)
	return s.doDownloadRequest(ctx, url, "backup.sql")
}

// DownloadSchemaBackup retrieves the latest schema-only backup SQL
// Matches JSON: "backup schema" (GET /backup/schema/download)
func (s *backupService) DownloadSchemaBackup(ctx context.Context) (io.ReadCloser, string, int64, error) {
	url := fmt.Sprintf("%s/backup/schema/download", s.endpoint)
	return s.doDownloadRequest(ctx, url, "schema_backup.sql")
}

func (s *backupService) doDownloadRequest(ctx context.Context, url string, defaultFilename string) (io.ReadCloser, string, int64, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, "", 0, err
	}

	req.Header.Set("X-API-KEY", s.apiKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, "", 0, err
	}

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, "", 0, fmt.Errorf("%s", string(bodyBytes))
	}

	// Extract filename
	filename := defaultFilename
	contentDisposition := resp.Header.Get("Content-Disposition")
	if contentDisposition != "" {
		_, params, err := mime.ParseMediaType(contentDisposition)
		if err == nil {
			if val, ok := params["filename"]; ok {
				filename = val
			}
		}
	}

	return resp.Body, filename, resp.ContentLength, nil
}
