package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/rasasaufar/finance-app/api/internal/handler"
	"github.com/rasasaufar/finance-app/api/internal/middleware"
	"github.com/rasasaufar/finance-app/api/internal/store"
)

func main() {
	databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if databaseURL == "" {
		databaseURL = "postgres://finance_user:finance_pass@localhost:5432/finance_app?sslmode=disable"
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		log.Fatalf("gagal inisialisasi postgres: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("gagal konek ke postgres: %v", err)
	}

	st := store.New(pool)
	if err := st.EnsureSchema(ctx); err != nil {
		log.Fatalf("gagal setup schema: %v", err)
	}
	if err := st.SeedDefaults(ctx); err != nil {
		log.Fatalf("gagal seed data default: %v", err)
	}

	h := handler.New(st)
	r := chi.NewRouter()

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{
			"http://localhost:5173",
			"http://localhost:5174",
		},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
	}))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	r.Post("/auth/login", h.HandleLogin)

	r.Group(func(pr chi.Router) {
		pr.Use(middleware.Auth)

		pr.Get("/me", h.HandleMe)
		pr.Put("/me", h.HandleUpdateMe)

		pr.Get("/dashboard/summary", h.HandleDashboardSummary)

		pr.Get("/transactions", h.HandleGetTransactions)
		pr.Post("/transactions", h.HandleCreateTransaction)
		pr.Put("/transactions/{id}", h.HandleUpdateTransaction)
		pr.Delete("/transactions/{id}", h.HandleDeleteTransaction)

		pr.Get("/categories", h.HandleGetCategories)
		pr.Post("/categories", h.HandleCreateCategory)
		pr.Put("/categories/{id}", h.HandleUpdateCategory)
		pr.Delete("/categories/{id}", h.HandleDeleteCategory)

		pr.Get("/budget-rules", h.HandleGetBudgetRules)
		pr.Post("/budget-rules", h.HandleCreateBudgetRule)
		pr.Put("/budget-rules/{id}", h.HandleUpdateBudgetRule)
		pr.Delete("/budget-rules/{id}", h.HandleDeleteBudgetRule)

		pr.Get("/salary-masters", h.HandleGetSalaryMasters)
		pr.Post("/salary-masters", h.HandleCreateSalaryMaster)
		pr.Put("/salary-masters/{id}", h.HandleUpdateSalaryMaster)
		pr.Delete("/salary-masters/{id}", h.HandleDeleteSalaryMaster)

		pr.Get("/reports/monthly", h.HandleMonthlyReport)
	})

	log.Printf("API running on http://localhost:8080 (postgres connected)")

	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatal(err)
	}
}
