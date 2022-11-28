package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/KaptainJosh/Project-Cerebro/cmd/cerebro/cli"
	"github.com/alecthomas/kong"
	"os"
	"time"

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
	//ctx := kong.Parse(&CLI)
	//switch ctx.Command() {
	//case "list":
	//	// create collection channel
	//	entriesCh := make(chan *mdns.ServiceEntry, 4)
	//	done := make(chan struct{})
	//	// start collector
	//	go cli.Collect(entriesCh, os.Stdout, !CLI.List.DisableIPv6, done)
	//
	//	// do lookup
	//	if err := lookup(entriesCh); err != nil {
	//		panic(err)
	//	}
	//
	//	<-done
	//default:
	//	panic(ctx.Command())
	//}

	// Setting up window for GUI interface
	myApp := app.New()
	myWindow := myApp.NewWindow("mDNS Browser")

	//log.Println("Hello World")
	numbers := make([][]string, 0) // Will hold all the values for the mDNS entry table
	var timeout time.Duration      // will hold the user's choice for timeout duration
	var disableIPv6 bool           // will hold user's choice for whether to enable showing IPv6 addresses in the table

	// This
	timeoutLabel := widget.NewLabel("Choose how long you would like the browser to run before returning results. Default is 1s:")

	// This creates a timeout select widget that allows the user to select how long they want the timeout duration to be
	timeoutChoice := widget.NewSelect([]string{"1s", "2s", "3s", "4s", "5s"}, func(value string) {
		switch value {
		case "1s":
			timeout = 1000000000
		case "2s":
			timeout = 2000000000
		case "3s":
			timeout = 3000000000
		case "4s":
			timeout = 4000000000
		case "5s":
			timeout = 5000000000
		}

		//log.Println("Select set to", timeout)
	})

	IPv4Check := widget.NewCheck("Include IPv6 addresses. Optional. Warning: May not work on some systems.", func(value bool) {
		disableIPv6 = value
	})

	startButton := widget.NewButton("Start mDNS Browsing", func() {
		//log.Println(timeout)
		numbers = browser(numbers, timeout, disableIPv6)
		list := widget.NewTable(
			func() (int, int) {
				return len(numbers[0]), len(numbers)
			},
			func() fyne.CanvasObject {
				return widget.NewLabel("table content")
			},
			func(i widget.TableCellID, o fyne.CanvasObject) {
				o.(*widget.Label).SetText(numbers[i.Col][i.Row])
			})
		//tableWindow := myApp.NewWindow("mDNS Table")
		list.SetColumnWidth(0, 570)
		list.SetColumnWidth(1, 450)
		list.SetColumnWidth(3, 50)
		//restartButton := widget.NewButton("Start mDNS Browsing", func() {
		browserWindow := myApp.NewWindow("mDNS Browser Results")
		browserContent := container.New(layout.NewMaxLayout(), list)
		browserWindow.SetContent(browserContent)
		browserWindow.Show()
	})

	sampleText := widget.NewLabel("Choose your options below then push the button at the bottom to start browsing")

	content := container.New(layout.NewVBoxLayout(), sampleText, timeoutLabel, timeoutChoice, IPv4Check, startButton)
	//log.Println("Bye World")

	myWindow.SetContent(content)
	myWindow.ShowAndRun()
}

func browser(numbers [][]string, timeout time.Duration, disableIPv6 bool) [][]string {
	//CLI.List.Timeout = timeout
	//log.Println()
	ctx := kong.Parse(&CLI)
	//log.Println(ctx.Command())
	switch ctx.Command() {
	case "list":
		// create collection channel
		entriesCh := make(chan *mdns.ServiceEntry, 4)
		done := make(chan [][]string)
		name := make([]string, 0)
		host := make([]string, 0)
		port := make([]string, 0)
		addrV4 := make([]string, 0)

		// start collector
		go cli.Collect(entriesCh, os.Stdout, !CLI.List.DisableIPv6, done, name, host, port, addrV4, numbers)

		// do lookup
		if err := lookup(entriesCh, timeout, disableIPv6); err != nil {
			panic(err)
		}

		numbers = <-done

		return numbers

	default:
		panic(ctx.Command())
	}
}

func lookup(entriesCh chan *mdns.ServiceEntry, timeout time.Duration, disableIPv6 bool) error {
	defer close(entriesCh)
	// setup up the query
	//log.Println(timeout)
	p := &mdns.QueryParam{
		Service: CLI.List.ServiceType,
		Domain:  CLI.List.Domain,
		Timeout: timeout,
		//Interface:           nil, // all interfaces
		Entries:             entriesCh,
		WantUnicastResponse: false,
		DisableIPv4:         CLI.List.DisableIPv4,
		DisableIPv6:         !disableIPv6,
	}

	// Start the lookup
	return mdns.Query(p)
}
