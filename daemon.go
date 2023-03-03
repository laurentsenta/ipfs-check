package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-ipns"
	logging "github.com/ipfs/go-log"
	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p-kad-dht/fullrt"
	dhtpb "github.com/libp2p/go-libp2p-kad-dht/pb"
	record "github.com/libp2p/go-libp2p-record"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/net/connmgr"
	"github.com/multiformats/go-multiaddr"
)

var log = logging.Logger("ipfs-check-daemon")

type daemon struct {
	h            host.Host
	dht          *fullrt.FullRT
	dhtMessenger *dhtpb.ProtocolMessenger
}

func NewDaemon() *daemon {
	rm, err := NewResourceManager()
	if err != nil {
		panic(err)
	}

	c, err := connmgr.NewConnManager(600, 900, connmgr.WithGracePeriod(time.Second*30))
	if err != nil {
		panic(err)
	}

	h, err := libp2p.New(
		libp2p.ConnectionManager(c),
		libp2p.ConnectionGater(&privateAddrFilterConnectionGater{}),
		libp2p.ResourceManager(rm),
	)
	if err != nil {
		panic(err)
	}

	d, err := fullrt.NewFullRT(h, "/ipfs",
		fullrt.DHTOption(
			dht.BucketSize(20),
			dht.Validator(record.NamespacedValidator{
				"pk":   record.PublicKeyValidator{},
				"ipns": ipns.Validator{},
			}),
			dht.BootstrapPeers(dht.GetDefaultBootstrapPeerAddrInfos()...),
			dht.Mode(dht.ModeClient),
		))

	if err != nil {
		panic(err)
	}

	pm, err := dhtProtocolMessenger("/ipfs/kad/1.0.0", h)
	if err != nil {
		panic(err)
	}

	daemon := &daemon{h: h, dht: d, dhtMessenger: pm}
	return daemon
}

func (d *daemon) MustStart() {
	// Wait for the DHT to be ready
	for !d.dht.Ready() {
		time.Sleep(time.Second * 10)
	}
}

func (d *daemon) runCheck(writer http.ResponseWriter, uristr string) error {
	u, err := url.ParseRequestURI(uristr)
	if err != nil {
		return err
	}

	mastr := u.Query().Get("multiaddr")
	cidstr := u.Query().Get("cid")

	if mastr == "" || cidstr == "" {
		return errors.New("missing argument")
	}

	ai, err := peer.AddrInfoFromString(mastr)
	if err != nil {
		return err
	}

	c, err := cid.Decode(cidstr)
	if err != nil {
		return err
	}

	log.Infoln("Checking", mastr, cidstr)

	ctx := context.Background()
	out := &Output{}

	connectionFailed := false

	log.Infoln("Searching CID in DHT")
	out.CidInDHT = providerRecordInDHT(ctx, d.dht, c, ai.ID)

	log.Infoln("Done searching CID in DHT: %s", out.CidInDHT)

	log.Infoln("Searching peers in DHT")
	addrMap, peerAddrDHTErr := peerAddrsInDHT(ctx, d.dht, d.dhtMessenger, ai.ID)
	out.PeerFoundInDHT = addrMap

	log.Infoln("Done searching peers in DHT: %s", addrMap)

	// If peerID given, but no addresses check the DHT
	if len(ai.Addrs) == 0 {
		if peerAddrDHTErr != nil {
			connectionFailed = true
			out.ConnectionError = peerAddrDHTErr.Error()
		}
		for a := range addrMap {
			ma, err := multiaddr.NewMultiaddr(a)
			if err != nil {
				log.Errorln("error parsing multiaddr %s: %w", a, err)
				continue
			}
			ai.Addrs = append(ai.Addrs, ma)
		}
	}

	testHost, err := libp2p.New(libp2p.ConnectionGater(&privateAddrFilterConnectionGater{}))
	if err != nil {
		return fmt.Errorf("server error: %w", err)
	}
	defer testHost.Close()

	if !connectionFailed {
		// Is the target connectable
		dialCtx, dialCancel := context.WithTimeout(ctx, time.Second*3)
		connErr := testHost.Connect(dialCtx, *ai)
		dialCancel()
		if connErr != nil {
			out.ConnectionError = connErr.Error()
			connectionFailed = true
		}
	}

	if connectionFailed {
		out.DataAvailableOverBitswap.Error = "could not connect to peer"
	} else {
		// If so is the data available over Bitswap
		bsOut := checkBitswapCID(ctx, testHost, c, *ai)
		out.DataAvailableOverBitswap = *bsOut
	}

	outputData, err := json.Marshal(out)
	if err != nil {
		return err
	}

	writer.Header().Add("Access-Control-Allow-Origin", "*")
	_, err = writer.Write(outputData)
	if err != nil {
		fmt.Printf("could not return data over HTTP: %v\n", err.Error())
	}

	return nil
}

