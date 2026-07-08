package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	bkanthropic "byakugan/internal/anthropic"
	"byakugan/internal/corpus"
	"byakugan/internal/store"
	"byakugan/internal/voyage"

	"github.com/joho/godotenv"
)

type askRequest struct {
	Question string      `json:"question"`
	Lang     corpus.Lang `json:"lang"`
}

type askResponse struct {
	Question string `json:"question"`
	Lang     string `json:"lang"`
	Answer   string `json:"answer"`
}

type byakuganServer struct {
	voyage    *voyage.Client
	store     *store.Store
	anthropic *bkanthropic.Client
}

func main() {
	ctx := context.Background()
	mux := http.NewServeMux()
	_ = godotenv.Load()

	server := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 240 * time.Second,
		IdleTimeout:  420 * time.Second,
	}

	voyageKey := os.Getenv("VOYAGE_API_KEY")
	if voyageKey == "" {
		log.Fatal("Voyage API key not set")
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://byakugan:byakugan@localhost:5433/byakugan"
	}

	anthropicKey := os.Getenv("ANTHROPIC_API_KEY")
	if anthropicKey == "" {
		log.Fatal("Anthropic API key not set")
	}

	st, err := store.Connect(ctx, dbURL)
	if err != nil {
		log.Fatalf("Could not connect to store: %v", err)
	}

	// Dependency injection
	srv := &byakuganServer{
		voyage:    voyage.New(voyageKey),
		store:     st,
		anthropic: bkanthropic.New(anthropicKey),
	}

	mux.HandleFunc("POST /ask", srv.handleAsk)

	fmt.Println("Server is running on http://localhost:8080")
	log.Fatal(server.ListenAndServe())
}

func (s *byakuganServer) handleAsk(w http.ResponseWriter, r *http.Request) {
	var req askRequest
	w.Header().Set("Content-Type", "application/json")

	searchK := 20
	topK := 5

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("error decoding req: %v", err)
		http.Error(w, "request does not follow expected json schema", http.StatusBadRequest)
		return
	}

	if len(req.Question) == 0 {
		http.Error(w, "Cannot ask an empty question", http.StatusBadRequest)
		return
	}

	// Default to english on invalid Lang param
	if !req.Lang.Valid() {
		log.Println("Defaulting to en...")
		req.Lang = "en"
	}

	vectors, err := s.voyage.Embed(r.Context(), voyage.Query, []string{req.Question})
	if err != nil {
		log.Printf("failed to embed question: %v", err)
		http.Error(w, "byakugan is having trouble", http.StatusServiceUnavailable)
		return
	}

	if len(vectors) == 0 {
		log.Printf("embed returned no vectors")
		http.Error(w, "byakugan is having trouble", http.StatusServiceUnavailable)
		return
	}

	results, err := s.store.Search(r.Context(), vectors[0], searchK, req.Lang)
	if err != nil {
		log.Printf("trouble searching DB: %v", err)
		http.Error(w, "byakugan's DB is having a sick day", http.StatusServiceUnavailable)
		return
	}

	if len(results) == 0 {
		json.NewEncoder(w).Encode(askResponse{
			Question: req.Question,
			Lang:     string(req.Lang),
			Answer:   "I can't find an answer for that - try rephrasing or adding some citations?",
		})
		return
	}

	var resultTexts []string
	for _, result := range results {
		resultTexts = append(resultTexts, result.Text)
	}

	reranked, err := s.voyage.Rerank(r.Context(), req.Question, resultTexts, topK)
	if err != nil {
		log.Printf("Reranker had some trouble: %v", err)
		http.Error(w, "byakugan's brain had a spasm", http.StatusServiceUnavailable)
		return
	}

	// reranked hits
	newHits := make([]store.Hit, 0, len(reranked))
	for _, rr := range reranked {
		newHits = append(newHits, results[rr.Index])
	}

	answer, interrupted, err := s.anthropic.Frame(r.Context(), req.Question, newHits)
	if err != nil {
		log.Printf("Claude API error: %v", err)
		http.Error(w, "byakugan is OK. our LLM ai is not. please try again later. (^._.^)ﾉ", http.StatusServiceUnavailable)
		return
	}
	if interrupted {
		log.Printf("[claude] answer may be truncated")
	}

	res := askResponse{
		Question: req.Question,
		Lang:     string(req.Lang),
		Answer:   fmt.Sprintf("Byakugan's Reply:\n%s", answer),
	}

	json.NewEncoder(w).Encode(res)
}
