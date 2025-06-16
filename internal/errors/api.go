package apierror

import (
	"errors"
	"log/slog"

	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5"
)

type ApiError struct {
	error
	Code int
}

type Api struct {
	Error string `json:"error"`
}


func WrapErrorApi(err error, code int) ApiError {
	return ApiError{error: err, Code: code}
}

func CheckApiErrorAndSend(err error, c fiber.Ctx) error {
	switch e := err.(type) {
	case ApiError:
		return c.Status(e.Code).JSON(fiber.Map{"error": e.Error()})
	default:
		if errors.Is(err, pgx.ErrNoRows) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "not found"})
		}
		slog.Error("", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}
}