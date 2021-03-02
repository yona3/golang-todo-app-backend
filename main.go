package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	gonanoid "github.com/matoous/go-nanoid/v2"
)

// requirements

// -[x] GET: get todos
// -[x] POST: create new todos
// -[x] PATCH: update isDone state
// -[x] DELETE: delete todo

type Todo struct {
	Title  string `json:"title"`
	ID     string `json:"id"`
	IsDone bool   `json:"isDone"`
	Date   string `json:"date"`
}

type todoHandlers struct {
	sync.Mutex
	store map[string]Todo
}

// Handlers
func (h *todoHandlers) todos(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Headers", "*")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set( "Access-Control-Allow-Methods","GET, POST, PUT, DELETE, OPTIONS" )

	switch r.Method {
	case "GET":
		h.getTodos(w, r)
		return
	case "POST":
		h.postTodo(w, r)
		return
	case "DELETE":
		h.deleteTodo(w, r)
		return
	case "PATCH":
		h.updateTodo(w, r)
		return
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("method not allowed"))
		return
	}
}

// GET
func (h *todoHandlers) getTodos(w http.ResponseWriter, r *http.Request) {
	todos := make([]Todo, len(h.store))

	h.Lock()
	i := 0
	for _, todo := range h.store {
		todos[i] = todo
		i++
	}
	h.Unlock()

	jsonBytes, err := json.Marshal(todos)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}

	w.Header().Add("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)
}

// POST
func (h *todoHandlers) postTodo(w http.ResponseWriter, r *http.Request) {
	bodyBytes, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	ct := r.Header.Get("content-type")
	if ct != "application/json" {
		w.WriteHeader(http.StatusUnsupportedMediaType)
		w.Write([]byte(fmt.Sprintf("need content-type 'application/json', but got '%s'", ct)))
		return
	}

	f, err := os.Create("./data.json")
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}
	defer f.Close()

	var todo Todo
	err = json.Unmarshal(bodyBytes, &todo)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	var id string
	id, err = gonanoid.New()
	if err != nil {
		id = fmt.Sprintf("%d", time.Now().UnixNano())
		fmt.Println(err)
	}
	todo.ID = id
	todo.Date = fmt.Sprintf("%d", time.Now().UnixNano())
	todo.IsDone = false

	h.Lock()
	h.store[todo.ID] = todo
	defer h.Unlock()
	err = json.NewEncoder(f).Encode(h.store)
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}

	var jsonBytes []byte
	jsonBytes, err = json.Marshal(todo)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}

	w.Header().Add("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)
}

// DELETE
func (h *todoHandlers) deleteTodo(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.String(), "/")
	if len(parts) != 3 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	h.Lock()
	todo, ok := h.store[parts[2]]
	h.Unlock()
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	f, err := os.Create("./data.json")
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}
	defer f.Close()

	h.Lock()
	delete(h.store, parts[2])
	h.Unlock()
	err = json.NewEncoder(f).Encode(h.store)
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}

	jsonBytes, err := json.Marshal(todo)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}

	w.Header().Add("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)
}

// PATCH
func (h *todoHandlers) updateTodo(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.String(), "/")
	if len(parts) != 3 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	bodyBytes, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	ct := r.Header.Get("content-type")
	if ct != "application/json" {
		w.WriteHeader(http.StatusUnsupportedMediaType)
		w.Write([]byte(fmt.Sprintf("need content-type 'application/json', but got '%s'", ct)))
		return
	}

	var body Todo
	err = json.Unmarshal(bodyBytes, &body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	
	var todo Todo
	h.Lock()
	todo, ok := h.store[parts[2]]
	h.Unlock()
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	f, err := os.Create("./data.json")
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}
	defer f.Close()

	todo.IsDone = body.IsDone
	if body.Title != "" {
		todo.Title = body.Title
	}
	h.Lock()
	h.store[parts[2]] = todo
	h.Unlock()
	err = json.NewEncoder(f).Encode(h.store)
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}

	jsonBytes, err := json.Marshal(todo)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}


	w.Header().Add("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)
}

// new handler
func newTodoHandlers() *todoHandlers {
		fmt.Println("init data")
		jsonFromFile, err := ioutil.ReadFile("./data.json")
		if err != nil {
			panic(err)
		}
		
		var jsonData map[string]Todo
		err = json.Unmarshal(jsonFromFile, &jsonData)
		if err != nil {
			panic(err)
		}

	return &todoHandlers{
		store: jsonData,
	}
}

func main() {
	todoHandlers := newTodoHandlers()

	http.HandleFunc("/todos", todoHandlers.todos)
	http.HandleFunc("/todos/", todoHandlers.todos)
	port := os.Getenv("PORT")
	if port == "" {
			port = "3000"
	} 
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		panic(err)
	}
}
