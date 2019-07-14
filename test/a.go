fi, err := net.InterfaceByName("eth0")
if err != nil {
    log.Fatalf("failed to open interface: %v", err)
}// Open a raw socket using same EtherType as our frame.
c, err := raw.ListenPacket(ifi, 0xcccc)
if err != nil {
    log.Fatalf("failed to listen: %v", err)
}
defer c.Close()// Accept frames up to interface's MTU in size.
b := make([]byte, ifi.MTU)
var f ethernet.Frame// Keep reading frames.
for {
    n, addr, err := c.ReadFrom(b)
    if err != nil {
        log.Fatalf("failed to receive message: %v", err)
    }    // Unpack Ethernet frame into Go representation.
    if err := (&f).UnmarshalBinary(b[:n]); err != nil {
        log.Fatalf("failed to unmarshal ethernet frame: %v", err)
    }    // Display source of message and message itself.
    log.Printf("[%s] %s", addr.String(), string(f.Payload))
}
