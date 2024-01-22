package api

import (
	db "SimpleBank/db/sqlc"
	"SimpleBank/token"
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
)

// binding defines what rules should be applied for value of this field.
// in this instance,
// by required we force our inputs to not be empty and always have a value.

// func (c *Context) JSON(code int, obj any)
// func (c *Context) ShouldBindJSON(obj any) error
type createAccountRequest struct {
	//Owner    string `json:"owner" binding:"required"`
	Currency string `json:"currency" binding:"required,oneof=USD EUR"`
}

func (server *Server) createAccount(ctx *gin.Context) {
	var request createAccountRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, errResponse(err))
		return
	}
	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	arg := db.CreateAccountParams{
		Owner:    authPayload.Username,
		Currency: request.Currency,
		Balance:  0,
	}
	account, err := server.store.CreateAccount(ctx, arg)
	//pq - A pure Go postgres driver for Go's database/sql package
	//In pkg pq
	//type Error struct {
	// 	Severity         string
	// 	Code             ErrorCode
	// 	Message          string
	// 	Detail           string
	// 	Hint             string
	// 	Position         string
	// 	InternalPosition string
	// 	InternalQuery    string
	// 	Where            string
	// 	Schema           string
	// 	Table            string
	// 	Column           string
	// 	DataTypeName     string
	// 	Constraint       string
	// 	File             string
	// 	Line             string
	// 	Routine          string
	// }
	//func (ec ErrorCode) Name() string
	if err != nil {
		if pqerr, ok := err.(*pq.Error); ok {
			switch pqerr.Code.Name() {
			case "foreign_key_violation", "unique_violation":
				ctx.JSON(http.StatusForbidden, errResponse(err))
				return
			}
		}
		ctx.JSON(http.StatusInternalServerError, errResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, account)
}

type getAccountRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

func (server *Server) getAccount(ctx *gin.Context) {
	fmt.Println("Inside getAccount func handler")
	var req getAccountRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errResponse(err))
		return
	}

	account, err := server.store.GetAccount(ctx, req.ID)

	if err != nil {
		if err1 := sql.ErrNoRows; err1 != nil {
			fmt.Println("No Account found for the mentioned id")
			ctx.JSON(http.StatusNotFound, errResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errResponse(err))
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	if authPayload.Username != account.Owner {
		ctx.JSON(http.StatusUnauthorized, errResponse(errors.New("account doesn't belong to the authorized User")))
	}
	ctx.JSON(http.StatusOK, account)

}

type listAccountReq struct {
	PageID   int32 `form:"page_id" binding:"required,min=1"`
	PageSize int32 `form:"page_size" binding:"required,min=5,max=10"`
}

func (server *Server) listAccounts(ctx *gin.Context) {
	fmt.Println("Inside listAccounts func handler")
	var req listAccountReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errResponse(err))
		return
	}
	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)

	arg := db.ListAccountsParams{
		Owner:  authPayload.Username,
		Limit:  req.PageSize,
		Offset: (req.PageID - 1) * req.PageSize,
	}
	account, err := server.store.ListAccounts(ctx, arg)

	if err != nil {
		if err1 := sql.ErrNoRows; err1 != nil {
			fmt.Println("No Account found for the mentioned page id and offset")
			ctx.JSON(http.StatusNotFound, errResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, account)

}
