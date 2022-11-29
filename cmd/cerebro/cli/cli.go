package cli

import (
	"io"

	"github.com/hashicorp/mdns"
	"github.com/jedib0t/go-pretty/v6/table"
)

func Collect(entriesCh chan *mdns.ServiceEntry, w io.Writer, ipv6 bool, done chan struct{}) {
	t := table.NewWriter()
	t.SetOutputMirror(w)
	// add header to table
	if ipv6 {
		t.AppendHeader(table.Row{"Name", "Host", "Port", "AddrV4", "AddrV6"})
	} else {
		t.AppendHeader(table.Row{"Name", "Host", "Port", "AddrV4"})
	}
	// append rows to table
	for entry := range entriesCh {
		if ipv6 {
			t.AppendRow(table.Row{entry.Name, entry.Host, entry.Port, entry.AddrV4, entry.AddrV6})
			continue
		}
		t.AppendRow(table.Row{entry.Name, entry.Host, entry.Port, entry.AddrV4})
	}
	t.Render()

	done <- struct{}{}
}
