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

func main() {
	log.SetOutput(os.Stderr)

	whiteflag.Alias("host", "whoisserver", "sets the whois-server used to perform the check")
	whiteflag.ParseCommandLine()
	whoisServer := whiteflag.GetString("whoisserver") + ":43"

	timeBegin := time.Now()

	conn, err := net.DialTimeout("tcp", whoisServer, 10*time.Second)
	if err != nil {
		log.Printf("could not connect to %s: %s\n", whoisServer, err.Error())
		fmt.Printf("%s %d %d\n", "sensu.whois.available", 0, timeBegin.Unix())
		fmt.Printf("%s %d %d\n", "sensu.whois.duration", 0, timeBegin.Unix())

		if conn != nil {
			conn.Close()
		}

		os.Exit(2)
	}

	_, err = conn.Write([]byte("alive@whois" + "\r\n"))
	if err != nil {
		log.Printf("could not write to whois: %s\n", err.Error())
		fmt.Printf("%s %d %d\n", "sensu.whois.available", 0, timeBegin.Unix())
		fmt.Printf("%s %d %d\n", "sensu.whois.duration", 0, timeBegin.Unix())
		_ = conn.Close()
		os.Exit(2)
	}

	buf, err := ioutil.ReadAll(conn)
	if err != nil {
		log.Printf("could not read from whois: %s\n", err.Error())
		fmt.Printf("%s %d %d\n", "sensu.whois.available", 0, timeBegin.Unix())
		fmt.Printf("%s %d %d\n", "sensu.whois.duration", 0, timeBegin.Unix())
		_ = conn.Close()
		os.Exit(2)
	}

	if bytes.Contains(buf, []byte("alive")) {
		whoisResponseTime := time.Since(timeBegin)

		fmt.Printf("%s %d %d\n", "sensu.whois.available", 1, timeBegin.Unix())
		fmt.Printf("%s %d %d\n", "sensu.whois.duration", whoisResponseTime.Milliseconds(), timeBegin.Unix())
		_ = conn.Close()
		os.Exit(0)
	} else {
		log.Printf("whois did not reply 'alive'\n")
		fmt.Printf("%s %d %d\n", "sensu.whois.available", 0, timeBegin.Unix())
		fmt.Printf("%s %d %d\n", "sensu.whois.duration", 0, timeBegin.Unix())
		_ = conn.Close()
		os.Exit(2)
	}
}
