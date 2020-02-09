// Inspired by ifstat and netstat utilities
package main

import (
	"context"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	scopenet "github.com/shanebarnes/stethoscope/net"
	"github.com/shanebarnes/stethoscope/internal/version"
	"github.com/urfave/cli/v2"
)

const (
	appName = "stethoscope"
)

func printNif(nif net.Interface) error {
	type addrInfo struct {
		Hw      string `json:"hardware"`
		Mcast []string `json:"multicast"`
		Ucast []string `json:"unicast"`
	}

	type nifInfo struct {
		Addrs     addrInfo `json:"addresses"`
		Flags   []string   `json:"flags"`
		Index     int      `json:"index"`
		Mtu       int      `json:"mtu"`
		Name      string   `json:"name"`
	}

	info := nifInfo{
		Addrs: addrInfo{
			Hw: nif.HardwareAddr.String(),
			Mcast: []string{},
			Ucast: []string{},
		},
		Flags: strings.Split(nif.Flags.String(), "|"),
		Index: nif.Index,
		Mtu:   nif.MTU,
		Name:  nif.Name,
	}

	if addrs, err := nif.Addrs(); err == nil {
		for _, addr := range addrs {
			if ip, _, err := net.ParseCIDR(addr.String()); err == nil {
				info.Addrs.Ucast = append(info.Addrs.Ucast, ip.String())
			}
		}
	}

	if addrs, err := nif.MulticastAddrs(); err == nil {
		for _, addr := range addrs {
			info.Addrs.Mcast = append(info.Addrs.Mcast, addr.String())
		}
	}

	iBuf, err := json.Marshal(&info)
	var pBuf bytes.Buffer
	json.Indent(&pBuf, iBuf, "", "    ")
	fmt.Println(string(pBuf.Bytes()) + ",")
	return err
}

func actionNetList(c *cli.Context) error {
	return scopenet.Walk(func(nif net.Interface, err error) error {
		return printNif(nif)
	}, c.Args().Slice()...)
}

func actionNetStat(c *cli.Context) error {
	var err error
	var wg sync.WaitGroup
	ifNames := c.StringSlice("interface")
	filters := c.StringSlice("filter")

	i := 0
	for _, name := range ifNames {
		_, err := os.Stat(name)
		if os.IsNotExist(err) {
			ifNames[i] = name
			i++
		} else {
			wg.Add(1)
			go func() {
				defer wg.Done()
				var diag scopenet.NifDiag
				diag.Capture(name, filters...)
			}()
		}
	}
	ifNames = ifNames[:i]

	if len(ifNames) > 0 {
		err = scopenet.Watch(context.Background(), 10 * time.Second, func(name string, ev scopenet.WatchEvent, err error) error {
			log.Printf("watch: %s event detected for %s\n", scopenet.WatchEventString(ev), name)
			if ev == scopenet.EventCreate {
				var diag scopenet.NifDiag
				go diag.Capture(name, filters...)
			}

			return nil
		}, ifNames...)
	}

	wg.Wait()
	return err
}

func actionVersion(c *cli.Context) error {
	fmt.Fprintf(os.Stdout, "%s version %s\n", appName, version.String())
	return nil
}

var actions = map[string]cli.ActionFunc {
	"net.list": actionNetList,
	"net.stat": actionNetStat,
	"version":  actionVersion,
}

func main() {
	app := &cli.App{
		Action: mainAction,
		Commands: []*cli.Command {
			{
				Action:  actions["version"],
				Aliases: []string{},
				Name:    "version",
				Usage:   "print version information",
			},
			{
				Aliases:     []string{},
				Name:        "net",
				Usage:       "options for network operations",
				Subcommands: []*cli.Command {
					{
						Action: actions["net.list"],
						Name:   "list",
						Usage:  "print network interface list",
					},
					{
						Action: actions["net.stat"],
						Flags: []cli.Flag {
							&cli.StringSliceFlag {
								Aliases: []string{"f"},
								Name:    "filter",
								Usage:   "capture filter",
							},
							&cli.StringSliceFlag {
								Aliases: []string{"i"},
								Name:    "interface",
								Usage:   "capture interface",
							},
						},
						Name:   "stat",
						Usage:  "print network interface statistics",
					},
				},
			},
		},
		Flags: []cli.Flag {
			&cli.BoolFlag {
				Aliases: []string{"v"},
				Name:    "version",
				Usage:   "print version",
				Value:   false,
			},
		},
		Name:  appName,
		Usage: "listen to your network",
	}

	sort.Sort(cli.FlagsByName(app.Flags))
	sort.Sort(cli.CommandsByName(app.Commands))

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func mainAction(c *cli.Context) error {
	if c.Bool("version") {
		actions["version"](c)
	}

	return nil
}
