package e2e

import (
	"os/exec"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/ipfs/go-cid"

	mh "github.com/multiformats/go-multihash"
)

const MyCID = "QmPZ9gcCEpqKTo6aq61g2nXGUhM4iCL3ewB6LDXZCtioEB" // CID from https://docs.ipfs.tech/how-to/command-line-quick-start/#initialize-the-repository
const SomeLibp2pAddr = "/p2p/12D3KooWNQ4EzGXzP2ben53ZWriXfPZMahfGFt6M3vw95cZhjCEb"

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
