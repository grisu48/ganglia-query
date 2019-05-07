/*
 *
 */
package main

import (
	"net"
	"fmt"
	"os"
	"io"
	"time"
	"sort"
	"bufio"
	"strings"
	"strconv"
	"encoding/xml"
)

type GangliaXML struct {
	Cluster Cluster `xml:"CLUSTER"`
}

type Cluster struct {
	Name string `xml:"NAME,attr"`
	Hosts []Host `xml:"HOST"`
}

type Host struct {
	Name string `xml:"NAME,attr"`
	Ip string  `xml:"IP,attr"`
	Time int64 `xml:"REPORTED,attr"`
	StartTime int64 `xml:"GMOND_STARTED,attr"`
	Metrics []Metric `xml:"METRIC"`
}

type Metric struct {
	Name string `xml:"NAME,attr"`
	Value string `xml:"VAL,attr"`
	Unit string `xml:"UNITS,attr"`
}


func HostRow(host Host) string {
	cpu := -1.0
	mem_tot := 0.0
	mem_free := 0.0
	load1 := -1.0
	load5 := -1.0
	load15 := -1.0

	// Get relevant metrics
	for _, metric := range host.Metrics {
		if metric.Name == "cpu_idle" {
			cpu, _ = strconv.ParseFloat(metric.Value, 32)
			cpu = 100 - cpu
		}
		if metric.Name == "mem_free" {
			mem_free, _ = strconv.ParseFloat(metric.Value, 32)
		}
		if metric.Name == "mem_total" {
			mem_tot, _ = strconv.ParseFloat(metric.Value, 32)
		}
		if metric.Name == "load_one" {
			load1, _ = strconv.ParseFloat(metric.Value, 32)
		}
		if metric.Name == "load_five" {
			load5, _ = strconv.ParseFloat(metric.Value, 32)
		}
		if metric.Name == "load_fifteen" {
			load15, _ = strconv.ParseFloat(metric.Value, 32)
		}
	}

	mem := -1.0
	if mem_tot > 0 && mem_free > 0 {
		mem_used := mem_tot - mem_free
		mem = mem_used/mem_tot
	}

	then := time.Unix(host.Time, 0)		// In UTC
	time := then.Format("2006-01-02-15:04:05")

	return fmt.Sprintf("%25s\t%15s\t%20s%5.0f %%\t%5.1f %%\t%5.1f %5.1f %5.1f", host.Name, host.Ip, time, cpu, mem*100.0, load1, load5, load15)
}

func readStream(reader io.Reader) ([]byte, error) {
	data := make([]byte, 0)
	buf := make([]byte, 2048)

	for {
		n, err := reader.Read(buf)
		if err != nil { 
			if err == io.EOF {
				break
			} else {
				return data, err
			}
		}
		if n == 0 {
			break
		} else {
			data = append(data, buf[:n]...)
		}
	}

	return data, nil
}

func toUTF8(data []byte) []byte {
	str := string(data)
	str = strings.Replace(str, "ISO-8859-1", "UTF-8", 1)
	return []byte(str)
}

func main() {
	args := os.Args
	port := 8649
	if len(args) < 2 {
		fmt.Printf("Usage: %s REMOTE [PORT]\n  REMOTE is a ganglia server\n  PORT defines the port to query (default: %d)\n", args[0], port)
		return
	}

	remote := args[1] + ":" + strconv.Itoa(port)
	conn, err := net.Dial("tcp", remote)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Connection error: ", err)
		os.Exit(1)
	}
	defer conn.Close()
	// Read from conn
	data, err := readStream(bufio.NewReader(conn))
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error reading file: \n", err)
		os.Exit(1)
	}

	// Hack: Replace "ISO-8859-1" to "UTF-8" in order to make it work with XML
	data = toUTF8(data)

	// Parse XML
	var ganglia GangliaXML
	if err := xml.Unmarshal(data, &ganglia); err != nil {
 		fmt.Fprintln(os.Stderr, "Error parsing xml: ", err)
 		return
 	}

 	fmt.Printf("Cluster: %s\n", ganglia.Cluster.Name);
 	hosts := ganglia.Cluster.Hosts
 	sort.Slice(hosts, func(i, j int) bool { return strings.Compare(hosts[i].Name, hosts[j].Name) < 0 })

 	// Header
 	fmt.Printf("%25s\t%15s\t%20s%7s\t%7s\t%17s\n", "Host", "Ip", "Last Update", "CPU", "Memory", "Load (1-5-15)")
 	for _, host := range hosts {
 		fmt.Printf("%s\n", HostRow(host))
 	}

}
