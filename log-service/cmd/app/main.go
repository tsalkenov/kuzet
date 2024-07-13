package main

import (
	"encoding/json"
	"log"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"google.golang.org/grpc"
	tele "gopkg.in/telebot.v3"

	"log-service/proto/api"
)

const (
	token = "7334864072:AAG35zPObBp-NQFpsj1y151l6XW_fuAhEEg"

	grpcHost = "0.0.0.0:50051"
)

type (
	ErrorResponse struct {
		Error       bool
		FailedField string
		Tag         string
		Value       interface{}
	}

	XValidator struct {
		validator *validator.Validate
	}

	GlobalErrorHandlerResp struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
	}
)

var validate = validator.New()

func (v XValidator) Validate(data interface{}) []ErrorResponse {
	validationErrors := []ErrorResponse{}

	errs := validate.Struct(data)
	if errs != nil {
		for _, err := range errs.(validator.ValidationErrors) {
			var elem ErrorResponse

			elem.FailedField = err.Field()
			elem.Tag = err.Tag()
			elem.Value = err.Value()
			elem.Error = true

			validationErrors = append(validationErrors, elem)
		}
	}

	return validationErrors
}

func main() {
	myValidator := &XValidator{
		validator: validate,
	}

	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(fiber.StatusBadRequest).JSON(GlobalErrorHandlerResp{
				Success: false,
				Message: err.Error(),
			})
		},
	})

	pref := tele.Settings{
		Token: token,
		Poller: &tele.LongPoller{
			Timeout: time.Second * 10,
		},
	}

	b, err := tele.NewBot(pref)
	if err != nil {
		log.Fatalf("failed to create bot: %v", err)
	}

	conn, err := grpc.NewClient(grpcHost, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}
	defer conn.Close()

	client := api.NewPortScannerServiceClient(conn)

	app.Post("/scan", func(c *fiber.Ctx) error {
		var req struct {
			Host      string `json:"host" validate:"required,ipv4"`
			StartPort *int32 `json:"start_port" validate:"omitempty,gte=1,lte=65535"`
			EndPort   *int32 `json:"end_port" validate:"omitempty,gte=1,lte=65535"`
		}

		if err = c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		errs := myValidator.Validate(req)
		if len(errs) > 0 {
			errMsgs := make([]string, 0, len(errs))
			for _, err := range errs {
				errMsgs = append(errMsgs, err.FailedField)
			}

			return &fiber.Error{
				Code:    fiber.ErrBadRequest.Code,
				Message: strings.Join(errMsgs, " and ") + " are invalid",
			}
		}

		if req.EndPort == nil {
			req.EndPort = new(int32)
			*req.EndPort = 32586
		}

		if req.StartPort == nil {
			req.StartPort = new(int32)
			*req.StartPort = 64
		}

		resp, err := client.ScanPorts(c.Context(), &api.ScanPortRequest{
			StartPort: *req.StartPort,
			EndPort:   *req.EndPort,
			Host:      req.Host,
		})
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		bbt, err := json.Marshal(resp)
		if err != nil {
			log.Println(err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{})
		}

		_, err = b.Send(&tele.User{ID: 1385377855}, string(bbt))
		if err != nil {
			log.Println(err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{})
		}

		return c.JSON(resp)
	})

	log.Fatal(app.Listen(":3000"))
}
