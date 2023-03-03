package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/routing"
)

type kademlia interface {
	routing.Routing
	GetClosestPeers(ctx context.Context, key string) ([]peer.ID, error)
}

func main() {
	daemon := NewDaemon()

	l, err := net.Listen("tcp", ":3333")
	if err != nil {
		panic(err)
	}

	fmt.Printf("listening on %v\n", l.Addr())

	go func() {
		daemon.MustStart()
		fmt.Println("Daemon is fully started and ready to serve requests.")
	}()

	/*
		1. Is the peer findable in the DHT?
		2. Does the multiaddr work? (what's the error)
		3. Is the CID in the DHT?
		4. Does the peer respond that it has the given data over Bitswap?
	*/
	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		if err := daemon.runCheck(writer, request.RequestURI); err != nil {
			writer.Header().Add("Access-Control-Allow-Origin", "*")
			writer.WriteHeader(http.StatusInternalServerError)
			_, _ = writer.Write([]byte(err.Error()))
			return
		}
	})

	http.HandleFunc("/find", func(writer http.ResponseWriter, request *http.Request) {
		out, err := daemon.runFindContent(request.Context(), request)
		outputJSONOrErr(writer, out, err)
	})

	err = http.Serve(l, nil)
	if err != nil {
		panic(err)
	}
}

func outputJSONOrErr(writer http.ResponseWriter, out interface{}, err error) {
	writer.Header().Add("Access-Control-Allow-Origin", "*")

	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		_, _ = writer.Write([]byte(err.Error()))
		return
	}

	outputJSON, err := json.Marshal(out)

	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		_, _ = writer.Write([]byte(err.Error()))
		return
	}

	writer.Header().Add("Content-Type", "application/json")
	_, err = writer.Write(outputJSON)

	if err != nil {
		fmt.Printf("could not return data over HTTP: %v\n", err.Error())
	}
}
