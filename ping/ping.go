package ping

import (  
    "net"
    "os"
    "golang.org/x/net/icmp"
    "golang.org/x/net/ipv4"
    "log"
    "fmt"
    "time"
    "math"
    "strings"
)

type Ping struct {
    Address string
    ICMPSeq int
    TimeSent int64
}

func (ping Ping) sendPing(connection *icmp.PacketConn) {
    fmt.Println("Sending Ping icmp_seq=", ping.ICMPSeq)
    echo := icmp.Echo {
        ID: os.Getpid() & 0xffff,
        Seq: ping.ICMPSeq,
        // ID is unique per ping process,
        // (ID, Seq) uniquely identifies ping response
        Data: make([]byte, 56),
    }
    
    message := icmp.Message {
        Type: ipv4.ICMPTypeEcho,
        Code: 0, // ICMP code for EchoRequest
        Body: &echo,
    }
    
    messageBytes, err := message.Marshal(nil)
	if err != nil {
		log.Fatal(err)
    }
    dstAddr, err := net.ResolveUDPAddr("udp4", ping.Address)
    if err != nil {
        log.Fatal(err)
    }

    _, err = connection.WriteTo(
        messageBytes,
        dstAddr,
        // Pass UDPAddr instead of IPAddr as connection is non-privilleged
    )
    if err != nil {
		log.Fatal(err)
	}
}

func printFloat(num float64) string {
    // returns string containing float truncated to two decimal places
    return fmt.Sprintf("%.2f", num) 
}

func receivePing(
    connection *icmp.PacketConn,
    timesChannel chan int64,
    pingBySeq map[int]Ping,
) {
    readBytes := make([]byte, 64)
    size, addr, err := connection.ReadFrom(readBytes)
	if err != nil {
		log.Fatal(err)
	}
    readMessage, err := icmp.ParseMessage(1, readBytes[:size])
    if err != nil {
        log.Println(err)
        return
    }
    fmt.Print("Received Echo from ", addr)
    switch messageBody := readMessage.Body.(type) {
        case *icmp.Echo:
            fmt.Print(": icmp_seq=", messageBody.Seq)
            echoDuration := time.Now().UnixNano() - pingBySeq[messageBody.Seq].TimeSent
            fmt.Println(" time=", printFloat(float64(echoDuration) / 1e6), "ms")
            // successful pings are counted by adding them to timesChannel
            timesChannel <- echoDuration
        default:
            log.Println("Received unexpected response, expected echo")
    }
}

func generateStats(timesChannel chan int64, pingsSent *int) {
    // Generates stats whenever a ping is received
    // (ie. when a new time duration is added to timesChannel)
    maxLatency := 0.0
    minLatency := math.MaxFloat64
    avgLatency := 0.0
    pingsReceived := 0
    for pingReceivedTimeDuration := range timesChannel {
        pingDurationMs := float64(pingReceivedTimeDuration) / 1e6
        pingsReceived++
        maxLatency = math.Max(maxLatency, pingDurationMs)
        minLatency = math.Min(minLatency, pingDurationMs)
        avgLatency =
            (avgLatency * float64(pingsReceived-1) + pingDurationMs) / float64(pingsReceived)
        pingsSucceeded := float64(pingsReceived) / float64(*pingsSent) * 100
        fmt.Println("Sent:", *pingsSent, " Received:", pingsReceived)
        fmt.Println("Success Rate:", printFloat(pingsSucceeded), "%")
        fmt.Println(
            "round-trip Max/Min/Avg =",
            printFloat(maxLatency), "/",
            printFloat(minLatency), "/",
            printFloat(avgLatency), "ms",
        )
    }
}

func StartPinging(dest string, waitTimeMs int) {
    fmt.Println("Begin Pinging ", dest)
    // Use IPV4 wildcard, listen to all incoming packets
    connection, err := icmp.ListenPacket("udp4", "0.0.0.0")
    if err != nil {
            log.Fatal(err)
    }
    if !strings.Contains(dest, ":") {
        // provide port number if not provided
        dest += ":7" // port for receiving echo requests
    }

    var ICMPSeq int = 0

    // channel for receiving the time duration of each echo response
    var timesChannel = make(chan int64)

    // global map for obtaining the original ping from its ICMP sequence
    pingBySeq := make(map[int]Ping)

    go generateStats(timesChannel, &ICMPSeq)

    for {
        ICMPSeq++
        ping := Ping {
            Address: dest,
            ICMPSeq: ICMPSeq,
            TimeSent: time.Now().UnixNano(),
        }
        pingBySeq[ICMPSeq] = ping
        go ping.sendPing(connection)
        go receivePing(connection, timesChannel, pingBySeq)
        time.Sleep(time.Duration(waitTimeMs) * time.Millisecond)
    }
}