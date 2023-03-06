package e2e

import (
	"net/http"
	"testing"

	"github.com/gavv/httpexpect/v2"
)

func TestFindCIDReturns200AndSomeProviders(t *testing.T) {
	e := httpexpect.Default(t, "http://localhost:3333")

	r := e.GET("/find").
		WithQuery("cid", MyCID).
		Expect().
		Status(http.StatusOK)

	r.Header("Access-Control-Allow-Origin").IsEqual("*")

	r.JSON().Object().
		ContainsKey("providers").
		NotContainsKey("error_find_providers").
		NotContainsKey("error_parse_cid")
}

func TestFindRandomCIDReturns200AndAProvidersError(t *testing.T) {
	e := httpexpect.Default(t, "http://localhost:3333")

	c, err := RandomCID()
	if err != nil {
		t.Fatal(err)
	}

	r := e.GET("/find").
		WithQuery("cid", c.String()).
		Expect().
		Status(http.StatusOK)

	r.Header("Access-Control-Allow-Origin").IsEqual("*")

	r.JSON().Object().
		NotContainsKey("providers").
		ContainsKey("error_find_providers").
		NotContainsKey("error_parse_cid")
}

func TestFindInvalidCIDReturns200AndACIDError(t *testing.T) {
	e := httpexpect.Default(t, "http://localhost:3333")

	c, err := RandomCID()
	if err != nil {
		t.Fatal(err)
	}

	r := e.GET("/find").
		WithQuery("cid", c.String() + "woopsie").
		Expect().
		Status(http.StatusOK)

	r.Header("Access-Control-Allow-Origin").IsEqual("*")

	r.JSON().Object().
		NotContainsKey("providers").
		NotContainsKey("error_find_providers").
		ContainsKey("error_parse_cid")
}

func TestFindWithoutCIDReturns400(t *testing.T) {
	e := httpexpect.Default(t, "http://localhost:3333")


	r := e.GET("/find").
		Expect().
		Status(http.StatusBadRequest)

	r.Header("Access-Control-Allow-Origin").IsEqual("*")
}
