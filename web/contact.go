package web

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type PublicSectorBody struct {
	Query string
}

type PublicSectorObject struct {
	ID       int64  `json:"id"`
	Ministry string `json:"ministry"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Tel      string `json:"tel"`
}

type ContactResponse struct {
	Message string               `json:"message"`
	Result  []PublicSectorObject `json:"result"`
}

// PublicSectorContact returns the contact for public sector and records what not found for improving later
func (app *WebApp) PublicSectorContact(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var body PublicSectorBody
	q := req.URL.Query()
	body.Query = q.Get("query")
	if len(body.Query) == 0 {
		w.WriteHeader(http.StatusNotFound)
		msg := []byte(`{"message":"not found"}`)
		w.Write(msg)
		return
	}
	obj, err := app.FindContact(fmt.Sprintf("%%%s%%", body.Query))
	if err != nil {
		_, err := app.CreateUnknownContact(body.Query)
		if err != nil {
			w.WriteHeader(http.StatusBadGateway)
			w.Write([]byte(fmt.Sprintf(`{"message":"%v"}`, err.Error())))
			return
		}
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"message":"not found"}`))
		return
	}
	w.WriteHeader(http.StatusOK)
	result := []PublicSectorObject{*obj}
	resp := ContactResponse{
		Message: "success",
		Result:  result,
	}
	jMessage, _ := json.Marshal(resp)
	w.Write(jMessage)
}

// FindContact returns media object from uuid query
func (app *WebApp) FindContact(query string) (*PublicSectorObject, error) {
	var m PublicSectorObject
	err := app.contactDB.QueryRow(`
		SELECT m.id, m.ministry, m.name, m.email, m.tel
		FROM dga_contacts m
		WHERE m.name LIKE $1 OR m.ministry LIKE $1`, query).Scan(
		&m.ID, &m.Ministry, &m.Name, &m.Email, &m.Tel,
	)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

// CreateUnknownContact stores what we can't find on `dga_contacts`
func (app *WebApp) CreateUnknownContact(query string) (int64, error) {
	var result int64
	err := app.contactDB.QueryRow(`
		INSERT INTO missing_contacts ("name")
		VALUES ($1)
		ON CONFLICT ON CONSTRAINT missing_contacts_name_key
		DO UPDATE SET total = missing_contacts."total" + 1
		RETURNING id
		`, query).Scan(&result)
	if err != nil {
		log.Printf("[CONTACT] create failed [=%v] %v", result, err)
		return -1, err
	}
	return result, nil
}
