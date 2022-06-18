package common

import (
	"net/http"

	"github.com/leijiru1994/go-sdk/ecode"

	"github.com/gin-gonic/gin"
)

func SendOut(ctx *gin.Context, resp interface{}) {
	SendWithError(ctx, resp, ecode.OK)
}

func SendError(ctx *gin.Context, err ecode.Code) {
	SendWithError(ctx, nil, err)
}

func SendWithError(ctx *gin.Context, resp interface{}, err ecode.Code) {
	if resp == nil {
		resp = struct {}{}
	}

	ctx.Set("code", err.Code())
	obj := gin.H{
		"code": err.Code(),
		"msg":  err.Message(),
		"data": resp,
	}
	ctx.JSON(http.StatusOK, obj)
}
