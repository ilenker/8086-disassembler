package sim

import (
	"fmt"
	"strings"
	"bytes"
)

const MEMORY_SIZE = 1<<20

type CPUContext struct{
	Registers [8]uint16
	Memory    [MEMORY_SIZE]byte
	Flags	  Flags
	IP	      int
}


type Argument struct{
	Type     ArgType
	// OperandReg: Index into CPUContext.Registers[]
	Reg struct{
		Name   RegName
		Subset RegSubset
	}

	// OperandMem: Index into CPUContext.Memory[]
	Mem struct{
		BaseReg  RegName // BX / BP
		IndexReg RegName // SI / DI
		Disp     int16
	}

	// OperandImm
	ImmValue int16
}

func (cpu *CPUContext) GetEffectiveAddress(arg *Argument) int {
	var baseVal, indexVal, dispVal int
	base  := arg.Mem.BaseReg
	index := arg.Mem.IndexReg
	
	if base == -1 {
		baseVal = 0
	} else {
		baseVal  = int(cpu.Registers[base])
	}
	if index == -1 {
		indexVal = 0
	} else {
		indexVal = int(cpu.Registers[index])
	}
	dispVal  = int(arg.Mem.Disp)
	return int(uint16(baseVal + indexVal + dispVal))
}

const (
	maskLo uint16 = 0x00ff
	maskHi uint16 = 0xff00
)

func (cpu *CPUContext) ValueOf(arg *Argument, wide bool) int16 {
	switch arg.Type {
	case ArgImm:
		return arg.ImmValue
	case ArgReg:
		switch arg.Reg.Subset {
		case SubsetX:
			return int16(cpu.Registers[arg.Reg.Name])
		case SubsetLo:
			return int16(cpu.Registers[arg.Reg.Name] & maskLo)
		case SubsetHi:
			return int16(cpu.Registers[arg.Reg.Name]>>8)
		}
	case ArgMem:
		addr := cpu.GetEffectiveAddress(arg)
		if wide {
			byteLo := int16(cpu.Memory[addr])
			byteHi := int16(cpu.Memory[addr+1]) << 8
			return byteLo | byteHi
		}
		return int16(int8(cpu.Memory[addr]))
	// TODO(Johan): how to handle "nil"
	case ArgNone:
		return 0
	}
	return 0
}

func (cpu *CPUContext) WriteNoFlags(dest *Argument, val int16, wide bool) {
	switch dest.Type {
	case ArgMem:
		if wide {
			cpu.WriteMemWide(cpu.GetEffectiveAddress(dest), val)
		} else {
			cpu.WriteMemByte(cpu.GetEffectiveAddress(dest), val)
		}
	case ArgReg:
		cpu.WriteReg(dest, val)
	}
}

func (cpu *CPUContext) WriteWithFlags(dest *Argument, val int16, wide bool) {
	cpu.UpdateSZFlags(val)
	switch dest.Type {
	case ArgMem:
		if wide {
			cpu.WriteMemWide(cpu.GetEffectiveAddress(dest), val)
			return
		}
		cpu.WriteMemByte(cpu.GetEffectiveAddress(dest), val)
	case ArgReg:
		cpu.WriteReg(dest, val)
	}
}

func (cpu *CPUContext) UpdateSZFlags(val int16) {
	switch {
	case val == 0:
		cpu.Flags &= ^FlagSign
		cpu.Flags |= FlagZero
	case val < 0:
		cpu.Flags |= FlagSign
		cpu.Flags &= ^FlagZero
	case val > 0:
		cpu.Flags &= ^FlagSign
		cpu.Flags &= ^FlagZero
	}
}

func (cpu *CPUContext) UpdateCFlag(result uint, w bool) {
	limit := uint(0xff)
	if w {
		limit = 0xffff
	}
	if result > limit {
		cpu.Flags |= FlagCarry
		return
	}
	cpu.Flags &= ^FlagCarry
}

