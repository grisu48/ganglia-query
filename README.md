# ganglia-query

[![Build Status](https://travis-ci.org/grisu48/ganglia-query.svg?branch=master)](https://travis-ci.org/grisu48/ganglia-query)

This is a small go project to read out a `gmond` instance with enabled `tcp` server

## Usage

    ganglia REMOTE

ganglia established a tcp connection to the server and prints the interpreted XML as console-friendly table output

## Compile

   go build ganglia

I usually put the resulting binary to `/usr/local/bin/`
