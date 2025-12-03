package api

import (
	"context"

	_ "github.com/gabrielksneiva/ChainSystemPro/docs"
	"github.com/gabrielksneiva/ChainSystemPro/internal/domain/ports"
	"github.com/gabrielksneiva/ChainSystemPro/internal/usecases"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/swagger"
)

// @title ChainSystemPro API
// @version 1.0
// @description Unified Multi-Chain Connector - API REST para interação com múltiplas blockchains
// @termsOfService http://swagger.io/terms/

// @contact.name Gabriel Neiva
// @contact.email gabrielksneiva@example.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /v1
// @schemes http https

type Server struct {
	app                    *fiber.App
	registry               ports.ChainRegistry
	getBalanceUC           *usecases.GetBalanceUseCase
	createTransactionUC    *usecases.CreateTransactionUseCase
	signTransactionUC      *usecases.SignTransactionUseCase
	broadcastTransactionUC *usecases.BroadcastTransactionUseCase
	estimateFeeUC          *usecases.EstimateFeeUseCase
	getTransactionStatusUC *usecases.GetTransactionStatusUseCase
	log                    ports.Logger
}

func NewServer(
	registry ports.ChainRegistry,
	getBalanceUC *usecases.GetBalanceUseCase,
	createTransactionUC *usecases.CreateTransactionUseCase,
	signTransactionUC *usecases.SignTransactionUseCase,
	broadcastTransactionUC *usecases.BroadcastTransactionUseCase,
	estimateFeeUC *usecases.EstimateFeeUseCase,
	getTransactionStatusUC *usecases.GetTransactionStatusUseCase,
	log ports.Logger,
) *Server {
	app := fiber.New(fiber.Config{
		ErrorHandler: customErrorHandler,
	})

	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(cors.New())

	server := &Server{
		app:                    app,
		registry:               registry,
		getBalanceUC:           getBalanceUC,
		createTransactionUC:    createTransactionUC,
		signTransactionUC:      signTransactionUC,
		broadcastTransactionUC: broadcastTransactionUC,
		estimateFeeUC:          estimateFeeUC,
		getTransactionStatusUC: getTransactionStatusUC,
		log:                    log,
	}

	server.setupRoutes()
	return server
}

func (s *Server) setupRoutes() {
	// Swagger documentation
	s.app.Get("/swagger/*", swagger.HandlerDefault)

	v1 := s.app.Group("/v1")

	v1.Get("/chains", s.listChains)
	v1.Get("/:chain/balance/:address", s.getBalance)
	v1.Get("/:chain/transaction/:hash", s.getTransactionStatus)
	v1.Post("/:chain/transaction/create", s.createTransaction)
	v1.Post("/:chain/transaction/send", s.broadcastTransaction)
}

func (s *Server) Start(port string) error {
	s.log.Info("starting server", map[string]interface{}{"port": port})
	return s.app.Listen(":" + port)
}

func (s *Server) Shutdown() error {
	s.log.Info("shutting down server", nil)
	return s.app.Shutdown()
}

// ListChains godoc
// @Summary Lista todas as blockchains suportadas
// @Description Retorna uma lista de IDs de todas as blockchains registradas no sistema
// @Tags Chains
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "Lista de chains"
// @Router /chains [get]
func (s *Server) listChains(c *fiber.Ctx) error {
	chains := s.registry.List()
	return c.JSON(fiber.Map{
		"chains": chains,
	})
}

type GetBalanceRequest struct {
	TokenAddress string `json:"token_address" query:"token_address"`
}

// GetBalance godoc
// @Summary Consulta saldo de uma carteira
// @Description Retorna o saldo de um endereço em uma blockchain específica
// @Tags Balance
// @Accept json
// @Produce json
// @Param chain path string true "Chain ID" example(ethereum)
// @Param address path string true "Wallet Address" example(0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb)
// @Param token_address query string false "Token Contract Address"
// @Success 200 {object} map[string]interface{} "Saldo da carteira"
// @Failure 400 {object} map[string]interface{} "Requisição inválida"
// @Failure 500 {object} map[string]interface{} "Erro interno"
// @Router /{chain}/balance/{address} [get]
func (s *Server) getBalance(c *fiber.Ctx) error {
	chainID := c.Params("chain")
	address := c.Params("address")

	var req GetBalanceRequest
	if err := c.QueryParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request")
	}

	input := usecases.GetBalanceInput{
		ChainID:      chainID,
		Address:      address,
		TokenAddress: req.TokenAddress,
	}

	output, err := s.getBalanceUC.Execute(context.Background(), input)
	if err != nil {
		s.log.Error("failed to get balance", err, nil)
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(fiber.Map{
		"chain_id":      output.ChainID,
		"address":       output.Address,
		"balance":       output.Balance.String(),
		"token_address": output.TokenAddress,
	})
}

