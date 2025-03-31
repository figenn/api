package llm

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

type Handler struct {
	s *Service
}

func NewHandler(s *Service) *Handler {
	return &Handler{s: s}
}

func (h *Handler) Bind(rg *echo.Group) {
	g := rg.Group("/llm")
	g.POST("/ask", h.Ask)
}

func (h *Handler) Ask(c echo.Context) error {
	var req Prompt
	fmt.Println("SALUT SALUT ")
	if err := c.Bind(&req); err != nil {
		return c.NoContent(http.StatusBadRequest)
	}

	resp, err := h.s.AskAgent(c.Request().Context(), req.Input)
	if err != nil {
		fmt.Println(err)
		return c.NoContent(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, Completion{Output: resp})
}
