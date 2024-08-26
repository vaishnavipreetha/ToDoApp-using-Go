package main

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3" // Import SQLite driver for testing
)

func setupRouter() *gin.Engine {
	r := gin.Default()
	r.GET("/todos", getTodos)
	r.POST("/todos", createTodo)
	r.GET("/todos/:id", getTodo)
	r.PUT("/todos/:id", updateTodo)
	r.DELETE("/todos/:id", deleteTodo)
	return r
}

// Setup the test database and assign it to the global db variable
func setupTestDB() {
	var err error
	db, err = sql.Open("sqlite3", ":memory:")
	if err != nil {
		panic(err)
	}

	// Create the todos table
	_, err = db.Exec("CREATE TABLE todos (id INTEGER PRIMARY KEY, title TEXT, description TEXT, completed BOOLEAN)")
	if err != nil {
		panic(err)
	}
}

func TestCreateTodo(t *testing.T) {
	setupTestDB()
	defer db.Close()

	router := setupRouter()

	todo := `{"title": "Test Todo", "description": "This is a test todo", "completed": false}`
	req, _ := http.NewRequest("POST", "/todos", strings.NewReader(todo))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status code 201, got %v", w.Code)
	}
}

func TestGetTodos(t *testing.T) {
	setupTestDB()
	defer db.Close()

	router := setupRouter()

	// Add a test todo
	_, err := db.Exec("INSERT INTO todos (title, description, completed) VALUES (?, ?, ?)", "Test Todo", "This is a test todo", false)
	if err != nil {
		t.Fatal(err)
	}

	req, _ := http.NewRequest("GET", "/todos", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code 200, got %v", w.Code)
	}
}

func TestGetTodo(t *testing.T) {
	setupTestDB()
	defer db.Close()

	router := setupRouter()

	// Add a test todo
	result, err := db.Exec("INSERT INTO todos (title, description, completed) VALUES (?, ?, ?)", "Test Todo", "This is a test todo", false)
	if err != nil {
		t.Fatal(err)
	}
	id, _ := result.LastInsertId()

	req, _ := http.NewRequest("GET", "/todos/"+strconv.Itoa(int(id)), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code 200, got %v", w.Code)
	}
}

func TestUpdateTodo(t *testing.T) {
	setupTestDB()
	defer db.Close()

	router := setupRouter()

	// Add a test todo
	result, err := db.Exec("INSERT INTO todos (title, description, completed) VALUES (?, ?, ?)", "Test Todo", "This is a test todo", false)
	if err != nil {
		t.Fatal(err)
	}
	id, _ := result.LastInsertId()

	updatedTodo := `{"title": "Updated Todo", "description": "This is an updated todo", "completed": true}`
	req, _ := http.NewRequest("PUT", "/todos/"+strconv.Itoa(int(id)), strings.NewReader(updatedTodo))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code 200, got %v", w.Code)
	}
}

func TestDeleteTodo(t *testing.T) {
	setupTestDB()
	defer db.Close()

	router := setupRouter()

	// Add a test todo
	result, err := db.Exec("INSERT INTO todos (title, description, completed) VALUES (?, ?, ?)", "Test Todo", "This is a test todo", false)
	if err != nil {
		t.Fatal(err)
	}
	id, _ := result.LastInsertId()

	req, _ := http.NewRequest("DELETE", "/todos/"+strconv.Itoa(int(id)), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code 200, got %v", w.Code)
	}
}
