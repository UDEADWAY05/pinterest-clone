package handlers

import (
	"net/http"
)

// Health проверка работоспособности сервера.
// @Summary      Health check
// @Description  Возвращает 200 OK, если сервер работает
// @Tags         system
// @Success      200  {string}  string  "ok"
// @Router       /health [get]
func Health(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
    _, _ = w.Write([]byte("ok"))
}
