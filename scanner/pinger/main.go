/*
 * Copyright (c) 2020 Antonios A. Chariton <daknob@daknob.net>
 *
 * Permission to use, copy, modify, and distribute this software for any
 * purpose with or without fee is hereby granted, provided that the above
 * copyright notice and this permission notice appear in all copies.
 *
 * THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
 * WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
 * MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
 * ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
 * WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
 * ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
 * OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.
 *
 */

package main

import (
	"flag"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/sparrc/go-ping"
)

/*
 * Struct that cointains the output file that the CSV is being written to, and
 * in addition to that, a RWMutex, so we can avoid writting to the CSV file in
 * parallel from multiple goroutines, and cause data corruption. By using the
 * appropriate locks and unlocks, only a single line can be written at the same
 * time.
 */
type stdout struct {
	out  *os.File
	lock sync.RWMutex
}

/* Global variable of type 'stdout' that has been defined above */
var out stdout

/*
 * This global variable is a channel that will accept values up to a certain
 * limit, definied by the user. Before each goroutine is spawned, a value will
 * be added. As soon as the goroutine finishes its code, it will remove a value
 * from the channel. Therefore, by setting a certain size to this channel, and
 * by using the blocking nature of channel writes to full channels, we can
 * ensure that only n (where n is defined by the user at runtime) scans can run
 * in parallel.
 */
var guard chan struct{}

func main() {

	ca := flag.Int("a", 147, "The /8  part of the network (first octet)")
	cb := flag.Int("b", 52, "The /16 part of the network (second octet)")
	wr := flag.Int("g", 128, "Number of parallel pings")
	flag.Parse()

	/*
	 * Configure the 'stdout' struct
	 * We create a file, that has a standard name that includes the /16 being
	 * scanned, the year, month, day, hour, and minute, and has the CSV format,
	 * separated by commas. We set the file to the global 'out' file.
	 */
	var err error
	out.out, err = os.Create(
		fmt.Sprintf(
			"pings-%d.%d-%04d-%02d-%02d-%02d-%02d.csv",
			*ca,
			*cb,
			time.Now().Year(),
			time.Now().Month(),
			time.Now().Day(),
			time.Now().Hour(),
			time.Now().Minute(),
		),
	)
	if err != nil {
		logrus.Fatalf("Failed to create file for writting output: %v", err)
	}

	/*
	 * Create the channel guard that will ensure only a user defined amount of
	 * parallel scans will run at the same time and no more.
	 */
	guard = make(chan struct{}, *wr)

	/* Iterate over the /24's and their /32's, and start one goroutine for
	 * every host that has to be scanned.
	 */
	for i := 0; i < 256; i++ {
		for j := 0; j < 256; j++ {
			/*
			 * The following is blocking and therefore ensures the max parallel
			 * goroutines are running and no more.
			 */
			guard <- struct{}{}

			/* Run the ping goroutine */
			go pingAndWrite(fmt.Sprintf("%d.%d.%d.%d", *ca, *cb, i, j))
		}
	}

	/* Starting of goroutines has finished */
	logrus.Infof("Starting of all goroutines finished.")

	/*
	 * Due to their timeout, if we wait 10 seconds, we can be sure that all
	 * will have terminated by then.
	 */
	logrus.Infof("Sleeping for 10 seconds to ensure they have terminated.")
	time.Sleep(10 * time.Second)

	/* Close the output CSV file */
	logrus.Infof("Closing file descriptor...")
	out.out.Close()
	logrus.Infof("Done")

}

/* pingAndWrite is a function that accepts an IPv4 Address addr, as a string,
 * and will then run a ping scan against the desired IP Address, and write the
 * results to the CSV file.
 */
func pingAndWrite(addr string) {
	/* Create a new pinger object */
	p, err := ping.NewPinger(addr)
	if err != nil {
		logrus.Fatalf("Failed to create pinger: %v", err)
	}

	/*
	 * Configure the object
	 * We are sending 4 pings, one every 100 ms, and wait up to one second for
	 * them to return, otherwise they time out. We are using a privileged mode,
	 * which requires either root / Administrator, the proper capabilities, or
	 * the equivalent so it can open the socket required.
	 */
	p.Count = 4
	p.Interval = 100 * time.Millisecond
	p.Timeout = 1 * time.Second
	p.SetPrivileged(true)

	/* Run the ping check */
	p.Run()

	/* Get the check results */
	s := p.Statistics()
	var up string
	if s.PacketsSent == s.PacketsRecv {
		up = "up"
	} else {
		up = "down"
	}

	/* Save the results to file */

	/* Lock the file for writting */
	out.lock.Lock()

	/* Output the CSV, properly formatted for use by imager */
	out.out.WriteString(fmt.Sprintf(
		"%s,%s,%d,%d,%d\n",
		s.Addr,
		up,
		s.PacketsSent,
		s.PacketsRecv,
		s.AvgRtt/time.Millisecond,
	))

	/* Output written IP Address */
	logrus.Printf("Saved: %s", addr)

	/* Unlock CSV output file */
	out.lock.Unlock()

	/* Remove an element from the guard channel */
	<-guard

	/* Finish the goroutine execution */
	return

}
