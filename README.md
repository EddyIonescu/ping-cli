# ping-cli
Ping your favourite websites with this CLI

## Installation:

go get github.com/spf13/cobra

go get golang.org/x/net/icmp

go get golang.org/x/net/ipv4

## Usage:

`go run main.go ping [HOST | IPV4]` or `./ping-cli ping [HOST | IPV4]`

For example:

`go run main.go ping cloudflare.com`

`go run main.go ping 1.1.1.1`

Terminate by entering CTRL-C

The CLI also supports setting the time spent between sending pings, in milliseconds (default is 1000):

`go run main.go ping cloudflare.com -w 100`

`go run main.go ping cloudflare.com --wait 100`

## Design:

- Command-line parsing is in cmd/ping.go, it was set up using Cobra
- The pinging starts at StartPinging, which is at the bottom of ping/ping.go
- From there, it enters an infinite loop of calling the goroutine SendPing (using go, so it is on another thread), which uses ICMP WriteTo to send the ping.
- For every SendPing call that is made, a ReceivePing call is also made (also on another thread), which waits to receive an echo.
- Once ReceivePing parses the message, it adds the time it took to receive it (the round-trip time, RTT) to a channel.
- GenerateStats is a function that prints statistics everytime a packet is received. It does this by ranging over the channel that ReceivePing adds RTTs to, meaning
  that the loop is iterated whenever a new RTT is added to the channel. It prints the min/max/avg RTTs as well as the ratio of received to sent packets.

## Tests:

- The CLI was tested for varying network scenarios locally using the Network Link Conditioner tool that is built into MacOS.
- It allows simulating less-than-ideal networks, including slowness and packet loss.
- It was through simulating a 10% packet loss and increased latency that the benefit of using goroutines became apparent, as the
  program would have otherwise hung and been unable to ping every second and receive packets.
