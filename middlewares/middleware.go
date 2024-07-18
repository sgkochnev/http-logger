package middlewares

import (
	"log/slog"
	"mylog/logger"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

const userIdFiled = "user_id"

func HTTPLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

		// инициализируем контекст
		// ctx := logger.CtxInit(r.Context())
		ctx := r.Context()

		ctx = logger.CtxWith(
			ctx,
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
		)
		if r.URL.RawQuery != "" {
			ctx = logger.CtxWith(ctx, slog.String("query", r.URL.RawQuery))
		}

		start := time.Now()
		next.ServeHTTP(ww, r.WithContext(ctx))
		duration := time.Since(start)

		slog.InfoContext(ctx, "http request log",
			slog.Int("status_code", ww.Status()),
			slog.String("duration", duration.String()),
		)
	})
}

func TokenToContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		token := r.Header.Get("X-Request-Id")
		if token != "" {
			// ctx = context.WithValue(ctx, userIdFiled, token)
			ctx = logger.CtxWith(ctx, slog.String(userIdFiled, token))
		}
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
