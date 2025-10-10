package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/rajnish-012/delivery-management-system/internal/auth"
	"github.com/rajnish-012/delivery-management-system/internal/models"
	"github.com/rajnish-012/delivery-management-system/internal/orders"
)

func RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/health", healthHandler).Methods("GET")
	r.HandleFunc("/register", registerHandler).Methods("POST")
	r.HandleFunc("/login", loginHandler).Methods("POST")

	// protected routes
	api := r.PathPrefix("/api").Subrouter()
	api.Use(auth.AuthMiddleware)
	api.HandleFunc("/orders", createOrderHandler).Methods("POST")
	api.HandleFunc("/orders", listOrdersHandler).Methods("GET")
	api.HandleFunc("/orders/{id}/cancel", cancelOrderHandler).Methods("POST")

	// admin
	api.HandleFunc("/admin/orders", adminListOrdersHandler).Methods("GET")
}

// Simple JSON helpers
func writeJSON(w http.ResponseWriter, v interface{}, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]string{"status": "ok"}, http.StatusOK)
}

type registerReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Role     string `json:"role"` // customer or admin
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	var req registerReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if req.Username == "" || req.Password == "" || req.Role == "" {
		http.Error(w, "missing fields", http.StatusBadRequest)
		return
	}
	u, err := models.CreateUser(r.Context(), req.Username, req.Password, req.Role)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, map[string]interface{}{"id": u.ID, "username": u.Username, "role": u.Role}, http.StatusCreated)
}

type loginReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	var req loginReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	u, err := models.GetUserByUsername(r.Context(), req.Username)
	if err != nil || !u.CheckPassword(req.Password) {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}
	token, err := auth.GenerateToken(u.ID, u.Role)
	if err != nil {
		http.Error(w, "could not generate token", http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]string{"token": token}, http.StatusOK)
}

func getClaims(r *http.Request) (*auth.Claims, error) {
	c := r.Context().Value("claims")
	if c == nil {
		return nil, errors.New("no claims")
	}
	claims, ok := c.(*auth.Claims)
	if !ok {
		return nil, errors.New("invalid claims")
	}
	return claims, nil
}

type createOrderReq struct {
	Item string `json:"item"`
}

func createOrderHandler(w http.ResponseWriter, r *http.Request) {
	claims, err := getClaims(r)
	if err != nil {
		http.Error(w, "unauth", http.StatusUnauthorized)
		return
	}
	var req createOrderReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	ord, err := models.CreateOrder(r.Context(), claims.UserID, req.Item)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// start progression in background
	orders.StartProgression(r.Context(), ord.ID)
	writeJSON(w, ord, http.StatusCreated)
}

func listOrdersHandler(w http.ResponseWriter, r *http.Request) {
	claims, err := getClaims(r)
	if err != nil {
		http.Error(w, "unauth", http.StatusUnauthorized)
		return
	}
	if claims.Role == "admin" {
		// admin can view all
		all, err := models.ListAllOrders(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, all, http.StatusOK)
		return
	}
	// customer: only their orders
	list, err := models.ListOrdersByCustomer(r.Context(), claims.UserID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, list, http.StatusOK)
}

func cancelOrderHandler(w http.ResponseWriter, r *http.Request) {
	claims, err := getClaims(r)
	if err != nil {
		http.Error(w, "unauth", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, _ := strconv.Atoi(idStr)
	// customers may cancel only their own orders
	ord, err := models.GetOrderByID(r.Context(), id)
	if err != nil {
		http.Error(w, "order not found", http.StatusNotFound)
		return
	}
	if claims.Role != "admin" && ord.CustomerID != claims.UserID {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	// Cancel in DB
	if err := models.CancelOrder(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// stop progression if running
	orders.CancelProgression(id)
	// publish cancellation
	orders.PublishImmediateUpdate(r.Context(), id, "cancelled")
	writeJSON(w, map[string]string{"status": "cancelled"}, http.StatusOK)
}

func adminListOrdersHandler(w http.ResponseWriter, r *http.Request) {
	// show all orders (admin-only)
	claims, err := getClaims(r)
	if err != nil || claims.Role != "admin" {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	all, err := models.ListAllOrders(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, all, http.StatusOK)
}
