package web

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/google/uuid"
)

type ReqData struct {
	ID           uuid.UUID  `json:"id"`
	Url          string     `json:"url"`
	StatusCode   int64      `json:"status_code"`
	ResponseTime float64    `json:"response_time_ms"`
	Size         float64    `json:"size_byte"`
	From         string     `json:"from"`
	Coords       [2]float64 `json:"from_coords"`
}

func (app *WebApp) HandleUptimeUpdate(w http.ResponseWriter, r *http.Request) {
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
	var body ReqData
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
	ID        uuid.UUID
	Name      string
	Url       string
	Group     string
	Frequency string
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

func (app *WebApp) addUptimeStat(record *UptimeRecord, body ReqData) error {
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
