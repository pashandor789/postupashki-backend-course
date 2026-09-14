package main

import (
	"fmt"
	"hw2/internal/config"
	"hw2/internal/handler"
	"hw2/internal/middleware"
	"hw2/internal/repository"
	"hw2/internal/service"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func main() {
	cfg := config.Load()

	if cfg.JWTSecret == "default-secret-change-me" {
		log.Fatal("Please set JWT_SECRET in your .env file")
	}

	// 1. Initialize Repositories
	userRepo := repository.NewUserRepository()
	cryptoRepo := repository.NewCryptoRepository()

	// 2. Initialize Clients
	geckoClient := service.NewCoinGeckoClient(cfg)

	if err := geckoClient.LoadCoinList(); err != nil {
		log.Printf("Warning: Failed to load CoinGecko coin list: %v", err)
	}

	// 3. Initialize Services
	authService := service.NewAuthService(userRepo, cfg.JWTSecret)
	cryptoService := service.NewCryptoService(cryptoRepo, geckoClient)
	scheduleService := service.NewScheduleService(cryptoService)

	// Start the scheduler
	scheduleService.Start()
	defer scheduleService.Stop()

	// 4. Initialize Handlers
	authHandler := handler.NewAuthHandler(authService)
	cryptoHandler := handler.NewCryptoHandler(cryptoService)
	scheduleHandler := handler.NewScheduleHandler(scheduleService)

	// 5. Set up the router
	r := chi.NewRouter()

	// ===== PUBLIC ROUTES =====
	r.Post("/auth/register", authHandler.Register)
	r.Post("/auth/login", authHandler.Login)
	r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"status":"ok"}`))
	})

	// ===== PROTECTED ROUTES (require JWT) =====
	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(cfg.JWTSecret))

		// CRYPTO ENDPOINTS - NOTE: NO /api PREFIX!
		r.Get("/crypto", cryptoHandler.ListAll)
		r.Post("/crypto", cryptoHandler.Create)
		r.Get("/crypto/{symbol}", cryptoHandler.GetBySymbol)
		r.Put("/crypto/{symbol}/refresh", cryptoHandler.Refresh)
		r.Get("/crypto/{symbol}/history", cryptoHandler.GetHistory)
		r.Get("/crypto/{symbol}/stats", cryptoHandler.GetStats)
		r.Delete("/crypto/{symbol}", cryptoHandler.Delete)

		// SCHEDULE ENDPOINTS - NOTE: NO /api PREFIX!
		r.Get("/schedule", scheduleHandler.Get)
		r.Put("/schedule", scheduleHandler.Update)
		r.Post("/schedule/trigger", scheduleHandler.Trigger)
	})

	// 6. Start the server
	fmt.Printf("Server starting on http://localhost:%s\n", cfg.Port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", cfg.Port), r))
}
