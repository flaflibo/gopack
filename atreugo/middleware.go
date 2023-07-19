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

		if len(authData) != 2 {
			return fmt.Errorf("auth header not valid")
		}

		claims, err := utils.VerifyJwt(authData[1], secret)
		if err != nil {
			ctx.SetStatusCode(http.StatusForbidden)
			return fmt.Errorf("not authorized")
		}
		ctx.SetUserValue("claims", claims)
		ctx.Response.Header.Set("Server", "mega-cool-server")
		return ctx.Next()
	}

	return f
}

func GetKeepAliveMiddleware(timeout uint32) func(ctx *atreugo.RequestCtx) error {
	return func(ctx *atreugo.RequestCtx) error {
		ctx.Response.Header.Set("Connection", "Keep-Alive")
		ctx.Response.Header.Set("Keep-Alive", fmt.Sprintf("timeout=%d", timeout))
		return ctx.Next()
	}
}
