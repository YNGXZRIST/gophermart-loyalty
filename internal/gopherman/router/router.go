package router

import (
	"gophermart-loyalty/internal/gopherman/handler/api"
	"gophermart-loyalty/internal/gopherman/middleware"

	"github.com/go-chi/chi/v5"
	mChi "github.com/go-chi/chi/v5/middleware"
)

func GetRouter(handler *api.Handler) *chi.Mux {
	r := chi.NewRouter()
	r.Group(func(r chi.Router) {
		r.Use(mChi.RequestID)
		r.Use(mChi.RealIP)
		r.Use(mChi.StripSlashes)
		r.Use(middleware.GzipCompressor)
		r.Route("/api", func(rApi chi.Router) {
			rApi.Use(mChi.RequestID)
			rApi.Use(mChi.RealIP)
			rApi.Route("/user", func(rU chi.Router) {
				rU.Post("/register", handler.Register)
				rU.Post("/login", handler.Login)
				rU.Group(func(rUserAuth chi.Router) {
					rUserAuth.Use(middleware.Authenticate)
					rUserAuth.Route("/orders", func(rUOrders chi.Router) {
						rUOrders.Get("", handler.GetOrders)
						rUOrders.Post("", handler.AddOrder)
					})
					rUserAuth.Route("/balance", func(rUBalance chi.Router) {
						rUBalance.Get("", handler.GetBalance)
						rUBalance.Post("/withdraw", handler.MakeWithdraw)
					})
					rUserAuth.Get("/withdrawals", handler.GetWithdrawals)
				})
			})
		})
	})
	return r
}
