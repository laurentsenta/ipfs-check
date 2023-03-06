package e2e

import (
	"net/http"
	"testing"

	"github.com/gavv/httpexpect/v2"
)

func TestFindPeerReturns200AndSomeInfo(t *testing.T) {
	e := httpexpect.Default(t, "http://localhost:3333")

	r := e.GET("/find-peer").
		WithQuery("addr", SomeLibp2pAddr).
		Expect().
		Status(http.StatusOK)

	r.Header("Access-Control-Allow-Origin").IsEqual("*")

	r.JSON().Object().
		NotContainsKey("parse_address_error").
		NotContainsKey("find_peer_error").
		ContainsKey("addresses")
}

func TestFindRandomPeerIdReturns200AndAProvidersError(t *testing.T) {
	e := httpexpect.Default(t, "http://localhost:3333")

	x, err := RandomPeerId()
	if err != nil {
		t.Fatal(err)
	}

	r := e.GET("/find-peer").
		WithQuery("addr", "/p2p/" + x.String()).
		Expect().
		Status(http.StatusOK)

	r.Header("Access-Control-Allow-Origin").IsEqual("*")

	r.JSON().Object().
		NotContainsKey("parse_address_error").
		ContainsKey("find_peer_error").
		NotContainsKey("addresses")
}

func TestFindInvalidPeerReturns200AndACIDError(t *testing.T) {
	e := httpexpect.Default(t, "http://localhost:3333")

	x, err := RandomPeerId()
	if err != nil {
		t.Fatal(err)
	}

	r := e.GET("/find-peer").
		WithQuery("addr", "/p2p/" + x.String() + "woopsie").
		Expect().
		Status(http.StatusOK)

	r.Header("Access-Control-Allow-Origin").IsEqual("*")

	r.JSON().Object().
		ContainsKey("parse_address_error").
		NotContainsKey("find_peer_error").
		NotContainsKey("addresses")
}

func TestFindWithoutPeerIdReturns400(t *testing.T) {
	e := httpexpect.Default(t, "http://localhost:3333")

	r := e.GET("/find-peer").
		Expect().
		Status(http.StatusBadRequest)

	r.Header("Access-Control-Allow-Origin").IsEqual("*")
}
