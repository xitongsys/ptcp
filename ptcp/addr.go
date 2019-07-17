package ptcp

type Addr struct {
	addr string
}

func NewAddr(addr string) *Addr {
	return &Addr{
		addr: addr,
	}
}

func (a *Addr) Network() string {
	return "ptcp"
}

func (a *Addr) String() string {
	return a.addr
}
