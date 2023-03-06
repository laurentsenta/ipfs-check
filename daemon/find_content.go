package daemon

import (
	"context"
	"net/http"
	"time"

	"github.com/aschmahmann/ipfs-check/utils"
	"github.com/ipfs/go-cid"
	"github.com/libp2p/go-libp2p/core/peer"
)

type FindContentOutput struct {
	ErrorParseCID      string          `json:"error_parse_cid,omitempty"`
	ErrorFindProviders string          `json:"error_find_providers,omitempty"`
	Providers          []peer.AddrInfo `json:"providers,omitempty"`
}

func (d *daemon) RunFindContent(ctx context.Context, request *http.Request) (FindContentOutput, error) {
	out := FindContentOutput{}

	cidstr := request.URL.Query().Get("cid")
	if cidstr == "" {
		return out, utils.NewHTTPError(http.StatusBadRequest, "missing cid parameter")
	}

	c, err := cid.Decode(cidstr)
	if err != nil {
		out.ErrorParseCID = err.Error()
		return out, nil
	}

	// Without this there are no peer when the rest of the code executes
	// The fullrt implementation provides a Ready() method.
	time.Sleep(5 * time.Second)

	dialCtx, dialCancel := context.WithTimeout(ctx, 5*time.Second)
	defer dialCancel()

	providers, err := d.dht.FindProviders(dialCtx, c)

	if err != nil {
		out.ErrorFindProviders = err.Error()
		return out, nil
	}

	if len(providers) == 0 {
		out.ErrorFindProviders = "no providers found"
	}

	out.Providers = providers

	return out, nil
}
