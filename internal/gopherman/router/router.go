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
		r.Use(mChi.StripSlashes)
		r.Route("/api", func(rApi chi.Router) {
			rApi.Route("/user", func(rU chi.Router) {
				rU.Group(func(rJson chi.Router) {
					rJson.Use(middleware.ContentTypeJSON)
					rJson.Post("/register", handler.Register)
					rJson.Post("/login", handler.Login)
				})

				rU.Group(func(rUserAuth chi.Router) {
					rUserAuth.Use(middleware.Authenticate(handler))
					rUserAuth.Route("/orders", func(rUOrders chi.Router) {
						rUOrders.Post("/", handler.AddOrder)
						rUOrders.Group(func(rUOrdersGet chi.Router) {
							rUOrdersGet.Use(middleware.GzipCompressor)
							rUOrdersGet.Get("/", handler.GetOrders)
						})
					})

					rUserAuth.Route("/balance", func(rUBalance chi.Router) {
						rUBalance.Get("/", handler.GetBalance)
						rUBalance.Post("/withdraw", handler.MakeWithdraw)
					})

					rUserAuth.Get("/withdrawals", handler.GetWithdrawals)

				})
			})
		})
	})
	return r
}
