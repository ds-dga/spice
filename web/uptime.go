package web

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/google/uuid"
)

func (app *WebApp) UptimeQueryByURL(w http.ResponseWriter, r *http.Request) {
	/* Get only

	Query: url	text
	*/
	url := r.URL.Query().Get("url")
	if url == "" {
		MessageJSONResponse(w, http.StatusNotAcceptable, MsgResponse{
			Message: "not acceptable",
		})
		return
	}
	rec, err := app.getUptimeRecordByURL(url)
	if err != nil {
		errMsg := err.Error()
		if ind := strings.Index(errMsg, "no rows"); ind != -1 {
			MessageJSONResponse(w, http.StatusNotFound, MsgResponse{
				Message: "not found",
			})
			return
		}
		log.Printf("[err1] %s", err.Error())
		MessageJSONResponse(w, http.StatusNotAcceptable, MsgResponse{
			Message: err.Error(),
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	jMessage, _ := json.Marshal(rec)
	w.Write(jMessage)
}

type ReqUptimeCreateData struct {
	Url       string           `json:"url"`
	Name      string           `json:"name"`
	Frequency string           `json:"frequency"`
	Group     string           `json:"group"`
	Extras    *json.RawMessage `json:"extras"`
}

func (app *WebApp) HandleUptimeCreate(w http.ResponseWriter, r *http.Request) {
	/* POST only

	body:

	* url 			text
	* name			text
	* frequency		text
	* group			text
	* extras		json
	*/

	if r.Body == nil {
		MessageJSONResponse(w, http.StatusNotAcceptable, MsgResponse{
			Message: "No request body",
		})
		return
	}
	var body ReqUptimeCreateData
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		MessageJSONResponse(w, http.StatusNotAcceptable, MsgResponse{
			Message: err.Error(),
		})
		return
	}
	if body.Url == "" {
		MessageJSONResponse(w, http.StatusNotAcceptable, MsgResponse{
			Message: "empty url",
		})
		return
	}
	_, err = url.Parse(body.Url)
	if err != nil {
		MessageJSONResponse(w, http.StatusNotAcceptable, MsgResponse{
			Message: "invalid url",
		})
		return
	}
	b, err := json.Marshal(&body.Extras)
	if err != nil {
		MessageJSONResponse(w, http.StatusNotAcceptable, MsgResponse{
			Message: "invalid extras",
		})
		return
	}

	rec, err := app.createUptimeRecord(body.Url, body.Name, body.Frequency, body.Group, string(b))
	if err != nil {
		MessageJSONResponse(w, http.StatusNotAcceptable, MsgResponse{
			Message: err.Error(),
		})
		return
	}
	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	jMessage, _ := json.Marshal(rec)
	w.Write(jMessage)
}

type ReqStatsUpdateData struct {
	ID           uuid.UUID  `json:"id"`
	Url          string     `json:"url"`
	StatusCode   int64      `json:"status_code"`
	ResponseTime float64    `json:"response_time_ms"`
	Size         float64    `json:"size_byte"`
	From         string     `json:"from"`
	Coords       [2]float64 `json:"from_coords"`
}

func (app *WebApp) HandleUptimeStatsUpdate(w http.ResponseWriter, r *http.Request) {
	/* POST only

	body:

	* url (or id)
	* status_code
	* response_time_ms
	* size_byte
	* from
	* from_coords
	*/
	if r.Body == nil {
		MessageJSONResponse(w, http.StatusNotAcceptable, MsgResponse{
			Message: "No request body",
		})
		return
	}
	var body ReqStatsUpdateData
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		MessageJSONResponse(w, http.StatusNotAcceptable, MsgResponse{
			Message: err.Error(),
		})
		return
	}
	log.Printf("Got: %v", body)
	if body.Url == "" && body.ID.String() == "00000000-0000-0000-0000-000000000000" {
		MessageJSONResponse(w, http.StatusNotAcceptable, MsgResponse{
			Message: "No id or url found",
		})
		return
	}

	if body.StatusCode == 0 {
		MessageJSONResponse(w, http.StatusNotAcceptable, MsgResponse{
			Message: "No status code",
		})
		return
	}

	rec, err := app.GetOrCreateUptimeRecord(body.ID, body.Url)
	if err != nil {
		MessageJSONResponse(w, http.StatusNotAcceptable, MsgResponse{
			Message: err.Error(),
		})
		return
	}

	err = app.addUptimeStat(rec, body)
	if err != nil {
		MessageJSONResponse(w, http.StatusBadGateway, MsgResponse{
			Message: err.Error(),
		})
		return
	}
	MessageJSONResponse(w, http.StatusCreated, MsgResponse{Message: "OK"})
}

type UptimeRecord struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Url       string    `json:"url"`
	Group     string    `json:"group"`
	Frequency string    `json:"frequency"`
}

