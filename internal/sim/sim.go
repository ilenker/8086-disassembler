package sim

import (
	"fmt"
	"strings"
)

type CPUContext struct {
	Registers [8]uint16
	Memory    [1<<20]byte
}

type ArgType uint8
const (
	ArgNone ArgType = iota
	ArgReg       // cpu.registers[abcd sp bp si di...]
	ArgMem       // cpu.memory[...]
	ArgImm       // Literal
)
func (a ArgType) String() string {
    return [...]string{
	"ArgNone",
	"ArgReg",
	"ArgMem",
	"ArgImm",
	}[a]
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

type RegName int8
const (
	AX RegName = iota 
	CX
	DX
	BX
	SP // Stack pointer
	BP // Base pointer
	SI // Source index
	DI // Destination index
)
const NilX RegName = -1

type RegSubset int8
const (
	SubsetX RegSubset = iota
	SubsetLo
	SubsetHi
)
func (r RegSubset) String() string {
	switch r {
	case SubsetX:
	return "(x)"
	case SubsetLo:
	return "(l)"
	case SubsetHi:
	return "(h)"
	default:
	return "?"
	}
}

func (r RegName) String() string {
	if r == -1 {
		return "[no reg]"
	}
	return [...]string{
		"a",
		"c",
		"d",
		"b",
		"sp",
		"bp",
		"si",
		"di",
	}[r]
}

func (arg Argument) String() string {
	str := strings.Builder{}
	switch arg.Type {
	case ArgNone:
		str.WriteString("none")

	case ArgReg:
		str.WriteString("Register: ")
		str.WriteString(arg.Reg.Name.String())
		str.WriteString(arg.Reg.Subset.String())

	case ArgMem:
		str.WriteString("Memory: ")
		str.WriteString("Base(")
		str.WriteString(arg.Mem.BaseReg.String())
		str.WriteString(") Index(")
		str.WriteString(arg.Mem.IndexReg.String())
		str.WriteString(") Disp(")
		str.WriteString(fmt.Sprintf("%d)", arg.Mem.Disp))

	case ArgImm:
		str.WriteString("Immediate: ")
		str.WriteString(fmt.Sprintf("%d", arg.ImmValue))

	}
	return str.String()
}

func (cpu *CPUContext) GetEffectiveAddress(arg Argument) uint16 {
	base  := int16(cpu.Registers[arg.Mem.BaseReg])
	index := int16(cpu.Registers[arg.Mem.IndexReg])
	disp  := arg.Mem.Disp
	return uint16(base + index + disp)
}

func (cpu *CPUContext) String() string {

	str := strings.Builder{}

	str.WriteString(fmt.Sprintf("A(%5d): %#x, %#x\n", cpu.Registers[AX], cpu.Registers[AX]>>8, byte(cpu.Registers[AX])))
	str.WriteString(fmt.Sprintf("C(%5d): %#x, %#x\n", cpu.Registers[CX], cpu.Registers[CX]>>8, byte(cpu.Registers[CX])))
	str.WriteString(fmt.Sprintf("D(%5d): %#x, %#x\n", cpu.Registers[DX], cpu.Registers[DX]>>8, byte(cpu.Registers[DX])))
	str.WriteString(fmt.Sprintf("B(%5d): %#x, %#x\n", cpu.Registers[BX], cpu.Registers[BX]>>8, byte(cpu.Registers[BX])))

	str.WriteString(fmt.Sprintf("SP(%5d)\n", cpu.Registers[SP]))
	str.WriteString(fmt.Sprintf("BP(%5d)\n", cpu.Registers[BP]))
	str.WriteString(fmt.Sprintf("SI(%5d)\n", cpu.Registers[SI]))
	str.WriteString(fmt.Sprintf("DI(%5d)\n", cpu.Registers[DI]))

	return str.String()
}
