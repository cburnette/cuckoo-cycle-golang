package main

import (
	"crypto/sha256"
	"fmt"
	"os"
)

const nthreads = 1
const maxPathLen = 4096
const edgeBits = 19
const nedges uint = 1 << edgeBits
const nodeBits = edgeBits + 1
const nnodes uint = 1 << nodeBits
const edgeMask uint = nedges - 1
const proofSize = 42

var header = "261"
var headerbytes = []byte(header)
var maxsols uint = 8
var easipct uint = 50
var easiness uint = easipct * nnodes / 100

var V [4]uint

//sols := make([][proofSize]int, maxsols)
var cuckoo = make([]uint, 1+nnodes)

//var cuckoo [1 + nnodes]uint
var nsols uint

func u8to64(p [32]byte, i uint) uint {
	return ((uint)(p[i]) & 0xff) |
		((uint)(p[i+1])&0xff)<<8 |
		((uint)(p[i+2])&0xff)<<16 |
		((uint)(p[i+3])&0xff)<<24 |
		((uint)(p[i+4])&0xff)<<32 |
		((uint)(p[i+5])&0xff)<<40 |
		((uint)(p[i+6])&0xff)<<48 |
		((uint)(p[i+7])&0xff)<<56
}

func sipnode(nonce uint, uOrV uint) uint {
	return siphash24(2*nonce+uOrV) & edgeMask
}

func path(u uint, us *[maxPathLen]uint) uint {
	var nu uint
	nu = 0
	for u != 0 {
		nu += 1
		if nu >= maxPathLen {
			fmt.Println("error calculating path. aborting...")
			os.Exit(1)
		}

		us[nu] = u
		u = cuckoo[u]
	}

	return nu
}

func main() {
	fmt.Printf("MAXPATHLEN: %d\n", maxPathLen)
	fmt.Printf("HEADERHEX: %x\n", header)
	fmt.Printf("EDGEBITS: %d\n", edgeBits)
	fmt.Printf("NEDGES: %d\n", nedges)
	fmt.Printf("NNODEBITS: %d\n", nodeBits)
	fmt.Printf("NNODES: %d\n", nnodes)
	fmt.Printf("EDGEMASK: %d\n", edgeMask)
	fmt.Printf("PROOFSIZE: %d\n", proofSize)
	fmt.Printf("EASINESS: %d\n", easiness)

	hdrkey := sha256.Sum256(headerbytes)
	fmt.Printf("hdrkey: %x\n", hdrkey)

	V[0] = u8to64(hdrkey, 0)
	V[1] = u8to64(hdrkey, 8)
	V[2] = u8to64(hdrkey, 16)
	V[3] = u8to64(hdrkey, 24)

	fmt.Printf("V: %d\n", V)

	var us [maxPathLen]uint
	var vs [maxPathLen]uint

	fmt.Printf("siphash: %d\n", siphash24(0))

	var nonce uint
	for nonce = 0; nonce < easiness; nonce += nthreads {
		us[0] = sipnode(nonce, 0)
		u := cuckoo[us[0]]

		vs[0] = nedges + sipnode(nonce, 1)
		v := cuckoo[vs[0]]

		if u == vs[0] || v == us[0] {
			continue
		}

		nu := path(u, &us)
		nv := path(v, &vs)

		if us[nu] == vs[nv] {
			var min uint
			if nu < nv {
				min = nu
			} else {
				min = nv
			}

			nu -= min
			nv -= min
			for us[nu] != vs[nv] {
				nu += 1
				nv += 1
			}

			len := nu + nv + 1
			fmt.Printf("%d-cycle found at %d%%\n", len, nonce*100/easiness)
			if len == proofSize && nsols < maxsols {
				fmt.Println("calculate solution")
			}

			continue
		}

		if nu < nv {
			for nu != 0 {
				nu -= 1
				cuckoo[us[nu+1]] = us[nu]
			}
			cuckoo[us[0]] = vs[0]
		} else {
			for nv != 0 {
				nv -= 1
				cuckoo[vs[nv+1]] = vs[nv]
			}
			cuckoo[vs[0]] = us[0]
		}
	}
}
