package daemon

import (
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/network"
	rcmgr "github.com/libp2p/go-libp2p/p2p/host/resource-manager"
)

func NewResourceManager() (network.ResourceManager, error) {
	// Copied from:
	// https://github.com/libp2p/go-libp2p/blob/98837aad1591a9c5834fb6589318ee443cd12fe3/p2p/host/resource-manager/README.md

	scalingLimits := rcmgr.DefaultLimits
	libp2p.SetDefaultServiceLimits(&scalingLimits)

	scaledDefaultLimits := scalingLimits.AutoScale()

	cfg := rcmgr.PartialLimitConfig{
		System: rcmgr.ResourceLimits{
			Conns:           rcmgr.Unlimited,
			ConnsInbound:    rcmgr.Unlimited,
			ConnsOutbound:   rcmgr.Unlimited,
			Streams:         rcmgr.Unlimited,
			StreamsInbound:  rcmgr.Unlimited,
			StreamsOutbound: rcmgr.Unlimited,
		},
		Transient: rcmgr.ResourceLimits{
			Conns:           rcmgr.Unlimited,
			ConnsInbound:    rcmgr.Unlimited,
			ConnsOutbound:   rcmgr.Unlimited,
			Streams:         rcmgr.Unlimited,
			StreamsInbound:  rcmgr.Unlimited,
			StreamsOutbound: rcmgr.Unlimited,
		},
		ServiceDefault: rcmgr.ResourceLimits{
			Conns:           rcmgr.Unlimited,
			ConnsInbound:    rcmgr.Unlimited,
			ConnsOutbound:   rcmgr.Unlimited,
			Streams:         rcmgr.Unlimited,
			StreamsInbound:  rcmgr.Unlimited,
			StreamsOutbound: rcmgr.Unlimited,
		},
		ServicePeerDefault: rcmgr.ResourceLimits{
			Conns:           rcmgr.Unlimited,
			ConnsInbound:    rcmgr.Unlimited,
			ConnsOutbound:   rcmgr.Unlimited,
			Streams:         rcmgr.Unlimited,
			StreamsInbound:  rcmgr.Unlimited,
			StreamsOutbound: rcmgr.Unlimited,
		},
		ProtocolDefault: rcmgr.ResourceLimits{
			Conns:           rcmgr.Unlimited,
			ConnsInbound:    rcmgr.Unlimited,
			ConnsOutbound:   rcmgr.Unlimited,
			Streams:         rcmgr.Unlimited,
			StreamsInbound:  rcmgr.Unlimited,
			StreamsOutbound: rcmgr.Unlimited,
		},
		ProtocolPeerDefault: rcmgr.ResourceLimits{
			Conns:           rcmgr.Unlimited,
			ConnsInbound:    rcmgr.Unlimited,
			ConnsOutbound:   rcmgr.Unlimited,
			Streams:         rcmgr.Unlimited,
			StreamsInbound:  rcmgr.Unlimited,
			StreamsOutbound: rcmgr.Unlimited,
		},
		Conn: rcmgr.ResourceLimits{
			Conns:           rcmgr.Unlimited,
			ConnsInbound:    rcmgr.Unlimited,
			ConnsOutbound:   rcmgr.Unlimited,
			Streams:         rcmgr.Unlimited,
			StreamsInbound:  rcmgr.Unlimited,
			StreamsOutbound: rcmgr.Unlimited,
		},
		Stream: rcmgr.ResourceLimits{
			Conns:           rcmgr.Unlimited,
			ConnsInbound:    rcmgr.Unlimited,
			ConnsOutbound:   rcmgr.Unlimited,
			Streams:         rcmgr.Unlimited,
			StreamsInbound:  rcmgr.Unlimited,
			StreamsOutbound: rcmgr.Unlimited,
		},
	}

	limits := cfg.Build(scaledDefaultLimits)
	limiter := rcmgr.NewFixedLimiter(limits)
	rm, err := rcmgr.NewResourceManager(limiter)

	return rm, err
}
