package user

import (
	"context"
	"encoding/json"
	"github.com/julienschmidt/httprouter"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
	"project/internal/handlers"
)

const (
	UsersUrl   = "/users"
	UserUrl    = "/user/:uuid"
	AuthUrl    = "/auth"
	SignUrl    = "/signUp"
	RootUrl    = "/"
	staticPath = "/static/"
	staticDir  = http.Dir("static")
	MainUrl    = "/main"
)

type handler struct {
	storage Storage
	Uid     uint32
}

// NewHandler creates a new handler instance.
func NewHandler(storage Storage) handlers.Handler {
	return &handler{
		storage: storage,
	}
}

func (h *handler) Register(router *httprouter.Router) {
	router.GET(RootUrl, func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		http.Redirect(w, r, "/auth", http.StatusSeeOther)
	})
	router.Handler("GET", staticPath+"*filepath", http.StripPrefix(staticPath, http.FileServer(staticDir)))
	router.GET(AuthUrl, ValidateTokenMiddleware(h.LoginUser))
	router.GET(MainUrl, h.MainPage)
	router.POST(AuthUrl, h.LoginUser)
	router.GET(SignUrl, h.RegisterUser)
	router.POST(SignUrl, h.RegisterUser)
	router.GET(UsersUrl, h.GetList)
	router.GET(UserUrl, h.GetUserByUUID)
	router.POST(UsersUrl, ValidateTokenMiddleware(h.CreateUser))
	router.PUT(UserUrl, ValidateTokenMiddleware(h.UpdateUser))
	router.DELETE(UserUrl, ValidateTokenMiddleware(h.DeleteUser))
}

func (h *handler) MainPage(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	http.ServeFile(w, r, "static/main.html")
}
func (h *handler) RegisterUser(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	if r.Method == "POST" {
		var newUser CreateUserDTO
		err := json.NewDecoder(r.Body).Decode(&newUser)
		if err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		// Хэширование пароля
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newUser.Password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, "Error hashing password", http.StatusInternalServerError)
			return
		}

		// Создание пользователя
		user := User{
			Username:     newUser.Username,
			PasswordHash: string(hashedPassword),
			Email:        newUser.Email,
		}

		// Сохранение пользователя в базе данных
		userID, err := h.storage.Register(context.Background(), user)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"id": userID})
		return
	}
	http.ServeFile(w, r, "static/index.html")
}

func (h *handler) LoginUser(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	if r.Method == "POST" {
		var loginUser CreateUserDTO
		err := json.NewDecoder(r.Body).Decode(&loginUser)
		if err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		// Поиск пользователя по имени пользователя
		existingUser, err := h.storage.FindOneByUsername(context.Background(), loginUser.Username)

		if err != nil {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}
		log.Println(loginUser)
		log.Println(existingUser)
		// Проверка пароля
		err = bcrypt.CompareHashAndPassword([]byte(existingUser.PasswordHash), []byte(loginUser.Password))
		if err != nil {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}
		// Генерация токена
		token, err := GenerateToken(existingUser)
		if err != nil {
			http.Error(w, "Error generating token", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		log.Println(token)
		json.NewEncoder(w).Encode(map[string]string{"token": token})
		return
	}
	http.ServeFile(w, r, "static/index.html")
}

//

func (h *handler) GetList(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	users, err := h.storage.FindAll(context.Background())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}
func (h *handler) GetUserByUUID(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	userID := params.ByName("uuid")

	user, err := h.storage.FindOne(context.Background(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(user)
	if err != nil {
		panic(err)
	}
}
func (h *handler) CreateUser(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

	var newUser User
	err := json.NewDecoder(r.Body).Decode(&newUser)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	userID, err := h.storage.Create(context.Background(), newUser)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"id": userID})

}
func (h *handler) UpdateUser(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	userID := params.ByName("uuid")

	var updatedUser User
	err := json.NewDecoder(r.Body).Decode(&updatedUser)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	updatedUser.ID = userID
	err = h.storage.Update(context.Background(), updatedUser)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
func (h *handler) DeleteUser(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	userID := params.ByName("uuid")
	err := h.storage.Delete(context.Background(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
