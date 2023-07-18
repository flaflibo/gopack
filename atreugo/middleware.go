package atreugo

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/flaflibo/gopack/utils"
	"github.com/savsgio/atreugo/v11"
)

func GetJwtMiddleware(secret string) func(ctx *atreugo.RequestCtx) error {
	f := func(ctx *atreugo.RequestCtx) error {

		authHeader := string(ctx.Request.Header.Peek("Authorization"))
		authData := strings.Split(authHeader, " ")
		fmt.Printf("authData: %v\n", authData)

		if len(authData) != 2 {
			return fmt.Errorf("auth header not valid")
		}

		claims, err := utils.VerifyJwt(authData[1], secret)
		if err != nil {
			ctx.SetStatusCode(http.StatusForbidden)
			return fmt.Errorf("not authorized")
		}

		ctx.SetUserValue("claims", claims)

		return ctx.Next()
	}

	return f
}
