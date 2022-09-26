package main

import (
	"io"
	"os"
	"time"

	"github.com/alecthomas/kong"
	"github.com/hashicorp/mdns"
	"github.com/jedib0t/go-pretty/v6/table"
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
	cli := kong.Parse(&CLI)
	switch cli.Command() {
	case "list":
		// create collection channel
		entriesCh := make(chan *mdns.ServiceEntry, 4)
		done := make(chan struct{})
		// start collector
		go collect(entriesCh, os.Stdout, done)

		// do lookup
		if err := lookup(entriesCh); err != nil {
			panic(err)
		}

		<-done
	default:
		panic(cli.Command())
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

func collect(entriesCh chan *mdns.ServiceEntry, w io.Writer, done chan struct{}) {
	t := table.NewWriter()
	t.SetOutputMirror(w)
	t.AppendHeader(table.Row{
		"Name",
		"Host",
		"AddrV4",
		"AddrV6",
		"Port",
	})
	for entry := range entriesCh {
		t.AppendRow(table.Row{
			entry.Name,
			entry.Host,
			entry.AddrV4,
			entry.AddrV6,
			entry.Port,
		})
	}
	t.Render()

	done <- struct{}{}
}
