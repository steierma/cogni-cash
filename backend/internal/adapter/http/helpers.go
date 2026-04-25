package http

import (
	"context"
	"encoding/json"
	"mime"
	"net"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

type contextKey string

const userIDKey contextKey = "userID"

func GetUserID(ctx context.Context) uuid.UUID {
	val := ctx.Value(userIDKey)
	if val == nil {
		return uuid.Nil
	}

	if id, ok := val.(uuid.UUID); ok {
		return id
	}

	if idStr, ok := val.(string); ok {
		id, err := uuid.Parse(idStr)
		if err != nil {
			return uuid.Nil
		}
		return id
	}

	return uuid.Nil
}

// mimeToExt returns the canonical file extension (including dot) for a MIME type.
func mimeToExt(mimeType string) string {
	switch mimeType {
	case "application/pdf":
		return ".pdf"
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/gif":
		return ".gif"
	case "image/webp":
		return ".webp"
	case "text/csv":
		return ".csv"
	case "application/vnd.ms-excel":
		return ".xls"
	}
	return ""
}

// isImageMIME returns true when the MIME type is a supported image format.
func isImageMIME(mimeType string) bool {
	switch mimeType {
	case "image/jpeg", "image/jpg", "image/png", "image/gif", "image/webp":
		return true
	}
	return false
}

// extToMIME maps common file extensions to their canonical MIME type so we can
// fill in the content-type when the browser doesn't provide it.
var extToMIME = map[string]string{
	".pdf":  "application/pdf",
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".png":  "image/png",
	".gif":  "image/gif",
	".webp": "image/webp",
}

// resolveMIME returns the canonical, lower-cased MIME type for an uploaded file.
func resolveMIME(contentTypeHeader, filename string) string {
	mt := contentTypeHeader
	if mt == "" || mt == "application/octet-stream" {
		ext := strings.ToLower(filepath.Ext(filename))
		if mapped, ok := extToMIME[ext]; ok {
			mt = mapped
		} else if detected := mime.TypeByExtension(ext); detected != "" {
			mt = detected
		}
	}
	if idx := strings.IndexByte(mt, ';'); idx != -1 {
		mt = strings.TrimSpace(mt[:idx])
	}
	mt = strings.ToLower(mt)
	if mt == "" {
		mt = "application/octet-stream"
	}
	return mt
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func getClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	return ip
}
