package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"

	"github.com/aschmahmann/ipfs-check/daemon"
	"github.com/aschmahmann/ipfs-check/utils"
)

func main() {
	daemon := daemon.NewDaemon()

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
		fmt.Println("Received request: ", request.RequestURI)
		if err := daemon.RunCheck(writer, request.RequestURI); err != nil {
			writer.Header().Add("Access-Control-Allow-Origin", "*")
			writer.WriteHeader(http.StatusInternalServerError)
			_, _ = writer.Write([]byte(err.Error()))
			return
		}
	})

	http.HandleFunc("/find", func(writer http.ResponseWriter, request *http.Request) {
		out, err := daemon.RunFindContent(request.Context(), request)
		outputJSONOrErr(writer, out, err)
	})

	http.HandleFunc("/find-peer", func(writer http.ResponseWriter, request *http.Request) {
		out, err := daemon.RunFindPeer(request.Context(), request)
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
		if httpErr, ok := err.(utils.HTTPError); ok {
			writer.WriteHeader(httpErr.Code)
			_, _ = writer.Write([]byte(httpErr.Message))
			return
		}

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
