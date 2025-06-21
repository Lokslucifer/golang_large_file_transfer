package v1

import (
	"large_fss/internals/services"
)

type Handler struct {
	ser        *services.Service
}

func NewHandler(ser *services.Service) *Handler {
	return &Handler{
		ser:ser,
	}
}

