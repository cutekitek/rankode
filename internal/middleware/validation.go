package middleware

import (
	"log/slog"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
)

var  (
	jsonValidator = validator.New()
)

func WrapJson[T any](h func(fiber.Ctx, T) error) fiber.Handler {
	return func (ctx fiber.Ctx) error{
		var v T
		if err := ctx.Bind().JSON(&v);  err != nil {
			return err
		}
		
		if err := jsonValidator.Struct(v); err != nil{
			return err
		} 
		return h(ctx, v)
	}
}

func WrapQuery[T any](h func(fiber.Ctx, T) error) fiber.Handler {
	return func (ctx fiber.Ctx) error  {
		var v T
		if err := ctx.Bind().Query(&v);  err != nil {
			return err
		}

		if err := jsonValidator.Struct(v); err != nil{
			return err
		} 
		return h(ctx, v)
	}
}

func NewJsonValidator[T any]() fiber.Handler {
	return func (ctx fiber.Ctx) error  {
		var v T
		if err := ctx.Bind().JSON(&v);  err != nil {
			return err
		}

		if err := jsonValidator.Struct(v); err != nil{
			return err
		} 
		ctx.Locals("json", v)
		return ctx.Next()
	}
}

func JsonValueFromContext[T any](ctx fiber.Ctx) T {
	v := ctx.Locals("json")
	return v.(T)
}

func NewQueryValidator[T any]() fiber.Handler {
	return func (ctx fiber.Ctx) error  {
		var v T
		if err := ctx.Bind().Query(&v);  err != nil {
			return ctx.SendStatus(fiber.StatusBadRequest)
		}

		if err := jsonValidator.Struct(v); err != nil {
			slog.Error("validation failed", "error", err)
			return ctx.SendStatus(fiber.StatusBadRequest)
		} 
		ctx.Locals("query", v)
		return ctx.Next()
	}
}

func QueryValueFromContext[T any](ctx fiber.Ctx) T {
	v := ctx.Locals("query")
	return v.(T)
}