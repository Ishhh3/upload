package main

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

func main() {
	var err error
	// Change DSN to match your MySQL credentials
	db, err = sql.Open("mysql", "root:@tcp(127.0.0.1:3306)/uppic")
	if err != nil {
		log.Fatal("Database connection failed:", err)
	}
	defer db.Close()

	// Serve the HTML form
	http.Handle("/", http.FileServer(http.Dir("./static")))

	// Handle image upload
	http.HandleFunc("/upload", uploadHandler)

	// Serve image by ID
	http.HandleFunc("/image", imageHandler)

	log.Println("Server started at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	file, header, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "Failed to read uploaded file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	imageData, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Failed to read file data", http.StatusInternalServerError)
		return
	}

	stmt, err := db.Prepare("INSERT INTO pictures (name, data) VALUES (?, ?)")
	if err != nil {
		fmt.Println("Errro")
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(header.Filename, imageData)
	if err != nil {
		http.Error(w, "Insert failed", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Image %s uploaded successfully!", header.Filename)
}

func imageHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Missing image ID", http.StatusBadRequest)
		return
	}

	var data []byte
	err := db.QueryRow("SELECT data FROM pictures WHERE id = ?", id).Scan(&data)
	if err != nil {
		http.Error(w, "Image not found", http.StatusNotFound)
		return
	}

	contentType := http.DetectContentType(data)
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}
