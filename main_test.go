package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os/exec"
	"testing"
	"time"
)

const myCID = "bafybeiemxf5abjwjbikoz4mc3a3dla6ual3jsgpdr4cjr3oz3evfyavhwq"
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
