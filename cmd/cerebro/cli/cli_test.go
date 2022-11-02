package cli_test

import (
	"bytes"
	"net"
	"testing"

	"github.com/KaptainJosh/Project-Cerebro/cmd/cerebro/cli"
	"github.com/hashicorp/mdns"
	"github.com/stretchr/testify/require"
)

func TestCollect(t *testing.T) {
	type args struct {
		entries []*mdns.ServiceEntry
		ipv6    bool
		done    chan struct{}
	}
	fixtureEntries := []*mdns.ServiceEntry{
		{
			Name:   "first",
			Host:   "localhost",
			AddrV4: net.IPv4(10, 11, 12, 13),
			AddrV6: net.IPv6zero,
			Port:   8080,
		},
		{
			Name:   "second",
			Host:   "localhost",
			AddrV4: net.IPv4(14, 15, 16, 27),
			AddrV6: net.IPv6linklocalallnodes,
			Port:   8081,
		},
	}
	tests := []struct {
		name  string
		args  args
		wantW string
	}{
		{
			name: "ipv4",
			args: args{
				entries: fixtureEntries,
				ipv6:    false,
				done:    make(chan struct{}),
			},
			wantW: `+--------+-----------+------+-------------+
| NAME   | HOST      | PORT | ADDRV4      |
+--------+-----------+------+-------------+
| first  | localhost | 8080 | 10.11.12.13 |
| second | localhost | 8081 | 14.15.16.27 |
+--------+-----------+------+-------------+
`,
		},
		{
			name: "ipv6",
			args: args{
				entries: fixtureEntries,
				ipv6:    true,
				done:    make(chan struct{}),
			},
			wantW: `+--------+-----------+------+-------------+---------+
| NAME   | HOST      | PORT | ADDRV4      | ADDRV6  |
+--------+-----------+------+-------------+---------+
| first  | localhost | 8080 | 10.11.12.13 | ::      |
| second | localhost | 8081 | 14.15.16.27 | ff02::1 |
+--------+-----------+------+-------------+---------+
`,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			// setup
			entriesCh := make(chan *mdns.ServiceEntry, len(tt.args.entries))
			for i := range tt.args.entries {
				entriesCh <- tt.args.entries[i]
			}
			close(entriesCh)

			// test
			w := &bytes.Buffer{}
			go cli.Collect(entriesCh, w, tt.args.ipv6, tt.args.done)
			done := <-tt.args.done
			require.NotNil(t, done)
			require.Equal(t, w.String(), tt.wantW)
		})
	}
}
