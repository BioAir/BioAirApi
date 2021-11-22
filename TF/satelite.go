package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {
	bufferIn := bufio.NewReader(os.Stdin)
	fmt.Print("Ingrese el puerto remoto:")
	puerto, _ := bufferIn.ReadString('\n')
	puerto = strings.TrimSpace(puerto)
	remotehost := fmt.Sprintf("localhost:%s", puerto)

	con, _ := net.Dial("tcp", remotehost)
	defer con.Close()
	fmt.Fprintln(con, 35)
}
