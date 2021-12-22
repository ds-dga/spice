package web

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi"
	"github.com/google/uuid"
)

// mediaReponse contains necessary stuffs for the next stuffs
type mediaReponse struct {
	Message    string `json:"message"`
	URL        string `json:"url"`
	ObjectType string `json:"object_type"`
	ObjectID   string `json:"object_id"`
}

func mediaJSONResponse(w http.ResponseWriter, httpCode int, resp mediaReponse) {
	w.WriteHeader(httpCode)
	w.Header().Set("Content-Type", "application/json")
	jMessage, _ := json.Marshal(resp)
	w.Write(jMessage)
}

func getFileFromURL(url string) (*[]byte, string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Sprintf("Bad URL [1]: %v", err), err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Sprintf("Bad URL [2]: %v", err), err
	}
	var buffer []byte

	defer resp.Body.Close()
	buffer, _ = ioutil.ReadAll(resp.Body)
	// There are 2 ways to get content-type: http.Header & buffer itself
	contentType := http.DetectContentType(buffer)
	// contentType := resp.Header.Get("Content-Type")

	// log.Printf("buffer: length = %d\n", len(buffer))
	// log.Printf("buffer: content-type = %s\n", contentType)
	ct := strings.Split(contentType, "/")
	ext := ct[len(ct)-1]
	if ext == "jpeg" { // only extension that has inconsistency name lol
		ext = "jpg"
	}
	extension := fmt.Sprintf(".%s", ext)
	return &buffer, extension, nil
}

// PurgeMediaHandler handles request to remove media
// - auth will be the same as Hasura webhook auth (from headers)
func (app *WebApp) PurgeMediaHandler(w http.ResponseWriter, r *http.Request) {
	/*
		DELETE /media/{mediaID which is UUID}
	*/
	// Find user record
	user, err := app.HeaderAuthenticate(r)
	if err != nil {
		mediaJSONResponse(w, http.StatusForbidden, mediaReponse{
			Message: fmt.Sprintf("No permission: %v", err),
		})
		return
	}
	oUUID, err := uuid.Parse(chi.URLParam(r, "mediaID"))
	if err != nil {
		mediaJSONResponse(w, http.StatusNotAcceptable, mediaReponse{
			Message: "Object ID: Not UUID",
		})
		return
	}
	m, err := app.FindMediaByUUID(user, oUUID)
	if err != nil {
		mediaJSONResponse(w, http.StatusNotAcceptable, mediaReponse{
			Message: fmt.Sprintf("Object not found: %v", err.Error()),
		})
		return
	}
	err = app.DeleteMediaObject(m.Path)
	if err != nil {
		mediaJSONResponse(w, http.StatusInternalServerError, mediaReponse{
			Message: fmt.Sprintf("DELETE ERROR on Media: %v", err.Error()),
		})
		return
	}
	count, err := app.DeleteMediaRecord(m.ID)
	if err != nil {
		mediaJSONResponse(w, http.StatusInternalServerError, mediaReponse{
			Message: fmt.Sprintf("DELETE ERROR on DB: %v", err.Error()),
		})
		return
	}
	mediaJSONResponse(w, http.StatusOK, mediaReponse{
		Message: fmt.Sprintf("Successfully delete media: %d", count),
	})
}

