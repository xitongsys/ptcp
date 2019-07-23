package netinfo

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

var ROUTEPATH = "/proc/net/route"

type RouteItem struct {
	Dest    uint32
	Gateway uint32
	Mask    uint32
	Device  string
}

func (ri *RouteItem) String() string {
	return fmt.Sprintf("{Dest:%v, GateWay:%v, Mask:%v, Device:%v}", ip2s(ri.Dest), ip2s(ri.Gateway), ip2s(ri.Mask), ri.Device)
}

type Route struct {
	routes []*RouteItem
}

func NewRoute() (*Route, error) {
	r := &Route{}
	err := r.Load(ROUTEPATH)
	return r, err
}

func (r *Route) String() string {
	res := "["
	for _, v := range r.routes {
		res += v.String()
	}
	res += "]"
	return res
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

	r.routes = []*RouteItem{}

	for {
		line, _, err := reader.ReadLine()
		if err == io.EOF {
			break
		}

		ss := strings.Fields(string(line))
		dev, dst, gateway, mask := ss[0], iprs2ip(ss[1]), iprs2ip(ss[2]), iprs2ip(ss[7])
		r.routes = append(r.routes, &RouteItem{
			Dest:    dst,
			Gateway: gateway,
			Mask:    mask,
			Device:  dev,
		})

	}
	return nil
}

func (r *Route) GetGateway(dst uint32) (uint32, error) {
	ln := len(r.routes)
	for i := ln - 1; i >= 0; i-- {
		if dst&r.routes[i].Mask == r.routes[i].Dest {
			return r.routes[i].Gateway, nil
		}
	}
	return 0, fmt.Errorf("can't find route")
}
