package net

import (
	"bufio"
	"io"
	"os"
	"strconv"
	"strings"
)

var ROUTEPATH = "/proc/net/route"

func ips2ip(s string) uint32 {
	s = s[6:8] + s[4:6] + s[2:4] + s[0:2]
	r, _ := strconv.ParseUint(s, 16, 32)
	return uint32(r)
}

type RouteItem struct {
	Dest    uint32
	Gateway uint32
	Mask    uint32
	Device  string
}

type Route struct {
	routes []RouteItem
}

func NewRoute() (*Route, error) {
	r := &Route{}
	err := r.Load(ROUTEPATH)
	return r, err
}

func (r *Route) Load(fname string) error {
	f, err := os.Open(fname)
	if err != nil {
		return err
	}

	defer f.Close()
	reader := bufio.NewReader(f)
	_, _, err = reader.ReadLine()
	if err != nil {
		return err
	}

	r.routes = []RouteItem{}

	for {
		line, _, err := reader.ReadLine()
		if err == io.EOF {
			break
		}

		ss := strings.Fields(string(line))
		dev, dst, gateway, mask := ss[0], ips2ip(ss[1]), ips2ip(ss[2]), ips2ip(ss[7])
		r.routes = append(r.routes, RouteItem{
			Dest:    dst,
			Gateway: gateway,
			Mask:    mask,
			Device:  dev,
		})

	}
	return nil
}

func (r *Route) GetGateway(dst uint32) uint32 {
	ln := len(r.routes)
	for i := ln - 1; i >= 0; i-- {
		if dst&r.routes[i].Mask == r.routes[i].Dest {
			return r.routes[i].Gateway
		}
	}
	return 0
}
