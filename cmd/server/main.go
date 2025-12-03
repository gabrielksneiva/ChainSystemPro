package main

import (
	"log"
	"os"

	apprpc "github.com/gabrielksneiva/ChainSystemPro/pkg/rpc"
	"github.com/gofiber/fiber/v2"
)

func main() {
	app := fiber.New()

	rpcURL := getEnv("BTC_RPC_URL", "http://localhost:8332")
	rpcUser := getEnv("BTC_RPC_USER", "")
	rpcPass := getEnv("BTC_RPC_PASS", "")

	client, err := apprpc.NewClient(rpcURL, rpcUser, rpcPass)
	if err != nil {
		log.Fatalf("failed to init rpc client: %v", err)
	}

	app.Post("/tx/broadcast", func(c *fiber.Ctx) error {
		var body struct {
			Hex string `json:"hex"`
		}
		if err := c.BodyParser(&body); err != nil || body.Hex == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
		}
		txid, err := client.SendRawTransaction(body.Hex)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(fiber.Map{"txid": txid})
	})

	app.Get("/tx/:txid/status", func(c *fiber.Ctx) error {
		txid := c.Params("txid")
		conf, confirmed, err := client.GetTransactionStatus(txid)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(fiber.Map{"txid": txid, "confirmations": conf, "confirmed": confirmed})
	})

	port := getEnv("PORT", "8080")
	if err := app.Listen(":" + port); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
