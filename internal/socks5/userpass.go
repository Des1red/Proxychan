package socks5

import (
	"errors"
	"io"
)

func readUserPassAuth(r io.Reader) (string, string, error) {
	// RFC1929:
	// VER=1, ULEN, UNAME, PLEN, PASSWD
	var ver [1]byte
	if _, err := io.ReadFull(r, ver[:]); err != nil {
		return "", "", err
	}
	if ver[0] != 0x01 {
		return "", "", errors.New("socks5: bad auth version")
	}

	var ulen [1]byte
	if _, err := io.ReadFull(r, ulen[:]); err != nil {
		return "", "", err
	}
	if ulen[0] == 0 {
		return "", "", errors.New("socks5: empty username")
	}

	uname := make([]byte, int(ulen[0]))
	if _, err := io.ReadFull(r, uname); err != nil {
		return "", "", err
	}

	var plen [1]byte
	if _, err := io.ReadFull(r, plen[:]); err != nil {
		return "", "", err
	}

	pass := make([]byte, int(plen[0]))
	if _, err := io.ReadFull(r, pass); err != nil {
		return "", "", err
	}

	return string(uname), string(pass), nil
}

func writeUserPassStatus(w io.Writer, status byte) error {
	// VER=1, STATUS
	_, err := w.Write([]byte{0x01, status})
	return err
}
