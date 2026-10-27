// service/gemini_service.go
package service

import (
	"agents-dao-platform/config"
	"context"
	"fmt"

	"google.golang.org/genai"
)

type GeminiService interface {
	AnalyzeDocumentContext(ctx context.Context, prompt string, documentContent []byte) (string, error)
}

type geminiService struct {
	cfg *config.AppConfig
}

func NewGeminiService(cfg *config.AppConfig) GeminiService {
	return &geminiService{cfg: cfg}
}

func (s *geminiService) AnalyzeDocumentContext(ctx context.Context, prompt string, documentContent []byte) (string, error) {
	// 1. Initialiser le client avec le nouveau SDK Google GenAI
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		Project:  s.cfg.ProjectID,
		Location: s.cfg.Region,
		Backend:  genai.BackendVertexAI,
	})
	if err != nil {
		return "", fmt.Errorf("impossible de créer le client Gemini : %v", err)
	}

	// 2. Préparer les parties multimodales (texte + document PDF)
	parts := []*genai.Part{
		genai.NewPartFromText(prompt),
	}
	if len(documentContent) > 0 {
		parts = append(parts, genai.NewPartFromBytes(documentContent, "application/pdf"))
	}

	// 3. Appeler le modèle configuré
	result, err := client.Models.GenerateContent(ctx, s.cfg.GeminiModel,
		genai.Text(prompt),
		&genai.GenerateContentConfig{
			Temperature: genai.Ptr[float32](0.2),
		},
	)
	if err != nil {
		return "", fmt.Errorf("erreur lors de la génération par Gemini : %v", err)
	}

	return result.Text(), nil
}