package api

import (
	"SimpleBank/token"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	authorizationHeaderKey  = "authorization"         //to get the header mapped to the authorization key from the request
	authorizationTypeBearer = "bearer"                // first field in the authorized token ,eg: bearer v2.ewwdwdwfweewuiwegwiewiew
	authorizationPayloadKey = "authorization_payload" //to set the context with this key and value as payload returned from verify token
)

func authMiddleWare(tokenMaker token.Maker) gin.HandlerFunc {

	return func(ctx *gin.Context) {
		authorizationHeader := ctx.GetHeader(authorizationHeaderKey)
		if len(authorizationHeader) == 0 {
			err := errors.New("authorisation Header not Provided")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errResponse(err))
		}
		fields := strings.Fields(authorizationHeader) //split the string seperated by whitespace into string array
		if len(fields) > 2 || len(fields) < 2 {
			err := errors.New("Invalid Authorization Header Format")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errResponse(err))
			return
		}

		authorizationType := strings.ToLower(fields[0])
		if authorizationType != authorizationTypeBearer {
			err := errors.New("Invalid Authorization Type")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errResponse(err))
			return
		}
		payload, err := tokenMaker.VerifyToken(fields[1])
		if err != nil {
			err := errors.New("Invalid Token")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errResponse(err))
			return
		}
		ctx.Set(authorizationPayloadKey, payload)
		ctx.Next() //Next should be used only inside middleware.
		// It executes the pending handlers in the chain inside the
		//calling handler
	}

}
