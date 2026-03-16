package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
)

type Book struct {
	ID     int    `json:"id"`
	Title  string `json:"title"`
	Author string `json:"author"`
	Year   int    `json:"year"`
}

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

var jwtKey = []byte("your-secret-key")
var booksFile = "../data/books.json"

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/login", loginHandler).Methods("POST")
	r.HandleFunc("/books", jwtMiddleware(listBooksHandler)).Methods("GET")
	r.HandleFunc("/books", jwtMiddleware(createBookHandler)).Methods("POST")
	r.HandleFunc("/books/{id}", jwtMiddleware(updateBookHandler)).Methods("PUT")
	r.HandleFunc("/books/{id}", jwtMiddleware(deleteBookHandler)).Methods("DELETE")

	log.Println("Сервер запущен на :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	var creds Credentials
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Простая проверка (в реальном проекте - из БД)
	if creds.Username != "admin" || creds.Password != "password" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	expirationTime := time.Now().Add(5 * time.Minute)
	claims := &jwt.RegisteredClaims{
		Subject:   creds.Username,
		ExpiresAt: jwt.NewNumericDate(expirationTime),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": tokenString})
}

func jwtMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Missing token", http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		claims := &jwt.RegisteredClaims{}

		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		next(w, r)
	}
}

func listBooksHandler(w http.ResponseWriter, r *http.Request) {
	books := loadBooks()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(books)
}

func createBookHandler(w http.ResponseWriter, r *http.Request) {
	var book Book
	err := json.NewDecoder(r.Body).Decode(&book)
	if err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	books := loadBooks()
	book.ID = len(books) + 1
	books = append(books, book)
	saveBooks(books)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(book)
}

func updateBookHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var updatedBook Book
	err = json.NewDecoder(r.Body).Decode(&updatedBook)
	if err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	books := loadBooks()
	for i, book := range books {
		if book.ID == id {
			updatedBook.ID = id
			books[i] = updatedBook
			saveBooks(books)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(updatedBook)
			return
		}
	}
	http.Error(w, "Book not found", http.StatusNotFound)
}

func deleteBookHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	books := loadBooks()
	for i, book := range books {
		if book.ID == id {
			books = append(books[:i], books[i+1:]...)
			saveBooks(books)
			w.WriteHeader(http.StatusNoContent)
			return
		}
	}
	http.Error(w, "Book not found", http.StatusNotFound)
}

func loadBooks() []Book {
	file, err := os.Open(booksFile)
	if err != nil {
		return []Book{}
	}
	defer file.Close()

	var books []Book
	data, err := ioutil.ReadAll(file)
	if err != nil {
		return []Book{}
	}

	json.Unmarshal(data, &books)
	return books
}

func saveBooks(books []Book) {
	data, err := json.MarshalIndent(books, "", "  ")
	if err != nil {
		log.Println("Error saving books:", err)
		return
	}

	err = ioutil.WriteFile(booksFile, data, 0644)
	if err != nil {
		log.Println("Error writing file:", err)
	}
}
