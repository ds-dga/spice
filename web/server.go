package web

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

func (app *WebApp) Serve() {

	r := chi.NewRouter()

	// A good base middleware stack
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/uptime", app.UptimeQueryByURL)
	r.Post("/uptime", app.HandleUptimeCreate)

	r.Post("/hook/uptime", app.HandleUptimeStatsUpdate)

	r.Post("/signup", app.SignUp)
	r.Post("/login", app.Login)
	r.Post("/forget-password-request", app.ForgetPassword)
	r.Get("/magic-link", app.MagicLink)
	r.Get("/email-confirmation", app.Confirmation)

	r.Route("/upload", func(r chi.Router) {
		r.Post("/", app.NewUploadHandler)
		// r.Delete("/{mediaID}", app.PurgeMediaHandler)
	})

	// For actually use, you must support HTTPS by using `ListenAndServeTLS`, reverse proxy or etc.
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})
	port := os.Getenv("PORT")
	if port == "" {
		port = "3300"
	}
	fmt.Printf("spice is listening on port %v\n", port)

	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatal(err)
	}
}

// MediaResponse contains necessary stuffs for the next stuffs
type MediaResponse struct {
	Message    string `json:"message"`
	URL        string `json:"url"`
	ObjectType string `json:"object_type"`
	ObjectID   string `json:"object_id"`
}

func MediaJSONResponse(w http.ResponseWriter, httpCode int, resp MediaResponse) {
	w.WriteHeader(httpCode)
	w.Header().Set("Content-Type", "application/json")
	jMessage, _ := json.Marshal(resp)
	w.Write(jMessage)
}

// MsgResponse contains necessary stuffs for the next stuffs
type MsgResponse struct {
	Message string `json:"message"`
}

func MessageJSONResponse(w http.ResponseWriter, httpCode int, resp MsgResponse) {
	w.WriteHeader(httpCode)
	w.Header().Set("Content-Type", "application/json")
	jMessage, _ := json.Marshal(resp)
	w.Write(jMessage)
}
