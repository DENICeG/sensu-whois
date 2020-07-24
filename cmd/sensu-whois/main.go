package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"time"

	"github.com/danielb42/whiteflag"
)

var (
	stringToLookFor = "alive"
	timeBegin       = time.Now()
	conn            net.Conn
)

func main() {

	var err error
	log.SetOutput(os.Stderr)

	whiteflag.Alias("s", "server", "sets the whois-server used to perform the check")
	whiteflag.ParseCommandLine()
	whoisServer := whiteflag.GetString("server") + ":43"

	conn, err = net.DialTimeout("tcp", whoisServer, 10*time.Second)
	if err != nil {
		printFailMetricsAndExit("could not connect to", whoisServer, err.Error())
	}
	defer conn.Close()

	_, err = conn.Write([]byte("alive@whois" + "\r\n"))
	if err != nil {
		printFailMetricsAndExit("could not send data to whois:", err.Error())
	}

	timeConnectDone := time.Now()

	buf, err := ioutil.ReadAll(conn)
	if err != nil {
		printFailMetricsAndExit("could not read data from whois:", err.Error())
	}

	durationConnect := timeConnectDone.Sub(timeBegin).Milliseconds()
	durationOrder := time.Now().Sub(timeConnectDone).Milliseconds() // nolint:gosimple
	duration := durationConnect + durationOrder

	if bytes.Contains(buf, []byte(stringToLookFor)) {
		log.Printf("OK: whois replied 'alive'\n\n")
		fmt.Printf("%s %d %d\n", "sensu.whois.available", 1, timeBegin.Unix())
		fmt.Printf("%s %d %d\n", "sensu.whois.duration", duration, timeBegin.Unix())
		fmt.Printf("%s %d %d\n", "sensu.whois.duration.connect", durationConnect, timeBegin.Unix())
		fmt.Printf("%s %d %d\n", "sensu.whois.duration.order", durationOrder, timeBegin.Unix())
	} else {
		printFailMetricsAndExit("whois did not reply 'alive'")
	}
}

func printFailMetricsAndExit(errors ...string) {

	errStr := "ERROR:"

	for _, err := range errors {
		errStr += " " + err
	}

	log.Printf("%s\n\n", errStr)
	fmt.Printf("%s %d %d\n", "sensu.whois.available", 0, timeBegin.Unix())
	fmt.Printf("%s %d %d\n", "sensu.whois.duration", 0, timeBegin.Unix())
	fmt.Printf("%s %d %d\n", "sensu.whois.duration.connect", 0, timeBegin.Unix())
	fmt.Printf("%s %d %d\n", "sensu.whois.duration.order", 0, timeBegin.Unix())

	conn.Close() // nolint:errcheck
	os.Exit(2)
}
