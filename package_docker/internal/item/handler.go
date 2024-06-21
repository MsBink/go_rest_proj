package item

import (
	"context"
	"encoding/json"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"project/internal/handlers"
)

const (
	ItemsUrl = "/items"
	ItemUrl  = "/item/:id"
)

type itemHandler struct {
	storage Storage
}

func NewItemHandler(storage Storage) handlers.Handler {
	return &itemHandler{
		storage: storage,
	}
}

func (h *itemHandler) Register(router *httprouter.Router) {
	router.GET(ItemsUrl, h.GetList)
	router.GET(ItemUrl, h.GetItemByID)
	router.POST(ItemsUrl, h.CreateItem)
	router.PUT(ItemUrl, h.UpdateItem)
	router.DELETE(ItemUrl, h.DeleteItem)
}

func (h *itemHandler) GetList(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	items, err := h.storage.FindAll(context.Background())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

func (h *itemHandler) GetItemByID(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	itemID := params.ByName("id")

	item, err := h.storage.FindOne(context.Background(), itemID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(item)
	if err != nil {
		panic(err)
	}
}

func (h *itemHandler) CreateItem(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var newItem CreateItemDTO
	err := json.NewDecoder(r.Body).Decode(&newItem)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	item := Item{
		Name:   newItem.Name,
		Price:  newItem.Price,
		UserID: newItem.UserID,
	}

	itemID, err := h.storage.Create(context.Background(), item)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"id": itemID})
}

func (h *itemHandler) UpdateItem(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	itemID := params.ByName("id")

	var updatedItem Item
	err := json.NewDecoder(r.Body).Decode(&updatedItem)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	updatedItem.ID = itemID
	err = h.storage.Update(context.Background(), updatedItem)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *itemHandler) DeleteItem(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	itemID := params.ByName("id")
	err := h.storage.Delete(context.Background(), itemID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
