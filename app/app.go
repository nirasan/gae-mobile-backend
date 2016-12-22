package app

import (
	"net/http"
	"github.com/nirasan/gae-mobile-backend/handler"
)

func init() {
	http.Handle("/", handler.NewHandler())
}
