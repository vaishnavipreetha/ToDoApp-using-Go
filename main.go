package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

type Todo struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Completed   bool   `json:"completed"`
}

// Read implements io.Reader.
func (t Todo) Read(p []byte) (n int, err error) {
	panic("unimplemented")
}

var db *sql.DB

func initDB() {
	var err error
	connStr := "user=postgres dbname=todoapp sslmode=disable"
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}
}

func main() {
	initDB()
	r := gin.Default()

	// Serve static files from the frontend directory
	r.Static("/frontend", "./frontend")

	r.GET("/todos", getTodos)
	r.POST("/todos", createTodo)
	r.GET("/todos/:id", getTodo)
	r.PUT("/todos/:id", updateTodo)
	r.DELETE("/todos/:id", deleteTodo)

	// Health check endpoint
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	go func() {
		// Wait for the server to be ready
		for {
			resp, err := http.Get("http://localhost:8080/healthz")
			if err == nil && resp.StatusCode == http.StatusOK {
				break
			}
			time.Sleep(100 * time.Millisecond)
		}

		// Measure API execution times
		measureAPITimes()
	}()

	r.Run(":8080")
}

func getTodos(c *gin.Context) {
	rows, err := db.Query("SELECT id, title, description, completed FROM todos")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var todos []Todo
	for rows.Next() {
		var todo Todo
		if err := rows.Scan(&todo.ID, &todo.Title, &todo.Description, &todo.Completed); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		todos = append(todos, todo)
	}

	c.JSON(http.StatusOK, todos)
}

func createTodo(c *gin.Context) {
	var todo Todo
	if err := c.BindJSON(&todo); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var id int
	err := db.QueryRow("INSERT INTO todos (title, description, completed) VALUES ($1, $2, $3) RETURNING id",
		todo.Title, todo.Description, todo.Completed).Scan(&id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	todo.ID = id
	c.JSON(http.StatusCreated, todo)
}

func getTodo(c *gin.Context) {
	id := c.Param("id")

	var todo Todo
	err := db.QueryRow("SELECT id, title, description, completed FROM todos WHERE id = $1", id).
		Scan(&todo.ID, &todo.Title, &todo.Description, &todo.Completed)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Todo not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, todo)
}

func updateTodo(c *gin.Context) {
	id := c.Param("id")
	var todo Todo
	if err := c.BindJSON(&todo); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err := db.Exec("UPDATE todos SET title=$1, description=$2, completed=$3 WHERE id=$4",
		todo.Title, todo.Description, todo.Completed, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Todo updated"})
}

func deleteTodo(c *gin.Context) {
	id := c.Param("id")

	_, err := db.Exec("DELETE FROM todos WHERE id=$1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Todo deleted"})
}

// Measure the execution time of each API endpoint
func measureAPITimes() {
	var wg sync.WaitGroup
	results := make(chan string)

	// Function to measure execution time of an API call
	measure := func(name string, url string) {
		defer wg.Done()
		start := time.Now()
		resp, err := http.Get(url)
		if err != nil {
			results <- name + " failed: " + err.Error()
			return
		}
		defer resp.Body.Close()
		elapsed := time.Since(start)
		results <- name + " took " + elapsed.String()
	}

	measureWithBody := func(name string, url string, method string, body io.Reader) {
		defer wg.Done()
		start := time.Now()
		req, err := http.NewRequest(method, url, body)
		if err != nil {
			results <- name + " failed: " + err.Error()
			return
		}
		req.Header.Set("Content-Type", "application/json") // Set content type for JSON
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			results <- name + " failed: " + err.Error()
			return
		}
		defer resp.Body.Close()
		elapsed := time.Since(start)
		results <- name + " took " + elapsed.String()
	}
	newTodo := Todo{
		Title:       "New Todo",
		Description: "This is a new todo item",
		Completed:   false,
	}

	// Serialize the newTodo to JSON
	body, err := json.Marshal(newTodo)
	if err != nil {
		log.Fatal("Failed to marshal newTodo:", err)
	}

	// Start measuring each API endpoint
	wg.Add(5) // We have 5 endpoints
	go measure("GET /todos", "http://localhost:8080/todos")
	go measureWithBody("POST /todos", "http://localhost:8080/todos", "POST", bytes.NewBuffer(body))   // You may need to modify this to include a body
	go measure("GET /todos/8", "http://localhost:8080/todos/8")                         // Make sure you have a todo with ID 1
	go measureWithBody("PUT /todos/8", "http://localhost:8080/todos/8", "PUT", bytes.NewBuffer(body)) // You may need to modify this to include a body
	go measure("DELETE /todos/8", "http://localhost:8080/todos/8")                      // Make sure you have a todo with ID 1

	// Close the results channel once all goroutines are done
	go func() {
		wg.Wait()
		close(results)
	}()

	// Print results as they come in
	for result := range results {
		log.Println(result)
	}
}
