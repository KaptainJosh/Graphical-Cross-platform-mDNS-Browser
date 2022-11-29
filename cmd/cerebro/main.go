package main

import (
	"os"
	"time"

	"github.com/alecthomas/kong"
	"github.com/hashicorp/mdns"

	"github.com/KaptainJosh/Project-Cerebro/cmd/cerebro/cli"
)

var args struct {
	List struct {
		ServiceType string        `default:"_services._dns-sd._udp"`
		Domain      string        `default:"local."`
		Timeout     time.Duration `default:"1s"`
		DisableIPv4 bool          `default:"false" name:"disable-ipv4" negatable:""`
		DisableIPv6 bool          `default:"true"  name:"disable-ipv6" negatable:""`
	} `cmd:"" help:"List services."`
}

func main() {
	ctx := kong.Parse(&args)
	switch ctx.Command() {
	case "list":
		// create collection channel
		entriesCh := make(chan *mdns.ServiceEntry, 0)
		done := make(chan struct{})
		// start collector
		go cli.Collect(entriesCh, os.Stdout, !args.List.DisableIPv6, done)

		// do lookup
		if err := lookup(entriesCh); err != nil {
			panic(err)
		}

		<-done
	default:
		panic(ctx.Command())
	}
}

func lookup(entriesCh chan *mdns.ServiceEntry) error {
	defer close(entriesCh)
	// setup up the query
	p := &mdns.QueryParam{
		Service: args.List.ServiceType,
		Domain:  args.List.Domain,
		Timeout: args.List.Timeout,
		//Interface:           nil, // all interfaces
		Entries:             entriesCh,
		WantUnicastResponse: false,
		DisableIPv4:         args.List.DisableIPv4,
		DisableIPv6:         args.List.DisableIPv6,
	}

	// Start the lookup
	return mdns.Query(p)
}
