package cli

import (
	"github.com/hashicorp/mdns"
	"github.com/jedib0t/go-pretty/v6/table"
	"io"
	"strconv"
)

func Collect(entriesCh chan *mdns.ServiceEntry, w io.Writer, ipv6 bool, done chan [][]string, name []string, host []string, port []string, addrV4 []string, numbers [][]string) {
	t := table.NewWriter()
	t.SetOutputMirror(w)
	// add header to table

	if ipv6 {
		t.AppendHeader(table.Row{"Name", "Host", "Port", "AddrV4", "AddrV6"})

	} else {
		t.AppendHeader(table.Row{"Name", "Host", "Port", "AddrV4"})

		name = append(name, "Name")

		host = append(host, "Host")

		port = append(port, "Port")

		addrV4 = append(addrV4, "AddrV4")

		//numbers = append(numbers, name, host, port, addrV4)
		//log.Println(numbers)
	}
	// append rows to table
	for entry := range entriesCh {
		if ipv6 {
			t.AppendRow(table.Row{entry.Name, entry.Host, entry.Port, entry.AddrV4, entry.AddrV6})
			continue
		}
		t.AppendRow(table.Row{entry.Name, entry.Host, entry.Port, entry.AddrV4})
		name = append(name, entry.Name)
		host = append(host, entry.Host)
		port = append(port, strconv.Itoa(entry.Port))
		addrV4 = append(addrV4, entry.AddrV4.String())

	}
	t.Render()
	numbers = append(numbers, name, host, port, addrV4)
	//log.Println(numbers)
	done <- numbers
}
