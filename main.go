package main

import (
	"flag"
	"sync"

	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"

	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/client9/reopen"
)

// The option82 program reads DHCPv4 packets via libpcap (network or file input)
// and ouputs JSON strings to a log file or to Stdout containing fields that
// should aid network troubelshooting, incident handling, or forensics.
func main() {
	srcFile := flag.String("f", "", "PCAP input file")
	srcInt := flag.String("i", "", "Capture interface")
	outFile := flag.String("o", "",
		"Log file (messages go to stdout if absent)")
	pidFile := flag.String("p", "", "PID file (optional)")

	flag.Parse()

	var handle *pcap.Handle = nil
	if *srcFile != "" && *srcInt != "" {
		log.Fatal("Cannot input from file and network at the same time")
	} else if *srcFile != "" {
		var err error = nil
		handle, err = pcap.OpenOffline(*srcFile)
		if err != nil {
			log.Fatalf("Problem opening pcap file: %s", err)
		}
	} else if *srcInt != "" {
		var err error = nil
		handle, err = pcap.OpenLive(*srcInt, 1600, true, pcap.BlockForever)
		if err != nil {
			log.Fatalf("Problem opening pcap interface: %s", err)
		}
	} else {
		log.Fatal("Aborting: you must specify -i XOR -f")
	}

	if *outFile != "" {
		f, err := reopen.NewFileWriter(*outFile)
		if err != nil {
			log.Fatalf("Unable to set output log: %s", err)
		}
		log.SetOutput(f)
		sighup := make(chan os.Signal, 1)
		signal.Notify(sighup, syscall.SIGHUP)
		go func() {
			for {
				<-sighup
				log.Println("Got a sighup, reopening log file.")
				f.Reopen()
			}
		}()
	} else {
		// Output to Stdout seems more useful if not logging to file.
		log.SetOutput(os.Stdout)
	}

	if *pidFile != "" {
		err := writePidFile(*pidFile)
		if err != nil {
			log.Fatalf("Problem writing pid file: %s", err)
		}
	}

	var wg sync.WaitGroup
	mapChan := make(chan map[string]interface{}, 1000)

	wg.Add(1)
	go metricsLoop(mapChan, &wg)

	// TODO: Should be possible to override BPF rule with a flag
	if err := handle.SetBPFFilter("port 67 or port 68 and udp"); err != nil {
		log.Fatalf("Unable to set BPF: %s", err)
	} else {
		packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
		for packet := range packetSource.Packets() {
			result, hasHostname := HandlePacket(packet)
			if hasHostname {
				mapChan <- *result
				// hasMetrics := false
				// if client_ip, ok := (*result)["client_ip"].(string); ok {
				// 	hasMetrics = discoverPrometheusEndpoint(client_ip)
				// 	(*result)["has_metrics"] = hasMetrics
				// }
			}
		}
	}
	close(mapChan)
	fmt.Println("Waiting for metricsLoop to finish")
	wg.Wait()
}

// Write a pid file, but first make sure it doesn't exist with a running pid.
// https://gist.github.com/davidnewhall/3627895a9fc8fa0affbd747183abca39
func writePidFile(pidFile string) error {
	// Read in the pid file as a slice of bytes.
	piddata, err := ioutil.ReadFile(pidFile)
	if err == nil {
		// Convert the file contents to an integer.
		pid, err := strconv.Atoi(string(piddata))
		if err == nil {
			// Look for the pid in the process list.
			process, err := os.FindProcess(pid)
			if err == nil {
				// Send the process a signal zero kill.
				err := process.Signal(syscall.Signal(0))
				if err == nil {
					// We only get an error if the pid isn't running,
					// or it's not ours.
					return fmt.Errorf("pid already running: %d", pid)
				}
			}
		}
	}
	// If we get here, then the pidfile didn't exist,
	// or the pid in it doesn't belong to the user running this app.
	return ioutil.WriteFile(pidFile,
		[]byte(fmt.Sprintf("%d", os.Getpid())), 0664)
}
