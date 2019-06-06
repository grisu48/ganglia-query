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
	//"golang.org/x/crypto/ssh/terminal"
)

// Terminal color codes
const KNRM = "\x1B[0m"
const KRED = "\x1B[31m"
const KGRN = "\x1B[32m"
const KYEL = "\x1B[33m"
const KBLU = "\x1B[34m"
const KMAG = "\x1B[35m"
const KCYN = "\x1B[36m"
const KWHT = "\x1B[37m"


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

func condIf(cond bool, true_s string, false_s string) string {
	if(cond) {
		return true_s
	} else {
		return false_s
	}
}

func HostRow(host Host, useColors bool) string {
	cpu := -1.0
	mem_tot := 0.0
	mem_free := 0.0
	load1 := -1.0
	load5 := -1.0
	load15 := -1.0
	cpu_num := 1

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
		if metric.Name == "cpu_num" {
			cpu_num, _ = strconv.Atoi(metric.Value)
		}
	}

	mem := -1.0
	if mem_tot > 0 && mem_free > 0 {
		mem_used := mem_tot - mem_free
		mem = mem_used/mem_tot
	}

	then := time.Unix(host.Time, 0)		// In UTC
	stime := then.Format("2006-01-02-15:04:05")

	if useColors {
		ret := fmt.Sprintf("%-23s\t", host.Name)
		ret += fmt.Sprintf("%s", condIf( time.Since(then).Hours()>1,KRED,KGRN))
		ret += fmt.Sprintf("%20s", stime)
		ret += fmt.Sprintf("%s", KNRM)
		ret += fmt.Sprintf("%s", condIf( cpu<0,KRED, condIf(cpu<25, KYEL, KGRN)))
		ret += fmt.Sprintf("%5.0f%%\t", cpu)
		ret += fmt.Sprintf("%s", KNRM)
		ret += fmt.Sprintf("%s", condIf( mem<0,KRED, condIf(mem>0.8, KRED, condIf(mem>0.6, KYEL, KGRN))))
		ret += fmt.Sprintf("%5.1f%%\t", mem*100.0)
		ret += fmt.Sprintf("%s", KNRM)

		fload1 := load1 / float64(cpu_num)
		fload5 := load5 / float64(cpu_num)
		fload15 := load15 / float64(cpu_num)

		ret += fmt.Sprintf("%s", condIf( fload1<0,KRED, condIf(fload1>0.8, KRED, condIf(fload1>0.6, KYEL, KGRN))))
		ret += fmt.Sprintf("%4.1f  ", load1)
		ret += fmt.Sprintf("%s", condIf( fload5<0,KRED, condIf(fload5>0.8, KRED, condIf(fload5>0.6, KYEL, KGRN))))
		ret += fmt.Sprintf("%4.1f  ", load5)
		ret += fmt.Sprintf("%s", condIf( fload15<0,KRED, condIf(fload15>0.8, KRED, condIf(fload15>0.6, KYEL, KGRN))))
		ret += fmt.Sprintf("%4.1f", load15)
		ret += fmt.Sprintf("%s", KNRM)
		return ret
	} else {
		return fmt.Sprintf("%-23s\t%20s%5.0f%%\t%5.1f%%\t%4.1f  %4.1f  %4.1f", host.Name, stime, cpu, mem*100.0, load1, load5, load15)
	}
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
	if len(args) < 2 {
		fmt.Printf("Usage: %s REMOTE[:PORT][,REMOTE[:PORT]]\n  REMOTE is a ganglia server, PORT (optionally) defines the port to query (default: %d)\n", args[0], 8649)
		fmt.Printf("       Multiple remote configurations are possible and then listed after each other\n")
		return
	}
	
	useColors := true //terminal.IsTerminal(int(os.Stdout.Fd()))
	for _, arg := range(args[1:]) {
		port := 8649
		remote := arg
		
		if i := strings.Index(remote, ":"); i > -1 {
			remote = arg[:i]
			port, _ = strconv.Atoi(arg[i+1:])
		}
		
		remote = remote + ":" + strconv.Itoa(port)
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

		data = toUTF8(data)	// Hack: Replace "ISO-8859-1" to "UTF-8" in order to make it work with XML
		// Parse XML
		var ganglia GangliaXML
		if err := xml.Unmarshal(data, &ganglia); err != nil {
	 		fmt.Fprintln(os.Stderr, "Error parsing xml: ", err)
	 		return
	 	}

	 	fmt.Printf("Cluster: %s\n\n", ganglia.Cluster.Name);
	 	hosts := ganglia.Cluster.Hosts
	 	sort.Slice(hosts, func(i, j int) bool { return strings.Compare(hosts[i].Name, hosts[j].Name) < 0 })

	 	// Header
	 	fmt.Printf("%-23s\t%20s%7s\t%7s\t%16s\n", "Host", "Last Update", "CPU", "Memory", "Load (1-5-15)")
	 	fmt.Println("--------------------------------------------------------------------------------")
	 	for _, host := range hosts {
	 		fmt.Printf("%s\n", HostRow(host, useColors))
	 	}
	 	fmt.Println("--------------------------------------------------------------------------------")
 	}

}
