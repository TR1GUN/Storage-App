package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

// Интерфейс зависимости (например, репозиторий)
type UserRepository interface {
	FindByID(id int) (*User, error)
}

// Реализация репозитория
type postgresUserRepo struct {
	db *sql.DB
}

func NewPostgresUserRepo(db *sql.DB) *postgresUserRepo {
	return &postgresUserRepo{db: db}
}

// Сервис, который использует репозиторий
type UserService struct {
	repo UserRepository
}

func NewUserService(repo UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) GetUser(id int) (*User, error) {
	return s.repo.FindByID(id)
}

// Обработчик, который зависит от сервиса
type UserHandler struct {
	svc *UserService
}

func NewUserHandler(svc *UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

// Реализация интерфейса http.Handler
func (h *UserHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	user, err := h.svc.GetUser(1)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintln(w, user)
}

//---------------- Сервис ---------------------------

// GetAllRecordsHandler Получаем все возможные записи из нашего хранилища.
func GetAllRecordsHandler(storage *StorageManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		//logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
		records, err := storage.getRecords()
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to get records", err.Error())
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(records); err != nil {
			logger.Error("Error encoding response: %v", err)
		}
	}
}

// GetRecordByIDHandler Получаем запись из хранилища по ее ID
func GetRecordByIDHandler(storage *StorageManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.URL.Path[len("/records/"):] // упрощённо; лучше использовать mux
		id, err := strconv.Atoi(idStr)
		if err != nil {
			writeError(w, http.StatusBadRequest, "Invalid ID", err.Error())
			return
		}

		record, err := storage.getRecordByID(id)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to get record", err.Error())
			return
		}
		if record == nil {
			writeError(w, http.StatusNotFound, "Record not found", "")
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(record); err != nil {
			log.Printf("Error encoding response: %v", err)
		}
	}
}

func makeCreateRecordHandler(storage *StorageManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var record Record
		if err := json.NewDecoder(r.Body).Decode(&record); err != nil {
			writeError(w, http.StatusUnprocessableEntity, "Invalid JSON body", err.Error())
			return
		}

		err := storage.createRecord(record)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to create record", err.Error())
			return
		}

		w.WriteHeader(http.StatusCreated)
	}
}
