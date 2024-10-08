// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !purego

package maphash

import (
	"internal/abi"
	"reflect"
	"unsafe"
)

//go:linkname runtime_rand runtime.rand
func runtime_rand() uint64

//go:linkname runtime_memhash runtime.memhash
//go:noescape
func runtime_memhash(p unsafe.Pointer, seed, s uintptr) uintptr

func rthash(buf []byte, seed uint64) uint64 {
	if len(buf) == 0 {
		return seed
	}
	len := len(buf)
	// The runtime hasher only works on uintptr. For 64-bit
	// architectures, we use the hasher directly. Otherwise,
	// we use two parallel hashers on the lower and upper 32 bits.
	if unsafe.Sizeof(uintptr(0)) == 8 {
		return uint64(runtime_memhash(unsafe.Pointer(&buf[0]), uintptr(seed), uintptr(len)))
	}
	lo := runtime_memhash(unsafe.Pointer(&buf[0]), uintptr(seed), uintptr(len))
	hi := runtime_memhash(unsafe.Pointer(&buf[0]), uintptr(seed>>32), uintptr(len))
	return uint64(hi)<<32 | uint64(lo)
}

func rthashString(s string, state uint64) uint64 {
	buf := unsafe.Slice(unsafe.StringData(s), len(s))
	return rthash(buf, state)
}

func randUint64() uint64 {
	return runtime_rand()
}

func comparableF[T comparable](h *Hash, v T) {
	t := abi.TypeFor[T]()
	// We can only use the raw memory contents for the hash,
	// if the raw memory contents are used for computing equality.
	// That works for some types (int),
	// but not others (float, string, structs with padding, etc.)
	if t.TFlag&abi.TFlagRegularMemory != 0 {
		ptr := unsafe.Pointer(&v)
		l := t.Size()
		h.Write(unsafe.Slice((*byte)(ptr), l))
		return
	}
	vv := reflect.ValueOf(v)
	appendT(h, vv)
}
