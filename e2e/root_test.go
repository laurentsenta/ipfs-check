package e2e

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"
)


func TestRootAPIReturn200AndSomePayload(t *testing.T) {
	resp, err := http.Get(fmt.Sprintf("http://localhost:3333?cid=%s&multiaddr=%s", url.QueryEscape(MyCID), url.QueryEscape(SomeLibp2pAddr)))
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
