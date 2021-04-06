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
	timeBegin = time.Now()

	stringToLookFor = "alive"
	conn            net.Conn
	whoisServer     string
	fails           int
)

func main() {
	whiteflag.Alias("s", "server", "sets the whois-server used to perform the check")
	whoisServer = whiteflag.GetString("server") + ":43"

	run()
}

func run() {
	var err error
	log.SetOutput(os.Stderr)
	log.SetPrefix("UTC ")
	log.SetFlags(log.Ltime | log.Lmsgprefix | log.LUTC)

	if conn != nil {
		conn.Close() // nolint:errcheck
	}

	conn, err = net.DialTimeout("tcp", whoisServer, 10*time.Second)
	if err != nil {
		printFailMetricsAndExit("could not connect to", whoisServer, err.Error())
	}

	durationConnect := time.Since(timeBegin).Milliseconds()

	_, err = conn.Write([]byte("alive@whois" + "\r\n"))
	if err != nil {
		printFailMetricsAndExit("could not send data to whois:", err.Error())
	}

	timeSendDone := time.Now()

	buf, err := ioutil.ReadAll(conn)
	if err != nil {
		printFailMetricsAndExit("could not read data from whois:", err.Error())
	}

	durationOrder := time.Since(timeSendDone).Milliseconds() + 1

	if bytes.Contains(buf, []byte(stringToLookFor)) {
		log.Printf("OK: whois replied 'alive'\n\n")
		fmt.Printf("extmon,service=%s %s=%d,%s=%d,%s=%d,%s=%d %d\n",
			"whois",
			"available", 1,
			"connect", durationConnect,
			"order", durationOrder,
			"total", durationConnect+durationOrder,
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