type Output struct {
	ConnectionError          string
	PeerFoundInDHT           map[string]int
	CidInDHT                 bool
	DataAvailableOverBitswap BsCheckOutput
}

func peerAddrsInDHT(ctx context.Context, d kademlia, messenger *dhtpb.ProtocolMessenger, p peer.ID) (map[string]int, error) {
	closestPeers, err := d.GetClosestPeers(ctx, string(p))
	if err != nil {
		return nil, err
	}

	wg := sync.WaitGroup{}
	wg.Add(len(closestPeers))

	resCh := make(chan *peer.AddrInfo, len(closestPeers))

	numSuccessfulResponses := execOnMany(ctx, 0.3, time.Second*3, func(ctx context.Context, peerToQuery peer.ID) error {
		endResults, err := messenger.GetClosestPeers(ctx, peerToQuery, p)
		if err == nil {
			for _, r := range endResults {
				if r.ID == p {
					resCh <- r
					return nil
				}
			}
			resCh <- nil
		}
		return err
	}, closestPeers, false)
	close(resCh)

	if numSuccessfulResponses == 0 {
		return nil, fmt.Errorf("host had trouble querying the DHT")
	}

	addrMap := make(map[string]int)
	for r := range resCh {
		if r == nil {
			continue
		}
		for _, addr := range r.Addrs {
			addrMap[addr.String()]++
		}
	}

	return addrMap, nil
}

func providerRecordInDHT(ctx context.Context, d kademlia, c cid.Cid, p peer.ID) bool {
	queryCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	provsCh := d.FindProvidersAsync(queryCtx, c, 0)
	for {
		select {
		case prov, ok := <-provsCh:
			if !ok {
				return false
			}
			if prov.ID == p {
				return true
			}
		case <-ctx.Done():
			return false
		}
	}
}

// Taken from the FullRT DHT client implementation
//
// execOnMany executes the given function on each of the peers, although it may only wait for a certain chunk of peers
// to respond before considering the results "good enough" and returning.
//
// If sloppyExit is true then this function will return without waiting for all of its internal goroutines to close.
// If sloppyExit is true then the passed in function MUST be able to safely complete an arbitrary amount of time after
// execOnMany has returned (e.g. do not write to resources that might get closed or set to nil and therefore result in
// a panic instead of just returning an error).
func execOnMany(ctx context.Context, waitFrac float64, timeoutPerOp time.Duration, fn func(context.Context, peer.ID) error, peers []peer.ID, sloppyExit bool) int {
	if len(peers) == 0 {
		return 0
	}

	// having a buffer that can take all of the elements is basically a hack to allow for sloppy exits that clean up
	// the goroutines after the function is done rather than before
	errCh := make(chan error, len(peers))
	numSuccessfulToWaitFor := int(float64(len(peers)) * waitFrac)

	putctx, cancel := context.WithTimeout(ctx, timeoutPerOp)
	defer cancel()

	for _, p := range peers {
		go func(p peer.ID) {
			errCh <- fn(putctx, p)
		}(p)
	}

	var numDone, numSuccess, successSinceLastTick int
	var ticker *time.Ticker
	var tickChan <-chan time.Time

	for numDone < len(peers) {
		select {
		case err := <-errCh:
			numDone++
			if err == nil {
				numSuccess++
				if numSuccess >= numSuccessfulToWaitFor && ticker == nil {
					// Once there are enough successes, wait a little longer
					ticker = time.NewTicker(time.Millisecond * 500)
					defer ticker.Stop()
					tickChan = ticker.C
					successSinceLastTick = numSuccess
				}
				// This is equivalent to numSuccess * 2 + numFailures >= len(peers) and is a heuristic that seems to be
				// performing reasonably.
				// TODO: Make this metric more configurable
				// TODO: Have better heuristics in this function whether determined from observing static network
				// properties or dynamically calculating them
				if numSuccess+numDone >= len(peers) {
					cancel()
					if sloppyExit {
						return numSuccess
					}
				}
			}
		case <-tickChan:
			if numSuccess > successSinceLastTick {
				// If there were additional successes, then wait another tick
				successSinceLastTick = numSuccess
			} else {
				cancel()
				if sloppyExit {
					return numSuccess
				}
			}
		}
	}
	return numSuccess
}

type FindContentOutput struct {
	ErrorParseCID      string          `json:"error_parse_cid,omitempty"`
	ErrorFindProviders string          `json:"error_find_providers,omitempty"`
	Providers          []peer.AddrInfo `json:"providers,omitempty"`
}

func (d *daemon) runFindContent(ctx context.Context, request *http.Request) (FindContentOutput, error) {
	out := FindContentOutput{}

	cidstr := request.URL.Query().Get("cid")
	if cidstr == "" {
		return out, errors.New("missing argument: cid")
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
