package dialer

import (
	"net"
	"proxychan/internal/socks5" // Import the socks5 package
)

// socks5ConnectOverConn is now replaced by socks5.ConnectOverConn
func socks5ConnectOverConn(c net.Conn, address string) error {
	return socks5.ConnectOverConn(c, address)
}
