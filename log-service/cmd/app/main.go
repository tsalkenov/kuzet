package main

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"google.golang.org/grpc"
	tele "gopkg.in/telebot.v3"

	"log-service/proto/api"
)

const (
	token = "7334864072:AAG35zPObBp-NQFpsj1y151l6XW_fuAhEEg"
)

func main() {
	app := fiber.New()

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

	conn, err := grpc.NewClient(":50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}
	defer conn.Close()

	client := api.NewPortScannerServiceClient(conn)

	app.Post("/scan", func(c *fiber.Ctx) error {
		var req struct {
			Host      string `json:"host"`
			StartPort *int32 `json:"start_port"`
			EndPort   *int32 `json:"end_port"`
		}

		if err = c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
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