// NewUploadHandler handles request from Hasura webhook auth (from headers)
func (app *WebApp) NewUploadHandler(w http.ResponseWriter, r *http.Request) {
	/* ParseMultipartForm
	https://github.com/golang/go/blob/go1.15.2/src/net/http/request.go#L1277
	body:
	* userID
	* objectID
	* objectType
	* photo				<File> first priority
	* url				<string> either this or photo (lower priority)
	* longitude			{optional}
	* latitude			{optional}
	*/
	// Find user record
	user, err := app.HeaderAuthenticate(r)
	if err != nil {
		mediaJSONResponse(w, http.StatusForbidden, mediaReponse{
			Message: fmt.Sprintf("No permission: %v", err),
		})
		return
	}
	// no need to allocate if it's bigger than 32 MB
	r.Body = http.MaxBytesReader(w, r.Body, 32<<20+512)

	err = r.ParseMultipartForm(32 << 20) // 32Mb
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	url := r.PostForm["url"]
	if r.MultipartForm == nil && len(url) == 0 {
		mediaJSONResponse(w, http.StatusBadRequest, mediaReponse{
			Message: "No multi-part form or url found",
		})
		return
	}
	hasPhoto := r.MultipartForm.File != nil && len(r.MultipartForm.File) > 0
	if !hasPhoto && len(url) == 0 {
		mediaJSONResponse(w, http.StatusBadRequest, mediaReponse{
			Message: "Missing argument: photo/url",
		})
		return
	}

	var objectID []string
	var objectType []string

	if objectID = r.PostForm["objectID"]; len(objectID) == 0 {
		mediaJSONResponse(w, http.StatusBadRequest, mediaReponse{
			Message: "Missing argument: objectID",
		})
		return
	}
	var objID int
	var objUUID uuid.UUID
	objUUID, err = uuid.Parse(objectID[0])
	if err != nil {
		objID, err = strconv.Atoi(objectID[0])
		if err != nil {
			mediaJSONResponse(w, http.StatusBadRequest, mediaReponse{
				Message: "Missing argument: objectID (not uuid nor id <int>)",
			})
			return
		}
	}
	if objectType = r.PostForm["objectType"]; len(objectType) == 0 {
		mediaJSONResponse(w, http.StatusBadRequest, mediaReponse{
			Message: "Missing argument: objectType",
		})
		return
	}

	pnt := [2]float64{0, 0}
	if lon := r.PostForm["longitude"]; len(lon) != 0 {
		pnt[0], err = strconv.ParseFloat(lon[0], 64)
		if err != nil {
			pnt[0] = 0
		}
		pnt[1], err = strconv.ParseFloat(r.PostForm["latitude"][0], 64)
		if err != nil {
			pnt[1] = 0
		}
	}
	var buffer []byte
	var extension string
	var source string
	// Get a file handle, store the file, and so on
	if hasPhoto {
		file, fileHeader, err := r.FormFile("photo")
		if err != nil {
			mediaJSONResponse(w, http.StatusBadRequest, mediaReponse{
				Message: fmt.Sprintf("File error: %v", err),
			})
			return
		}
		defer file.Close()
		buffer = make([]byte, fileHeader.Size)
		file.Read(buffer)
		extension = filepath.Ext(fileHeader.Filename)
	} else { // fetch photo from URL
		buff, extension, err := getFileFromURL(url[0])
		if err != nil {
			mediaJSONResponse(w, http.StatusBadRequest, mediaReponse{
				Message: extension,
			})
			return
		}
		buffer = *buff
		source = url[0]
	}
	now := time.Now()
	randomName := uuid.New()
	// filename includes full path "/{client}.{objectType}/{yyyymm}/{randomName}.{extension}"
	newFileName := fmt.Sprintf("%s.%s/%s/%s%s", user.Client, objectType[0], now.Format("2006-01"), randomName, extension)

	_, err = app.UploadToMedia(buffer, newFileName)
	if err != nil {
		mediaJSONResponse(w, http.StatusBadRequest, mediaReponse{
			Message: fmt.Sprintf("Media error: %v", err),
		})
		return
	}

	record := Media{
		Path:       newFileName,
		ObjectType: objectType[0],
		ObjectUUID: objUUID,
		ObjectID:   int64(objID),
		UserID:     user.ID,
		Point:      pnt,
	}
	if len(source) > 0 {
		record.Extras.Source = source
	}

	// insert to DB
	_, err = app.NewMedia(record)
	if err != nil {
		mediaJSONResponse(w, http.StatusBadGateway, mediaReponse{
			Message: fmt.Sprintf("GW error: %v", err),
		})
		return
	}

	mediaURL := fmt.Sprintf("%s/%s", os.Getenv("Media_ENDPOINT"), newFileName)
	msg := mediaReponse{
		Message:    "success",
		ObjectID:   fmt.Sprintf("%v", record.ObjectID),
		ObjectType: record.ObjectType,
		URL:        mediaURL,
	}
	mediaJSONResponse(w, http.StatusOK, msg)
}
