package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/Ramcache/travel-backend/internal/helpers"
	"github.com/Ramcache/travel-backend/internal/models"
	"github.com/Ramcache/travel-backend/internal/repository"
)

type UserHandler struct {
	repo *repository.UserRepository
	log  *zap.SugaredLogger
}

func NewUserHandler(repo *repository.UserRepository, log *zap.SugaredLogger) *UserHandler {
	return &UserHandler{repo: repo, log: log}
}

// List
// @Summary Get all users
// @Tags users
// @Security Bearer
// @Produce json
// @Success 200 {array} models.User
// @Router /admin/users [get]
func (h *UserHandler) List(w http.ResponseWriter, r *http.Request) {
	users, err := h.repo.GetAll(r.Context())
	if err != nil {
		helpers.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	helpers.JSON(w, http.StatusOK, users)
}

// Get
// @Summary Get user by id
// @Tags users
// @Security Bearer
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} models.User
// @Router /admin/users/{id} [get]
func (h *UserHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	user, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		helpers.Error(w, http.StatusNotFound, "user not found")
		return
	}
	helpers.JSON(w, http.StatusOK, user)
}

// Create
// @Summary Create user
// @Tags users
// @Security Bearer
// @Accept json
// @Produce json
// @Param data body models.CreateUserRequest true "User data"
// @Success 200 {object} models.User
// @Router /admin/users [post]
func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req models.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		helpers.Error(w, http.StatusBadRequest, "invalid request")
		return
	}

	hash, err := helpers.HashPassword(req.Password)
	if err != nil {
		helpers.Error(w, http.StatusInternalServerError, "cannot hash password")
		return
	}

	user := &models.User{
		Email:    req.Email,
		Password: hash,
		FullName: req.FullName,
		RoleID:   req.RoleID,
	}

	if err := h.repo.Create(r.Context(), user); err != nil {
		helpers.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	helpers.JSON(w, http.StatusOK, user)
}

// Update
// @Summary Update user
// @Tags users
// @Security Bearer
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Param data body models.UpdateUserRequest true "User update"
// @Success 200 {object} models.User
// @Router /admin/users/{id} [put]
func (h *UserHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	var req models.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		helpers.Error(w, http.StatusBadRequest, "invalid request")
		return
	}

	user, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		helpers.Error(w, http.StatusNotFound, "user not found")
		return
	}

	if req.FullName != nil {
		user.FullName = *req.FullName
	}
	if req.RoleID != nil {
		user.RoleID = *req.RoleID
	}

	if err := h.repo.Update(r.Context(), user); err != nil {
		helpers.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	helpers.JSON(w, http.StatusOK, user)
}

// Delete
// @Summary Delete user
// @Tags users
// @Security Bearer
// @Param id path int true "User ID"
// @Success 204 "No Content"
// @Router /admin/users/{id} [delete]
func (h *UserHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	if err := h.repo.Delete(r.Context(), id); err != nil {
		helpers.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
