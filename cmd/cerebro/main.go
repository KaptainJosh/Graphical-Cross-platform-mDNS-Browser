package main

import (
	"github.com/KaptainJosh/Project-Cerebro/cmd/cerebro/cli"
	"os"
	"time"

	"github.com/alecthomas/kong"
	"github.com/hashicorp/mdns"
)

var CLI struct {
	List struct {
		ServiceType string        `default:"_services._dns-sd._udp"`
		Domain      string        `default:"local."`
		Timeout     time.Duration `default:"1s"`
		DisableIPv4 bool          `default:"false" name:"disable-ipv4" negatable:""`
		DisableIPv6 bool          `default:"true"  name:"disable-ipv6" negatable:""`
	} `cmd:"" help:"List services."`
}

func main() {
	ctx := kong.Parse(&CLI)
	switch ctx.Command() {
	case "list":
		// create collection channel
		entriesCh := make(chan *mdns.ServiceEntry, 4)
		done := make(chan struct{})
		// start collector
		go cli.Collect(entriesCh, os.Stdout, !CLI.List.DisableIPv6, done)

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
		Service: CLI.List.ServiceType,
		Domain:  CLI.List.Domain,
		Timeout: CLI.List.Timeout,
		//Interface:           nil, // all interfaces
		Entries:             entriesCh,
		WantUnicastResponse: false,
		DisableIPv4:         CLI.List.DisableIPv4,
		DisableIPv6:         CLI.List.DisableIPv6,
	}

	// Start the lookup
	return mdns.Query(p)
}
