package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/KEPTANy/plag-check/analysis-service/internal/model"
	"github.com/KEPTANy/plag-check/analysis-service/internal/repository"
	"github.com/KEPTANy/plag-check/analysis-service/internal/storage"
)

type AnalysisService interface {
	CheckPlagiarism(ctx context.Context) ([]model.PlagiarismResult, error)
	GetWordCloud(ctx context.Context, fileID int) ([]byte, error)
}

type analysisService struct {
	db      repository.FileRepository
	storage storage.Storage
}

func NewAnalysisService(db repository.FileRepository, storage storage.Storage) AnalysisService {
	return &analysisService{db: db, storage: storage}
}

func (s *analysisService) CheckPlagiarism(ctx context.Context) ([]model.PlagiarismResult, error) {
	results, err := s.db.GetPlagiarismGroups(ctx)
	if err != nil {
		return nil, fmt.Errorf("Failed to get plagiarism groups: %w", err)
	}

	return results, nil
}

func (s *analysisService) GetWordCloud(ctx context.Context, fileID int) ([]byte, error) {
	file, err := s.db.GetFileByID(ctx, fileID)
	if err != nil {
		return nil, fmt.Errorf("Failed to get file info: %w", err)
	}

	fileReader, _, err := s.storage.GetFile(ctx, file.StoragePath)
	if err != nil {
		return nil, fmt.Errorf("Failed to read file: %w", err)
	}
	defer fileReader.Close()

	fileContent, err := io.ReadAll(fileReader)
	if err != nil {
		return nil, fmt.Errorf("Failed to read file content: %w", err)
	}

	text := string(fileContent)

	wordCloudConfig := map[string]interface{}{
		"type": "wordcloud",
		"data": map[string]interface{}{
			"text": text,
		},
	}

	configJSON, err := json.Marshal(wordCloudConfig)
	if err != nil {
		return nil, fmt.Errorf("Failed to marshal word cloud config: %w", err)
	}

	apiURL := fmt.Sprintf("https://quickchart.io/chart?c=%s", url.QueryEscape(string(configJSON)))

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("Failed to call quickchart.io API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("QuickChart API returned error: %d, body: %s", resp.StatusCode, string(body))
	}

	imageData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Failed to read image data: %w", err)
	}

	return imageData, nil
}
