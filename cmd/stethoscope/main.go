// Inspired by ifstat and netstat utilities

package main

import (
	"context"
	"log"
	"time"

	scopenet "github.com/shanebarnes/stethoscope/net"
)

func main() {
	scopenet.Watch(context.Background(), 10 * time.Second, func(name string, ev scopenet.WatchEvent, err error) error {
		log.Printf("watch: %s event detected for %s\n", scopenet.WatchEventString(ev), name)
		return nil
	})
}
