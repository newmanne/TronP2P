import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"
)

func main() {
	addr, err := net.ResolveUDPAddr("udp", ":11235")
	checkError(err)

	conn, err := net.ListenUDP("udp", addr)
	checkError(err)

	buf := make([]byte, 1024)
	for {
		_, raddr, err := conn.ReadFromUDP(buf)
		checkError(err)
		fmt.Fprintf(os.Stdout, "Received packet from: %s\n", raddr)
		buf = []byte(getFortune())
		n, err := conn.WriteToUDP(buf, raddr)
		fmt.Fprintf(os.Stdout, "Wrote fortune of %d bytes\n", n)
	}
}

// Returns random quote of the day using the build-in fortune
// program. If this fails, then it returns a hardcoded quote.
func getFortune() string {
	out, err := exec.Command("fortune", "-s").Output()
	if err != nil {
		return "The goal of Computer Science is to build something that will last at least until we've finished building it."
	}
	return strings.TrimSpace(string(out))
}

// If error is non-nil, print it out and halt.
func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error ", err.Error())
		os.Exit(1)
	}
}