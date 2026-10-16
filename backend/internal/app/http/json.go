// Package http — общие помощники HTTP-транспорта всех доменов.
package http

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"expensebot/internal/app/apperr"
)

func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v) // заголовки уже отправлены, сообщить об ошибке некуда
}

func WriteError(w http.ResponseWriter, status int, msg string) {
	WriteJSON(w, status, map[string]string{"error": msg})
}

// Fail отвечает по коду apperr; остальные ошибки — 500 без подробностей наружу.
func Fail(w http.ResponseWriter, log *slog.Logger, err error) {
	e, ok := apperr.As(err)
	switch {
	case ok && e.Code == apperr.CodeValidation:
		WriteJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": e.Message, "field": e.Field})
	case ok && e.Code == apperr.CodeNotFound:
		WriteError(w, http.StatusNotFound, e.Message)
	default:
		log.Error("api request failed", "err", err)
		WriteError(w, http.StatusInternalServerError, "внутренняя ошибка")
	}
}

func Decode(w http.ResponseWriter, r *http.Request, dst any) bool {
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 64<<10))
	dec.DisallowUnknownFields()

	if err := dec.Decode(dst); err != nil {
		WriteError(w, http.StatusBadRequest, "некорректный JSON")
		return false
	}

	return true
}

func PathID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		WriteError(w, http.StatusNotFound, "не найдено")
		return 0, false
	}

	return id, true
}

// QueryInt — неотрицательное число из query; пустое значение — 0.
func QueryInt(r *http.Request, name string) (int64, error) {
	v := r.URL.Query().Get(name)
	if v == "" {
		return 0, nil
	}

	n, err := strconv.ParseInt(v, 10, 64)
	if err != nil || n < 0 {
		return 0, apperr.ErrValidation(name, "должно быть неотрицательным числом")
	}

	return n, nil
}

// Recovery превращает панику обработчика в 500, не роняя процесс.
func Recovery(next http.Handler, log *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if v := recover(); v != nil {
				log.Error("panic in handler", "path", r.URL.Path, "panic", v)
				WriteError(w, http.StatusInternalServerError, "внутренняя ошибка")
			}
		}()

		next.ServeHTTP(w, r)
	})
}
