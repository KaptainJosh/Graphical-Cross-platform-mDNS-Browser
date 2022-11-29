package main

import (
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/hashicorp/mdns"
)

type config struct {
	timeout    time.Duration
	enableIPv4 bool
	enableIPv6 bool
}

func main() {
	// config defaults
	cfg := config{
		timeout:    time.Second,
		enableIPv4: true,
		enableIPv6: false,
	}

	myApp := app.New()
	myWindow := myApp.NewWindow("mDNS Browser")

	topLabel := widget.NewLabel("Choose your options below then push the button at the bottom to start browsing")

	startButton := widget.NewButton("Start mDNS browsing", startBrowsing(myApp, &cfg))

	// This creates a timeout select widget that allows the user to select how long they want the timeout duration to be
	timeoutLabel := widget.NewLabel("Choose how long you would like the browser to run before returning results.")
	timeoutSlider, timeoutSliderLabel := newDurationSliderWithLabel(time.Second, 5*time.Second, 500*time.Millisecond, &cfg.timeout)

	ipv4Check := widget.NewCheckWithData(
		"Include IPv4 addresses.",
		binding.BindBool(&cfg.enableIPv4),
	)
	ipv6Check := widget.NewCheckWithData(
		"Include IPv6 addresses. Warning: May not work on some systems.",
		binding.BindBool(&cfg.enableIPv6),
	)

	myWindow.SetContent(
		container.New(
			layout.NewVBoxLayout(),
			topLabel,
			timeoutLabel,
			timeoutSlider,
			timeoutSliderLabel,
			ipv4Check,
			ipv6Check,
			startButton,
		),
	)

	myWindow.ShowAndRun()
}

func startBrowsing(myApp fyne.App, cfg *config) func() {
	return func() {
		data := collect(cfg.timeout, cfg.enableIPv4, cfg.enableIPv6)
		tableDimensions := func() (int, int) {
			if len(data) == 0 {
				return 0, 0
			}
			return len(data), len(data[0])
		}
		createCell := func() fyne.CanvasObject {
			return widget.NewLabel(strings.Repeat("*", 40)) // I am not sure what this sizing is accomplishing
		}
		updateCell := func(id widget.TableCellID, object fyne.CanvasObject) {
			object.(*widget.Label).SetText(data[id.Row][id.Col])
		}
		table := widget.NewTable(tableDimensions, createCell, updateCell)

		// not sure what these column widths really should be
		table.SetColumnWidth(0, 600)
		table.SetColumnWidth(1, 300)
		table.SetColumnWidth(2, 100)
		if cfg.enableIPv4 || cfg.enableIPv6 {
			table.SetColumnWidth(3, 250)
		}
		if cfg.enableIPv4 && cfg.enableIPv6 {
			table.SetColumnWidth(4, 250)
		}

		browserWindow := myApp.NewWindow("mDNS Browser Results")
		browserWindow.SetContent(container.NewMax(table)) // I feel like it's not maxing out the browserWindow maybe?
		browserWindow.Show()
	}
}

func collect(timeout time.Duration, ipv4, ipv6 bool) [][]string {
	entriesCh := make(chan *mdns.ServiceEntry)
	done := make(chan struct{})

	data := [][]string{{"Name", "Host", "Port"}}
	if ipv4 {
		data[0] = append(data[0], "AddrV4")
	}
	if ipv6 {
		data[0] = append(data[0], "AddrV6")
	}

	// add entries to data
	go func() {
		for entry := range entriesCh {
			row := []string{entry.Name, entry.Host, strconv.Itoa(entry.Port)}
			if ipv4 {
				row = append(row, entry.AddrV4.String())
			}
			if ipv6 {
				row = append(row, entry.AddrV6.String())
			}
			data = append(data, row)
		}
		done <- struct{}{}
	}()

	if err := lookup(entriesCh, timeout, ipv4, ipv6); err != nil {
		panic(err)
	}

	// wait for all entries to be added
	<-done

	return data
}

func lookup(entriesCh chan *mdns.ServiceEntry, timeout time.Duration, ipv4, ipv6 bool) error {
	defer close(entriesCh)
	// setup up the query
	p := &mdns.QueryParam{
		Service: "_services._dns-sd._udp",
		Domain:  "local.",
		Timeout: timeout,
		//Interface:           nil, // all interfaces
		Entries:             entriesCh,
		WantUnicastResponse: false,
		DisableIPv4:         !ipv4,
		DisableIPv6:         !ipv6,
	}

	// Start the lookup
	return mdns.Query(p)
}

func newDurationSliderWithLabel(min, max, step time.Duration, d *time.Duration) (*widget.Slider, *widget.Label) {
	// string binding for label text
	b := binding.NewString()
	_ = b.Set(d.String())
	l := widget.NewLabelWithData(b)

	// setup float slider
	s := widget.NewSlider(float64(min), float64(max))
	s.Step = float64(step)
	s.SetValue(float64(*d))

	// event to change bound string value
	s.OnChanged = func(f float64) {
		*d = time.Duration(f)
		_ = b.Set(d.String())
	}

	return s, l
}