func (cpu *CPUContext) UpdateCOFlags(lhs, rhs int16, w, isAdd bool) {
	if isAdd {
		cpu.UpdateCFlag(uint(uint16(lhs))+uint(uint16(rhs)), w)
	} else {
		cpu.UpdateCFlag(uint(uint16(lhs))-uint(uint16(rhs)), w)
	}

	if isAdd {
		result := lhs + rhs
		// pos + pos = neg
		if (lhs>0 && rhs>0) && result < 0 {
			cpu.Flags |= FlagOverflow
			return
		}
		// neg + neg = pos
		if (lhs<0 && rhs<0) && result > 0 {
			cpu.Flags |= FlagOverflow
			return
		}
		cpu.Flags &= ^FlagOverflow
		return
	}
	// Sub
	result := lhs - rhs
	// pos - neg = neg
	if (lhs>0 && rhs<0) && result < 0 {
		cpu.Flags |= FlagOverflow
		return
	}
	// neg - pos = pos
	if (lhs<0 && rhs>0) && result > 0 {
		cpu.Flags |= FlagOverflow
		return
	}
	cpu.Flags &= ^FlagOverflow
}


func (cpu *CPUContext) WriteMem(addr int, val int16, wide bool) {
	if wide {
		cpu.WriteMemWide(addr, val)
		return
	}
	cpu.WriteMemByte(addr, val)
}

func (cpu *CPUContext) WriteMemWide(addr int, val int16) {
	// TODO(Johan): check bounds
	cpu.Memory[addr]   = byte(val & int16(maskLo))
	cpu.Memory[addr+1] = byte(val>>8)
}

func (cpu *CPUContext) WriteMemByte(addr int, val int16) {
	// TODO(Johan): check bounds
	cpu.Memory[addr] = byte(val)
}

func (cpu *CPUContext) WriteReg(dest *Argument, val int16) {
	switch dest.Reg.Subset {
	case SubsetX:
		cpu.Registers[dest.Reg.Name] = uint16(val)
	case SubsetLo:
		cpu.WriteRegLo(dest.Reg.Name, val)
	case SubsetHi:
		cpu.WriteRegHi(dest.Reg.Name, val)
	}
}

func (cpu *CPUContext) WriteRegLo(reg RegName, val int16) {
	regPrev := cpu.Registers[reg]
	cpu.Registers[reg] = uint16((regPrev & maskHi) | uint16(uint8(val)))
}

func (cpu *CPUContext) WriteRegHi(reg RegName, val int16) {
	regPrev := cpu.Registers[reg]
	cpu.Registers[reg] = uint16((regPrev & maskLo) |
		uint16(val<<8))
	/*
	val := 0x1234
	Ax[0xffaa]
	   0xffaa &
	   0x00ff =
	   0x00aa |
	  (0x0034
	      <<8 =
	   0x3400)
	   0x34aa
	*/
}

func (cpu *CPUContext) InspectMemory(start, nBytes, bytesPerRow int, hiliteAddr int) string {
	if start >= MEMORY_SIZE || start < 0 {
		return "Inspect Memory: out of bounds"
	}
	if start+nBytes > MEMORY_SIZE {
		nBytes = MEMORY_SIZE-start
	}
	str := strings.Builder{}
	buf := bytes.Buffer{}
	buf.Grow(64)
	i := 0
	for {
		fmt.Fprintf(&str, "\x1b[35m%#04x:\x1b[39m", start+i)

		for range bytesPerRow {
			if start+i == hiliteAddr {
				fmt.Fprintf(&str, " \x1b[32m%02x\x1b[39m", cpu.Memory[start+i])
			} else {
				if cpu.Memory[start+i] == 0 {
					fmt.Fprintf(&str, " \x1b[38;5;242m%02x\x1b[39m", cpu.Memory[start+i])
				} else {
					fmt.Fprintf(&str, " %02x", cpu.Memory[start+i])
				} 
			}
			i++
			nBytes--
			if nBytes <= 0 {
				str.WriteRune('\n')
				return str.String()
			}
		}
		str.WriteRune('\n')
	}
}