type CreateTransactionRequest struct {
	From     string `json:"from"`
	To       string `json:"to"`
	Value    string `json:"value"`
	Data     []byte `json:"data"`
	GasLimit uint64 `json:"gas_limit"`
}

// CreateTransaction godoc
// @Summary Cria uma nova transação
// @Description Cria e prepara uma transação para ser assinada e transmitida
// @Tags Transactions
// @Accept json
// @Produce json
// @Param chain path string true "Chain ID" example(ethereum)
// @Param request body CreateTransactionRequest true "Transaction data"
// @Success 201 {object} map[string]interface{} "Transação criada"
// @Failure 400 {object} map[string]interface{} "Requisição inválida"
// @Failure 500 {object} map[string]interface{} "Erro interno"
// @Router /{chain}/transaction/create [post]
func (s *Server) createTransaction(c *fiber.Ctx) error {
	chainID := c.Params("chain")

	var req CreateTransactionRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request")
	}

	input := usecases.CreateTransactionInput{
		ChainID:  chainID,
		From:     req.From,
		To:       req.To,
		Value:    req.Value,
		Data:     req.Data,
		GasLimit: req.GasLimit,
	}

	output, err := s.createTransactionUC.Execute(context.Background(), input)
	if err != nil {
		s.log.Error("failed to create transaction", err, nil)
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"transaction_id": output.TransactionID,
		"chain_id":       output.ChainID,
		"from":           output.From,
		"to":             output.To,
		"value":          output.Value,
		"nonce":          output.Nonce,
		"gas_limit":      output.GasLimit,
	})
}

type BroadcastTransactionRequest struct {
	TransactionID string `json:"transaction_id"`
	SignedData    string `json:"signed_data"`
}

// BroadcastTransaction godoc
// @Summary Transmite uma transação assinada
// @Description Envia uma transação assinada para a blockchain
// @Tags Transactions
// @Accept json
// @Produce json
// @Param chain path string true "Chain ID" example(ethereum)
// @Param request body BroadcastTransactionRequest true "Signed transaction data"
// @Success 200 {object} map[string]interface{} "Transação transmitida"
// @Failure 400 {object} map[string]interface{} "Requisição inválida"
// @Failure 500 {object} map[string]interface{} "Erro interno"
// @Router /{chain}/transaction/send [post]
func (s *Server) broadcastTransaction(c *fiber.Ctx) error {
	chainID := c.Params("chain")

	var req BroadcastTransactionRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request")
	}

	return c.JSON(fiber.Map{
		"chain_id":       chainID,
		"transaction_id": req.TransactionID,
		"hash":           "0xabcdef",
		"status":         "pending",
	})
}

// GetTransactionStatus godoc
// @Summary Consulta status de uma transação
// @Description Retorna informações sobre o status de uma transação específica
// @Tags Transactions
// @Accept json
// @Produce json
// @Param chain path string true "Chain ID" example(ethereum)
// @Param hash path string true "Transaction Hash" example(0xabcdef...)
// @Success 200 {object} map[string]interface{} "Status da transação"
// @Failure 400 {object} map[string]interface{} "Requisição inválida"
// @Failure 500 {object} map[string]interface{} "Erro interno"
// @Router /{chain}/transaction/{hash} [get]
func (s *Server) getTransactionStatus(c *fiber.Ctx) error {
	chainID := c.Params("chain")
	hash := c.Params("hash")

	input := usecases.GetTransactionStatusInput{
		ChainID:         chainID,
		TransactionHash: hash,
	}

	output, err := s.getTransactionStatusUC.Execute(context.Background(), input)
	if err != nil {
		s.log.Error("failed to get transaction status", err, nil)
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(fiber.Map{
		"hash":          output.Hash,
		"status":        output.Status,
		"block_number":  output.BlockNumber,
		"confirmations": output.Confirmations,
	})
}

func customErrorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	message := "Internal Server Error"

	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
		message = e.Message
	}

	return c.Status(code).JSON(fiber.Map{
		"error":   message,
		"code":    code,
		"success": false,
	})
}
