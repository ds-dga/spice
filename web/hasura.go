package web

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// Response return message
type Response struct {
	UserID       string `json:"X-Hasura-User-Id"`
	Role         string `json:"X-Hasura-Role"`
	AllowedRoles string `json:"X-Hasura-Allowed-Roles"`
	CacheControl string `json:"Cache-Control"` // "max-age=600"
	// IsOwner      bool     `json:"X-Hasura-Is-Owner"`
	// Expires      string   `json:"Expires"`       // "Mon, 30 Mar 2020 13:25:18 GMT"
}

func jsonResponse(w http.ResponseWriter, httpCode int, resp Response) {
	w.WriteHeader(httpCode)
	w.Header().Set("Content-Type", "application/json")
	jMessage, _ := json.Marshal(resp)
	w.Write(jMessage)
}

// HasuraHook handles request from Hasura webhook auth (from headers)
func (app *WebApp) HasuraHook(w http.ResponseWriter, req *http.Request) {
	user, err := app.HeaderAuthenticate(req)
	if err != nil {
		// If you want to deny the GraphQL request, return a 401 Unauthorized exception.
		// https://hasura.io/docs/1.0/graphql/core/auth/authentication/webhook.html#failure
		w.WriteHeader(401)
		fmt.Fprintf(w, "Unauthorized")
		log.Printf("[hook-401] err=%v", err)
		log.Printf("[hook-401] req=%v", req)
		return
	}

	log.Printf("[hook-200] user=%v", user)
	log.Printf("[hook-200] req=%v", req)
	msg := Response{
		UserID:       user.ID.String(),
		Role:         "user",
		AllowedRoles: "user",
		CacheControl: "max-age=600",
	}
	jsonResponse(w, 200, msg)
}

/*

curl --header "x-everyday-app: sugar" \
	--header "x-everyday-client: next" \
	--header "x-everyday-social-app: null" \
	--header "x-everyday-uid: null" \
-XGET http://localhost:3300/hasura-hook


*/
