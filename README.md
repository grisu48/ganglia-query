# ganglia-query

[![Build Status](https://travis-ci.org/grisu48/ganglia-query.svg?branch=master)](https://travis-ci.org/grisu48/ganglia-query)

This is a small go project to read out a `gmond` instance with enabled `tcp` server

## Usage

    ./ganglia REMOTE[:PORT][,REMOTE[:PORT]]
    e.g. ./ganglia cluster1 cluster2:8922

ganglia establishes a tcp connection to the `gmond`-server, fetches the XML contents and prints the stats as console-friendly table.

### Example

    ~/git/ganglia-query$ ./ganglia beowulf-cluster server-farm
    Cluster: beowulf-cluster
    
    Host                   	         Last Update 	CPU 	Memory	   Load (1-5-15)
    --------------------------------------------------------------------------------
    beowulf-cluster        	 	 2019-05-07-17:46:32 	8%	 95.9%	 0.7   0.8   0.8
    beowulf01               	 2019-05-07-17:46:24    0%	 88.3%	 0.0   0.0   0.1
    beowulf02               	 2019-05-07-17:46:20    0%	 88.0%	 0.0   0.0   0.1
    beowulf03               	 2019-05-07-17:46:25    0%	 85.6%	 0.0   0.0   0.1
    beowulf04               	 2019-05-07-17:46:27    0%	 88.8%	 0.0   0.0   0.1
    beowulf05               	 2019-05-07-17:46:36    0%	 91.2%	 0.0   0.0   0.1
    beowulf06               	 2019-05-07-17:46:27    8%	 98.5%	 0.9   1.0   1.0
    --------------------------------------------------------------------------------
    
    Cluster: server-farm
    
    Host                   	         Last Update 	CPU 	Memory	   Load (1-5-15)
    --------------------------------------------------------------------------------
    frontend             	 2019-05-07-17:46:31 	0%  	75.8%	 0.0   0.0   0.1
    database01             	 2019-05-07-17:46:36    1%  	5.5%	 0.0   0.0   0.1
    database02             	 2019-05-07-17:46:36    21%  	11.5%	 7.1   5.0   5.1
    --------------------------------------------------------------------------------


## Compile

    go build ganglia

Requirements

* `go >= 1.10.x`