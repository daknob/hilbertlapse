package main

import (
	"fmt"
	"net/netip"
	"strconv"
	"strings"
	"time"
)

type ScanLine struct {
	Status    ScanStatus
	Protocol  string
	Port      uint16
	Address   netip.Addr
	Timestamp time.Time
}

func NewScanLine(line string) (ScanLine, error) {
	var ret ScanLine

	fields := strings.Split(line, " ")

	if len(fields) != 5 {
		return ScanLine{}, fmt.Errorf("parsed fields are not 5")
	}

	/* Status */
	if fields[0] == "open" {
		ret.Status = open
	} else {
		ret.Status = closed
	}

	/* Protocol */
	ret.Protocol = fields[1]

	/* Port */
	i, err := strconv.Atoi(fields[2])
	if err != nil {
		return ScanLine{}, err
	}
	ret.Port = uint16(i)

	/* Address */
	paddr, err := netip.ParseAddr(fields[3])
	if err != nil {
		return ScanLine{}, err
	}
	ret.Address = paddr

	/* Timestamp */
	tsp, err := strconv.Atoi(fields[4])
	if err != nil {
		return ScanLine{}, err
	}
	ret.Timestamp = time.Unix(int64(tsp), 0)

	return ret, nil
}

type ScanStatus int

const (
	open ScanStatus = iota
	closed
)
