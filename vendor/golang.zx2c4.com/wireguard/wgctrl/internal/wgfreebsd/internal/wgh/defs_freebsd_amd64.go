//go:build freebsd && amd64
// +build freebsd,amd64

// Code generated by cmd/cgo -godefs; DO NOT EDIT.
// cgo -godefs defs.go

package wgh

const (
	SizeofIfgreq = 0x10

	SIOCGWG = 0xc02069d3
	SIOCSWG = 0xc02069d2
)

type Ifgroupreq struct {
	Name   [16]byte
	Len    uint32
	Pad1   [4]byte
	Groups *Ifgreq
	Pad2   [8]byte
}

type Ifgreq struct {
	Ifgrqu [16]byte
}

type WGDataIO struct {
	Name [16]byte
	Data *byte
	Size uint64
}
