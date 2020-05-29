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
)

func main() {
	log.SetOutput(os.Stderr)

	whiteflag.Alias("h", "host", "sets the whois-server used to perform the check")
	whiteflag.ParseCommandLine()
	whoisServer := whiteflag.GetString("host") + ":43"

	conn, err := net.DialTimeout("tcp", whoisServer, 10*time.Second)
	if err != nil {
		log.Printf("could not connect to %s: %s\n", whoisServer, err.Error())
		fmt.Printf("%s %d %d\n", "sensu.whois.available", 0, time.Now().Unix())
		fmt.Printf("%s %d %d\n", "sensu.whois.duration", 0, time.Now().Unix())

		if conn != nil {
			conn.Close()
		}

		os.Exit(2)
	}

	timeBegin := time.Now()

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

	whoisResponseTime := time.Since(timeBegin).Milliseconds()

	if bytes.Contains(buf, []byte(stringToLookFor)) {
		fmt.Printf("%s %d %d\n", "sensu.whois.available", 1, timeBegin.Unix())
		fmt.Printf("%s %d %d\n", "sensu.whois.duration", whoisResponseTime, timeBegin.Unix())
		_ = conn.Close()
		os.Exit(0)
	} else {
		log.Printf("whois did not reply 'alive'\n")
		fmt.Printf("%s %d %d\n", "sensu.whois.available", 0, timeBegin.Unix())
		fmt.Printf("%s %d %d\n", "sensu.whois.duration", whoisResponseTime, timeBegin.Unix())
		_ = conn.Close()
		os.Exit(2)
	}
}
