package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/mux"
)

// Comment структура для комментария
type Comment struct {
	ID        int64     `json:"id"`
	Content   string    `json:"content"`
	ParentID  *int64    `json:"parent_id,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	Children  []Comment `json:"children,omitempty"`
}

// Response структура для ответа с пагинацией
type Response struct {
	Comments []Comment `json:"comments"`
	Page     int       `json:"page"`
	Limit    int       `limit"`
	Total    int       `total"`
	Pages    int       `pages"`
}

// CommentStore хранилище комментариев в памяти
type CommentStore struct {
	sync.RWMutex
	comments map[int64]*Comment
	nextID   int64
}

var store = &CommentStore{
	comments: make(map[int64]*Comment),
	nextID:   1,
}

func main() {
	// Настройка роутера
	r := mux.NewRouter()
	r.HandleFunc("/comments", createComment).Methods("POST")
	r.HandleFunc("/comments", getComments).Methods("GET")
	r.HandleFunc("/comments/{id}", deleteComment).Methods("DELETE")
	// Статические файлы: только для / и /public/*
	r.HandleFunc("/", serveIndex).Methods("GET")
	r.PathPrefix("/public/").Handler(http.StripPrefix("/public/", http.FileServer(http.Dir("public"))))

	// Запуск сервера
	log.Printf("Сервер запущен на http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}

// serveIndex отдаёт index.html для корневого пути
func serveIndex(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "public/index.html")
}

// createComment создаёт новый комментарий
func createComment(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Content  string `json:"content"`
		ParentID *int64 `json:"parent_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Неверный JSON", http.StatusBadRequest)
		return
	}
	if input.Content == "" {
		http.Error(w, "Содержимое обязательно", http.StatusBadRequest)
		return
	}

	store.Lock()
	defer store.Unlock()

	comment := Comment{
		ID:        store.nextID,
		Content:   input.Content,
		ParentID:  input.ParentID,
		CreatedAt: time.Now(),
	}
	store.comments[comment.ID] = &comment
	store.nextID++

	// Если есть parent_id, добавляем в children родителя
	if input.ParentID != nil {
		parent, exists := store.comments[*input.ParentID]
		if !exists {
			http.Error(w, "Родительский комментарий не найден", http.StatusBadRequest)
			return
		}
		parent.Children = append(parent.Children, comment)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(comment)
}

// getCommentTree строит дерево комментариев
func getCommentTree(comments map[int64]*Comment, parentID *int64, search string) []Comment {
	var result []Comment
	for _, comment := range comments {
		if (parentID == nil && comment.ParentID == nil) || (parentID != nil && comment.ParentID != nil && *comment.ParentID == *parentID) {
			if search == "" || strings.Contains(strings.ToLower(comment.Content), strings.ToLower(search)) {
				c := Comment{
					ID:        comment.ID,
					Content:   comment.Content,
					ParentID:  comment.ParentID,
					CreatedAt: comment.CreatedAt,
				}
				c.Children = getCommentTree(comments, &c.ID, search)
				result = append(result, c)
			}
		}
	}
	// Сортировка по created_at
	sort.Slice(result, func(i, j int) bool {
		return result[i].CreatedAt.Before(result[j].CreatedAt)
	})
	return result
}

// getComments получает комментарии с пагинацией и поиском
func getComments(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	parentIDStr := query.Get("parent")
	page, _ := strconv.Atoi(query.Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(query.Get("limit"))
	if limit < 1 {
		limit = 10
	}
	offset := (page - 1) * limit
	search := strings.TrimSpace(query.Get("search"))
	sortOrder := query.Get("sort")
	if sortOrder == "" {
		sortOrder = "asc"
	}

	store.RLock()
	defer store.RUnlock()

	var comments []Comment
	if parentIDStr != "" {
		parentID, err := strconv.ParseInt(parentIDStr, 10, 64)
		if err != nil {
			http.Error(w, "Неверный parent ID", http.StatusBadRequest)
			return
		}
		comments = getCommentTree(store.comments, &parentID, search)
	} else {
		comments = getCommentTree(store.comments, nil, search)
	}

	// Пагинация
	total := len(comments)
	start := offset
	end := offset + limit
	if start > total {
		start = total
	}
	if end > total {
		end = total
	}
	pagedComments := comments[start:end]

	// Сортировка
	if sortOrder == "desc" {
		for i, j := 0, len(pagedComments)-1; i < j; i, j = i+1, j-1 {
			pagedComments[i], pagedComments[j] = pagedComments[j], pagedComments[i]
		}
	}

	resp := Response{
		Comments: pagedComments,
		Page:     page,
		Limit:    limit,
		Total:    total,
		Pages:    (total + limit - 1) / limit,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// deleteComment удаляет комментарий и его поддерево
func deleteComment(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Неверный ID", http.StatusBadRequest)
		return
	}

	store.Lock()
	defer store.Unlock()

	if _, exists := store.comments[id]; !exists {
		http.Error(w, "Комментарий не найден", http.StatusNotFound)
		return
	}

	var deletedIDs []int64
	var collectIDs func(id int64)
	collectIDs = func(id int64) {
		for _, comment := range store.comments {
			if comment.ParentID != nil && *comment.ParentID == id {
				collectIDs(comment.ID)
			}
		}
		deletedIDs = append(deletedIDs, id)
	}
	collectIDs(id)

	count := 0
	for _, delID := range deletedIDs {
		delete(store.comments, delID)
		count++
	}

	for _, comment := range store.comments {
		if comment.Children != nil {
			newChildren := []Comment{}
			for _, child := range comment.Children {
				if !contains(deletedIDs, child.ID) {
					newChildren = append(newChildren, child)
				}
			}
			comment.Children = newChildren
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int{"deleted": count})
}

// contains проверяет, есть ли ID в слайсе
func contains(ids []int64, id int64) bool {
	for _, i := range ids {
		if i == id {
			return true
		}
	}
	return false
}
