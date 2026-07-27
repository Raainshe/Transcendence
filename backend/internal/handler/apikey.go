package handler

import (
	"backend/internal/service"
)

type APIKeyHandler struct {
	keys *service.APIKeyService
}

func NewAPIKeyHandler(keys *service.APIKeyService) *APIKeyHandler {
	return &APIKeyHandler{keys: keys}
}

func (h *APIKeyHandler) Create() {

}

func (h *APIKeyHandler) List() {

}

func (h *APIKeyHandler) Revoke() {

}