func (app *WebApp) GetOrCreateUptimeRecord(ID uuid.UUID, Url string) (*UptimeRecord, error) {
	// ID is the first priority
	if ID.String() != "00000000-0000-0000-0000-000000000000" {
		rec, err := app.getUptimeRecordByID(ID)
		if err == nil {
			// if found, return record
			return rec, nil
		}
	}
	rec, err := app.getUptimeRecordByURL(Url)
	if err != nil {
		// It's likely that there is no record exists
		log.Printf("Errror: %v", err.Error())
		rec, err = app.addUptimeRecord(Url)
		if err != nil {
			return nil, err
		}
	}
	return rec, nil
}

func (app *WebApp) getUptimeRecordByID(ID uuid.UUID) (*UptimeRecord, error) {
	var rec UptimeRecord
	var group sql.NullString
	var freq sql.NullString
	err := app.pdb.QueryRow(`
		SELECT "id", "name", "url", "group", "frequency"
		FROM api
		WHERE id = $1`, ID).Scan(&rec.ID, &rec.Name, &rec.Url, &group, &freq)
	if err != nil {
		return nil, err
	}
	if group.Valid {
		rec.Group = group.String
	}
	if freq.Valid {
		rec.Frequency = freq.String
	}
	return &rec, nil
}

func (app *WebApp) getUptimeRecordByURL(URL string) (*UptimeRecord, error) {
	var rec UptimeRecord
	var group sql.NullString
	var freq sql.NullString
	err := app.pdb.QueryRow(`
		SELECT "id", "name", "url", "group", "frequency"
		FROM api
		WHERE "url" = $1`, URL).Scan(&rec.ID, &rec.Name, &rec.Url, &group, &freq)
	if err != nil {
		return nil, err
	}
	if group.Valid {
		rec.Group = group.String
	}
	if freq.Valid {
		rec.Frequency = freq.String
	}
	return &rec, nil
}

// createUptimeRecord is for simple uptime creation w/ minimal information
func (app *WebApp) createUptimeRecord(URL string, name string, freq string, group string, extras string) (*UptimeRecord, error) {
	_, err := url.Parse(URL)
	if err != nil {
		return nil, err
	}
	var id uuid.UUID
	err = app.pdb.QueryRow(`
		INSERT INTO api("name", "url", "frequency", "group", "extras")
		VALUES($1, $2, $3, $4, $5) RETURNING id
	`, name, URL, freq, group, extras).Scan(&id)
	if err != nil {
		return nil, err
	}
	return app.getUptimeRecordByID(id)
}

// addUptimeRecord is for simple uptime creation w/ minimal information
func (app *WebApp) addUptimeRecord(URL string) (*UptimeRecord, error) {
	u, err := url.Parse(URL)
	if err != nil {
		return nil, err
	}
	var id uuid.UUID
	err = app.pdb.QueryRow(`
		INSERT INTO api("name", "url")
		VALUES($1, $2) RETURNING id
	`, u.Hostname(), URL).Scan(&id)
	if err != nil {
		return nil, err
	}
	return app.getUptimeRecordByID(id)
}

func (app *WebApp) addUptimeStat(record *UptimeRecord, body ReqStatsUpdateData) error {
	var id int64
	var err error
	if body.Coords != [2]float64{0, 0} {
		point := fmt.Sprintf("POINT(%.8f %.8f)", body.Coords[0], body.Coords[1])
		err = app.pdb.QueryRow(`
			INSERT INTO api_stats(
				"api_id", "status_code", "response_time_ms", "size_byte",
				"from", "from_coords"
			) VALUES($1, $2, $3, $4, $5, $6) RETURNING id
		`, record.ID, body.StatusCode, body.ResponseTime, body.Size, body.From, point).Scan(&id)
	} else {
		err = app.pdb.QueryRow(`
			INSERT INTO api_stats(
				"api_id", "status_code", "response_time_ms", "size_byte", "from"
			) VALUES($1, $2, $3, $4, $5) RETURNING id
		`, record.ID, body.StatusCode, body.ResponseTime, body.Size, body.From).Scan(&id)
	}
	if err != nil {
		return err
	}
	return nil
}
