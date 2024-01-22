package api

import (
	db "SimpleBank/db/sqlc"
	"SimpleBank/db/util"
	"SimpleBank/token"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

type Server struct {
	config util.Config
	store  db.Store
	router *gin.Engine
	maker  token.Maker
}

func NewServer(store db.Store, config util.Config) (*Server, error) {
	tokenMaker, err := token.NewPasetoMaker(config.TokenSymmetricKey)
	if err != nil {
		return nil, err
		//panic("The Asymmetric is Invalid")
	}
	server := &Server{store: store,
		maker:  tokenMaker,
		config: config,
	}
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("currency", validCurrency) //first param is the tag, second is the function
	}
	server.SetupRouter()
	return server, nil
}
func (server *Server) SetupRouter() {
	router := gin.Default()
	router.POST("/users", server.createUser)
	router.POST("/users/login", server.loginUser)

	authRoutes := router.Group("/").Use(authMiddleWare(server.maker))
	authRoutes.POST("/accounts", server.createAccount)
	authRoutes.GET("/accounts/:id", server.getAccount)
	authRoutes.GET("/accounts", server.listAccounts) //for list acccounts
	authRoutes.POST("/transfers", server.createTransfer)

	server.router = router
}
func (server *Server) Start(address string) error {
	return server.router.Run(address)
}

func errResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
