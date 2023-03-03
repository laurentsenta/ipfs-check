package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os/exec"
	"testing"
	"time"

	"github.com/gavv/httpexpect/v2"
	"github.com/google/uuid"
	"github.com/ipfs/go-cid"
	mh "github.com/multiformats/go-multihash"
)

const myCID = "QmPZ9gcCEpqKTo6aq61g2nXGUhM4iCL3ewB6LDXZCtioEB" // CID from https://docs.ipfs.tech/how-to/command-line-quick-start/#initialize-the-repository
const someLibp2pAddr = "/p2p/12D3KooWNQ4EzGXzP2ben53ZWriXfPZMahfGFt6M3vw95cZhjCEb"

func setup(t *testing.T) func() {
	t.Helper()

	// start the server binary
	cmd := exec.Command("go", "run", "./")
	err := cmd.Start()
	if err != nil {
		t.Fatal(err)
	}

	// wait for the server to start
	time.Sleep(5 * time.Second)

	return func() {
		cmd.Process.Kill()
	}
}

func TestRootAPIReturn200AndSomePayload(t *testing.T) {
	// teardown := setup(t)
	// defer teardown()

	resp, err := http.Get(fmt.Sprintf("http://localhost:3333?cid=%s&multiaddr=%s", url.QueryEscape(myCID), url.QueryEscape(someLibp2pAddr)))
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatal("status code not OK: ", resp.StatusCode)
	}

	// check response, Should be a non empty JSON:
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	sb := string(body)
	if sb == "" {
		t.Fatal("empty response")
	}
}

func TestFindCIDReturns200AndSomeProviders(t *testing.T) {
	// teardown := setup(t)
	// defer teardown()

	e := httpexpect.Default(t, "http://localhost:3333")

	r := e.GET("/find").
		WithQuery("cid", myCID).
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

// https://github.com/ipfs/go-cid#creating-a-cid-from-scratch
func IdentityCID(s string) (cid.Cid, error) {
	pref := cid.Prefix{
		Version:  1,
		Codec:    uint64(mh.IDENTITY),
		MhType:   mh.SHA2_256,
		MhLength: -1,
	}

	c, err := pref.Sum([]byte(s))
	if err != nil {
		return cid.Undef, err
	}

	return c, err
}

func RandomCID() (cid.Cid, error) {
	return IdentityCID(uuid.New().String())
}
