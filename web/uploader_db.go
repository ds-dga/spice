package web

import (
	"encoding/json"
	"log"

	"github.com/google/uuid"
)

// NewMedia to insert new media to gw.everyday.in.th
func (app *WebApp) NewMedia(record Media) (*uuid.UUID, error) {
	var result uuid.UUID
	var err error
	extras, _ := json.Marshal(record.Extras)

	err = app.pdb.QueryRow(`
		INSERT INTO media("name", "size", "uploaded_by", "object_type", "object_id", "object_uuid", "path", "extras")
		VALUES($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id
		`, record.Name, record.Size, record.UserID, record.ObjectType, record.ObjectID, record.ObjectUUID, record.Path, extras).Scan(&result)
	if err != nil {
		log.Printf("[save2psql-create] %v", err)
		return nil, err
	}
	return &result, nil
}

// FindMediaByUUID returns media object from uuid query
func (app *WebApp) FindMediaByUUID(user *User, uuid uuid.UUID) (*Media, error) {
	var m Media
	var buff []byte
	err := app.pdb.QueryRow(`
		SELECT m.uuid, m.path, m.user_id, m.extras
		FROM media m WHERE uuid = $1 and user_id = $2`, uuid, user.ID).Scan(
		&m.ID, &m.Path, &m.UserID, &buff,
	)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

// DeleteMediaRecord delete media record from DB
func (app *WebApp) DeleteMediaRecord(uuid uuid.UUID) (int64, error) {
	query := `DELETE FROM media WHERE uuid = $1`
	res, err := app.pdb.Exec(query, uuid)
	if err != nil {
		return 0, err
	}
	count, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}
	return count, nil
}
