package remote

import (
	"fmt"
	"io/ioutil"
	slog "log"
	"net"
	"os"
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/AsynkronIT/protoactor-go/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
)

var (
	s         *grpc.Server
	edpReader *endpointReader
)

// Start the remote server
func Start(address string, options ...RemotingOption) {
	grpclog.SetLogger(slog.New(ioutil.Discard, "", 0))

	host, port, err := net.SplitHostPort(address)
	if err != nil {
		plog.Error("failed to split host port", log.Error(err))
		os.Exit(1)
	}

	var listenAddr string
	if host == "127.0.0.1" {
		listenAddr = address
	} else {
		listenAddr = fmt.Sprintf("0.0.0.0:%s", port)
	}

	lis, err := net.Listen("tcp", listenAddr)
	if err != nil {
		plog.Error("failed to listen", log.Error(err))
		os.Exit(1)
	}
	config := defaultRemoteConfig()
	for _, option := range options {
		option(config)
	}

	// address = lis.Addr().String()
	actor.ProcessRegistry.RegisterAddressResolver(remoteHandler)
	actor.ProcessRegistry.Address = address

	spawnActivatorActor()
	startEndpointManager(config)

	s = grpc.NewServer(config.serverOptions...)
	edpReader = &endpointReader{}
	RegisterRemotingServer(s, edpReader)
	plog.Info("Starting Proto.Actor server", log.String("address", address))
	go s.Serve(lis)
}

func Shutdown(graceful bool) {
	if graceful {
		edpReader.suspend(true)
		stopEndpointManager()
		stopActivatorActor()

		//For some reason GRPC doesn't want to stop
		//Setup timeout as walkaround but need to figure out in the future.
		//TODO: grpc not stopping
		c := make(chan bool, 1)
		go func() {
			s.GracefulStop()
			c <- true
		}()

		select {
		case <-c:
			plog.Info("Stopped Proto.Actor server")
		case <-time.After(time.Second * 10):
			s.Stop()
			plog.Info("Stopped Proto.Actor server", log.String("err", "timeout"))
		}
	} else {
		s.Stop()
		plog.Info("Killed Proto.Actor server")
	}
}
