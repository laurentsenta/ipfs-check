package daemon

import (
	"context"
	"net/http"
	"time"

	"github.com/aschmahmann/ipfs-check/utils"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
)

type FindPeerOutput struct {
	ID                peer.ID               `json:"id,omitempty"`
	ParseAddressError string                `json:"parse_address_error,omitempty"`
	FindPeerError     string                `json:"find_peer_error,omitempty"`
	Addresses         []multiaddr.Multiaddr `json:"addresses,omitempty"`
}

func (d *daemon) RunFindPeer(ctx context.Context, request *http.Request) (FindPeerOutput, error) {
	out := FindPeerOutput{}

	maddr := request.URL.Query().Get("addr")
	if maddr == "" {
		return out, utils.NewHTTPError(http.StatusBadRequest, "missing addr parameter")
	}

	ai, err := peer.AddrInfoFromString(maddr)

	if err != nil {
		out.ParseAddressError = err.Error()
		return out, nil
	}

	out.ID = ai.ID

	dialCtx, dialCancel := context.WithTimeout(ctx, 10*time.Second)
	defer dialCancel()

	addr, err := d.dht.FindPeer(dialCtx, ai.ID)
	if err != nil {
		out.FindPeerError = err.Error()
		return out, nil
	}

	out.Addresses = addr.Addrs

	return out, nil
}