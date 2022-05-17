package web

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

// ReqMachineReadableUpdate displays possible argument for grading result
type ReqMachineReadableUpdate struct {
	Uri         string `json:"uri"`
	PackageID   string `json:"package_id"`
	ResourceID  string `json:"resource_id"`
	Format      string `json:"format"`
	Grade       string `json:"grade"`
	Points      int64  `json:"points"`
	Encoding    string `json:"encoding"`
	Note        string `json:"note"`
	InspectedBy string `json:"inspected_by"`
}

// HandleMachineReadableUpdate save and store all grading result in opendata_* tables
func (app *WebApp) HandleMachineReadableUpdate(w http.ResponseWriter, r *http.Request) {
	// (1) parse body
	var body ReqMachineReadableUpdate
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		MessageJSONResponse(w, http.StatusNotAcceptable, MsgResponse{
			Message: err.Error(),
		})
		return
	}
	// (2) Validate a bit
	if body.Uri == "" {
		MessageJSONResponse(w, http.StatusNotAcceptable, MsgResponse{
			Message: "empty uri",
		})
		return
	}
	if body.InspectedBy == "" {
		MessageJSONResponse(w, http.StatusNotAcceptable, MsgResponse{
			Message: "empty inspector",
		})
		return
	}
	fmt.Printf("%v", body)
	if body.PackageID == "" || body.ResourceID == "" || body.Grade == "" {
		MessageJSONResponse(w, http.StatusNotAcceptable, MsgResponse{
			Message: "incompleted data",
		})
		return
	}
	// (2) get or create opendata_item
	// as far as CKAN is concerned, package_id & resource_id should be keys to store everything
	// URI might be changed, this will try to keep opendata_item updated too
	rec, err := app.getOrCreateOpendataRecord(body.PackageID, body.ResourceID, body.Uri, body.Format)
	if err != nil {
		MessageJSONResponse(w, http.StatusNotAcceptable, MsgResponse{
			Message: err.Error(),
		})
		return
	}
	// (3) add opendata_stats
	stat, err := app.createOpendataStat(rec.ID, &body)
	if err != nil {
		MessageJSONResponse(w, http.StatusNotAcceptable, MsgResponse{
			Message: err.Error(),
		})
		return
	}
	// we don't need to wait for opendata-item update; put in thread
	go app.updateOpendataRecord(rec.ID, stat.Uri, stat.InspectedAt)

	MessageJSONResponse(w, http.StatusCreated, MsgResponse{
		Message: "OK",
	})
}

// OpendataStatRecord stores opendata item data
type OpendataRecord struct {
	ID            int64     `json:"id"`
	Uri           string    `json:"uri"`
	PackageID     string    `json:"package_id"`
	ResourceID    string    `json:"resource_id"`
	Format        string    `json:"format"`
	CreatedAt     time.Time `json:"created_at"`
	LastCheckedAt time.Time `json:"last_checked_at"`
}

func (app *WebApp) getOrCreateOpendataRecord(packageID, resourceID, uri, format string) (*OpendataRecord, error) {
	rec, err := app.getOpendataRecord(packageID, resourceID)
	if err != nil {
		// It's likely that there is no record exists
		log.Printf("Errror: %v", err.Error())
		rec, err = app.createOpendataRecord(packageID, resourceID, uri, format)
		if err != nil {
			return nil, err
		}
	}
	return rec, nil
}

func (app *WebApp) getOpendataRecord(packageID, resourceID string) (*OpendataRecord, error) {
	var rec OpendataRecord
	var format sql.NullString
	var lastCheckedAt sql.NullTime
	err := app.pdb.QueryRow(`
		SELECT "id", "uri", "format", "created_at", "last_checked_at"
		FROM opendata_item
		WHERE resource_id = $1 AND package_id = $2`,
		resourceID, packageID).Scan(&rec.ID, &rec.Uri, &format, &rec.CreatedAt, &lastCheckedAt)
	if err != nil {
		return nil, err
	}
	if format.Valid {
		rec.Format = format.String
	}
	if lastCheckedAt.Valid {
		rec.LastCheckedAt = lastCheckedAt.Time
	}
	return &rec, nil
}

func (app *WebApp) getOpendataRecordByID(ID int64) (*OpendataRecord, error) {
	var rec OpendataRecord
	var format sql.NullString
	err := app.pdb.QueryRow(`
		SELECT "id", "package_id", "resource_id", "uri", "format", "created_at", "last_checked_at"
		FROM opendata_item
		WHERE id = $1`, ID).Scan(&rec.ID, &rec.PackageID, &rec.ResourceID, &rec.Uri, &format, &rec.CreatedAt, &rec.LastCheckedAt)
	if err != nil {
		return nil, err
	}
	if format.Valid {
		rec.Format = format.String
	}
	return &rec, nil
}

// createOpendataRecord is for simple opendata item creation
func (app *WebApp) createOpendataRecord(packageID, resourceID, uri, format string) (*OpendataRecord, error) {
	rec := OpendataRecord{
		PackageID:  packageID,
		ResourceID: resourceID,
		Uri:        uri,
		Format:     format,
	}
	err := app.pdb.QueryRow(`
		INSERT INTO opendata_item("uri", "package_id", "resource_id", "format")
		VALUES($1, $2, $3, $4) RETURNING id, created_at
	`, uri, packageID, resourceID, format).Scan(&rec.ID, &rec.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &rec, nil
}

// updateOpendataRecord is for quick update opendata-item
func (app *WebApp) updateOpendataRecord(itemID int64, Uri string, lastCheckedAt time.Time) error {
	var id int64
	err := app.pdb.QueryRow(`
		UPDATE opendata_item SET uri = $2, last_checked_at = $3
		WHERE id = $1
		RETURNING id
	`, itemID, Uri, lastCheckedAt).Scan(&id)
	if err != nil {
		return err
	}
	return nil
}

// OpendataStatRecord stores necessary info for opendata stats
type OpendataStatRecord struct {
	ID          int64     `json:"id"`
	ItemID      int64     `json:"item_id"`
	InspectedAt time.Time `json:"inspected_at"`
	InspectedBy string    `json:"inspected_by"`
	Grade       string    `json:"grade"`
	Uri         string    `json:"uri"`
	Points      int64     `json:"points"`
}

// createOpendataStat is for creating opendata_stats item
func (app *WebApp) createOpendataStat(itemID int64, body *ReqMachineReadableUpdate) (*OpendataStatRecord, error) {
	rec := OpendataStatRecord{
		ItemID:      itemID,
		InspectedBy: body.InspectedBy,
		Grade:       body.Grade,
		Points:      body.Points,
		Uri:         body.Uri,
	}
	err := app.pdb.QueryRow(`
		INSERT INTO opendata_stats
		("item_id", "inspected_by", "uri", "grade", "points", "encoding", "note", "inspected_at")
		VALUES($1, $2, $3, $4, $5, $6, $7, NOW()) 
		RETURNING id, inspected_at
	`, itemID, body.InspectedBy, body.Uri, body.Grade, body.Points, body.Encoding, body.Note).Scan(&rec.ID, &rec.InspectedAt)
	if err != nil {
		return nil, err
	}
	return &rec, nil
}
