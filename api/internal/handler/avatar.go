package handler

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rasasaufar/finance-app/api/internal/httputil"
)

// AvatarUploadDir returns the absolute path to the avatar upload directory.
// Priority: AVATAR_UPLOAD_DIR env var → <working_dir>/uploads
func AvatarUploadDir() string {
	if dir := strings.TrimSpace(os.Getenv("AVATAR_UPLOAD_DIR")); dir != "" {
		return dir
	}
	// Use current working directory so "go run" and compiled binary both work
	cwd, err := os.Getwd()
	if err != nil {
		return "uploads"
	}
	return filepath.Join(cwd, "uploads")
}

// avatarUploadDir is the unexported alias used within this package.
func avatarUploadDir() string { return AvatarUploadDir() }

// HandleUploadAvatar handles POST /me/avatar
// Accepts multipart/form-data with field "avatar" (image file, max 2 MB).
// Saves the file to the static images folder and returns the public path.
func (h *Handler) HandleUploadAvatar(w http.ResponseWriter, r *http.Request) {
	const maxSize = 2 << 20 // 2 MB

	if err := r.ParseMultipartForm(maxSize); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "file terlalu besar atau format tidak valid")
		return
	}

	file, header, err := r.FormFile("avatar")
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "field 'avatar' tidak ditemukan")
		return
	}
	defer file.Close()

	// Validate MIME type by reading first 512 bytes
	buf := make([]byte, 512)
	n, err := file.Read(buf)
	if err != nil && err != io.EOF {
		httputil.WriteError(w, http.StatusBadRequest, "gagal membaca file")
		return
	}
	mimeType := http.DetectContentType(buf[:n])
	if !strings.HasPrefix(mimeType, "image/") {
		httputil.WriteError(w, http.StatusBadRequest, "file harus berupa gambar")
		return
	}

	// Seek back to start after reading for MIME detection
	if seeker, ok := file.(io.Seeker); ok {
		if _, err := seeker.Seek(0, io.SeekStart); err != nil {
			httputil.WriteInternalServerError(w, err)
			return
		}
	}

	// Determine extension from original filename
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext == "" {
		ext = ".jpg"
	}
	// Only allow safe image extensions
	switch ext {
	case ".jpg", ".jpeg", ".png", ".webp", ".gif":
		// ok
	default:
		ext = ".jpg"
	}

	// Generate unique filename using timestamp
	filename := fmt.Sprintf("avatar_%d%s", time.Now().UnixMilli(), ext)

	uploadDir := avatarUploadDir()
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		httputil.WriteInternalServerError(w, err)
		return
	}

	destPath := filepath.Join(uploadDir, filename)
	dest, err := os.Create(destPath)
	if err != nil {
		httputil.WriteInternalServerError(w, err)
		return
	}
	defer dest.Close()

	if _, err := io.Copy(dest, file); err != nil {
		httputil.WriteInternalServerError(w, err)
		return
	}

	// Return the public URL path (served by SvelteKit static)
	publicPath := "/images/" + filename
	httputil.WriteJSON(w, http.StatusOK, map[string]string{
		"path": publicPath,
	})
}

// HandleDeleteAvatar handles DELETE /me/avatar?path=/images/filename.jpg
// Deletes the avatar file from disk given its public path as a query param.
func (h *Handler) HandleDeleteAvatar(w http.ResponseWriter, r *http.Request) {
	inputPath := strings.TrimSpace(r.URL.Query().Get("path"))
	if inputPath == "" {
		httputil.WriteError(w, http.StatusBadRequest, "parameter 'path' tidak ditemukan")
		return
	}

	// Sanitize: only allow paths under /images/
	clean := filepath.Clean(inputPath)
	if !strings.HasPrefix(clean, "/images/") {
		httputil.WriteError(w, http.StatusBadRequest, "path tidak valid")
		return
	}

	filename := filepath.Base(clean)
	uploadDir := avatarUploadDir()
	target := filepath.Join(uploadDir, filename)

	// Best-effort delete — ignore not-found errors
	if err := os.Remove(target); err != nil && !os.IsNotExist(err) {
		httputil.WriteInternalServerError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
