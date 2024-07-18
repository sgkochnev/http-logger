package main

import (
	"context"
	"errors"
	"log/slog"
	"mylog/logger"
	"mylog/middlewares"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
)

type Phone string

const (
	phoneDefault Phone = "+78880001100"
	phone1       Phone = "+78880001101"
	phone2       Phone = "+78880001102"
)

func (p Phone) LogValue() slog.Value {
	return slog.StringValue(strings.Repeat("*", len(p)-4) + string(p[len(p)-4:]))
}

func TransmitSMS(ctx context.Context, gate, message string, phone Phone) error {
	if phone == phone2 {
		return errors.New("transmit SMS: network error")
	}

	// slog.InfoContext(ctx, "Transmit SMS gateway OK")
	return nil
}

// -----------------------------------------------

func ResolveGate(ctx context.Context, phone Phone) (string, error) {
	if phone == phone1 {
		return "", errors.New("gate not found")
	}
	gate := "RHO"
	return gate, nil
}

// -----------------------------------------------

func SendSMS(ctx context.Context, phone Phone) error {
	gate, err := ResolveGate(ctx, phone)
	if err != nil {
		return err
	}
	logger.Group(ctx, "send_sms", slog.String("gate", gate))

	message := "Спасибо"
	logger.Group(ctx, "send_sms", slog.String("message", message))

	err = TransmitSMS(ctx, gate, message, phone)
	if err != nil {
		return err
	}
	return nil
}

// -----------------------------------------------

func GetPhoenByID(ctx context.Context, userID int) (Phone, error) {
	phone := phoneDefault

	switch userID {
	case 1:
		phone = phone1
	case 2:
		phone = phone2
	case 3:
		return "", errors.New("phone not found")
	}

	return phone, nil
}

// -----------------------------------------------

func ServiceAccept(ctx context.Context, userID int) error {
	phone, err := GetPhoenByID(ctx, userID)
	if err != nil {
		return err
	}
	logger.With(ctx, slog.Any("phone", phone))

	err = SendSMS(ctx, phone)
	if err != nil {
		return err
	}

	return nil
}

type ctxKey string

const pSlogFields ctxKey = "p_slog_fields"

func ProcessHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.Atoi(strings.TrimSpace(r.Header.Get("X-Request-Id")))

	// инициализируем контекст если не инициализирован в миддлваре
	ctx := logger.CtxInit(r.Context())

	// ctx := context.Background()
	// var p *[]slog.Attr
	// ctx = context.WithValue(ctx, pSlogFields, p)
	// With(ctx, slog.Int("user_id", 1111))

	// ctx := context.WithValue(r.Context(), pSlogFields, &[]slog.Attr{})
	err = ServiceAccept(ctx, userID)
	if err == nil {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello, World!"))
		slog.InfoContext(ctx, "ServiceAccept")
		return
	}
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte("SomethingWentWrong!"))
	slog.ErrorContext(ctx, "SomethingWentWrong", slog.String("error", err.Error()))
	return
}

func main() {
	if err := logger.InitLogger("app", "1.0.0", "info"); err != nil {
		return
	}

	r := chi.NewRouter()
	r.Use(
		middlewares.TokenToContextMiddleware,
		middlewares.HTTPLogger,
	)

	r.Get("/sms", ProcessHandler)

	http.ListenAndServe(":3000", r)
}
