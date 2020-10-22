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
	fails           int
)

func main() {
	run()
}

func run() {

	var err error
	log.SetOutput(os.Stderr)

	whiteflag.Alias("s", "server", "sets the whois-server used to perform the check")
	whiteflag.ParseCommandLine()
	whoisServer := whiteflag.GetString("server") + ":43"

	conn, err = net.DialTimeout("tcp", whoisServer, 10*time.Second)
	if err != nil {
		printFailMetricsAndExit("could not connect to", whoisServer, err.Error())
	}

	timeConnectDone := time.Now()

	_, err = conn.Write([]byte("alive@whois" + "\r\n"))
	if err != nil {
		printFailMetricsAndExit("could not send data to whois:", err.Error())
	}

	buf, err := ioutil.ReadAll(conn)
	if err != nil {
		printFailMetricsAndExit("could not read data from whois:", err.Error())
	}

	durationConnect := timeConnectDone.Sub(timeBegin).Milliseconds()
	durationOrder := time.Now().Sub(timeConnectDone).Milliseconds() // nolint:gosimple
	durationTotal := durationConnect + durationOrder

	if bytes.Contains(buf, []byte(stringToLookFor)) {
		log.Printf("OK: whois replied 'alive'\n\n")
		fmt.Printf("extmon,service=%s %s=%d,%s=%d,%s=%d,%s=%d %d\n",
			"whois",
			"available", 1,
			"connect", durationConnect,
			"order", durationOrder,
			"total", durationTotal,
			timeBegin.Unix())
	} else {
		printFailMetricsAndExit("whois did not reply 'alive'")
	}

	conn.Close()
	os.Exit(0)
}

func printFailMetricsAndExit(errors ...string) {

	if fails < 3 {
		fails++
		run()
	}

	errStr := "ERROR:"

	for _, err := range errors {
		errStr += " " + err
	}

	log.Printf("%s\n\n", errStr)

	fmt.Printf("extmon,service=%s %s=%d,%s=%d,%s=%d,%s=%d %d\n",
		"whois",
		"available", 0,
		"connect", 0,
		"order", 0,
		"total", 0,
		timeBegin.Unix())

	if conn != nil {
		conn.Close() // nolint:errcheck
	}

	os.Exit(2)
}
