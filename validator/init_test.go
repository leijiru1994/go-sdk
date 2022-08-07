package validator

import (
	"encoding/json"
	"io"
	"net"
	"net/http"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	ut "github.com/go-playground/universal-translator"
	"github.com/leijiru1994/go-sdk/validator/option"
	"gopkg.in/go-playground/validator.v9"
)

type Account struct {
	Username string `json:"username" binding:"required,oneof=leelei leijiru" trans_display:"用户名(username)"`
	Password string `json:"password" binding:"required,gt=5" trans_display:"密码"`
	Phone    string `json:"phone" binding:"is_china_phone" trans_display:"该手机号"`
}

// go test -v -count=1 init_test.go init.go
func TestTrans(t *testing.T) {
	phoneOpt := option.Option{
		Tag:        "is_china_phone",
		ValidateFn: IsValidChinaPhone,
		RegisterFn: func(ut ut.Translator) error {
			return ut.Add("is_china_phone", "{0}不是合法的中国大陆手机号", false)
		},
		TranslationFn: func(ut ut.Translator, fe validator.FieldError) string {
			t, _ := ut.T("is_china_phone", fe.Field(), fe.Field())
			return t
		},
	}

	err := Init([]option.Option{phoneOpt}...)
	if err != nil {
		t.Error(err)

		return
	}

	engine := gin.New()
	engine.POST("validate", func(ctx *gin.Context) {
		m := map[string]interface{}{
			"code": 0,
		}

		user := &Account{}
		tmpErr := ctx.BindJSON(user)
		if tmpErr != nil {
			m["code"] = http.StatusBadRequest
			m["message"] = ErrorTipAfterTranslate(tmpErr)
			ctx.JSON(http.StatusBadRequest, m)

			return
		}

		ctx.JSON(http.StatusOK, m)
	})

	s := &http.Server{
		Addr:           ":8888",
		Handler:        engine,
		MaxHeaderBytes: 1 << 20,
	}
	var listener net.Listener
	listener, err = net.Listen("tcp", ":8888")
	if err != nil {
		t.Error(err)

		return
	}

	go func() {
		err = s.Serve(listener)
		if err != nil {
			t.Error(err)
		}
	}()

	account2 := `{"username":"leelei001","password":"123"}`
	r2 := strings.NewReader(account2)
	resp, err := http.Post("http://127.0.0.1:8888/validate", "application/json", r2)
	if err != nil {
		t.Error(err)

		return
	}

	if resp.StatusCode != http.StatusBadRequest {
		t.Error("status code should as 400, but now ", resp.StatusCode)
	}

	bs1, _ := io.ReadAll(resp.Body)
	m := map[string]interface{}{}
	err = json.Unmarshal(bs1, &m)
	if err != nil {
		t.Error(err)
	}

	t.Log(string(bs1))
	_ = resp.Body.Close()

	_ = s.Close()

	return
}
