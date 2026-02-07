package sim


// Flag masks
type Flags uint16
const (
	FlagOverflow  Flags = 0b1000_0000_0000
	FlagDirection Flags = 0b0100_0000_0000
	FlagInterrupt Flags = 0b0010_0000_0000
	FlagTrap      Flags = 0b0001_0000_0000
	FlagSign      Flags = 0b0000_1000_0000
	FlagZero      Flags = 0b0000_0100_0000
	FlagAuxCarry  Flags = 0b0000_0001_0000
	FlagParity    Flags = 0b0000_0000_0100
	FlagCarry     Flags = 0b0000_0000_0001
	
	// [_ _ _ _] [o d i t] [s z _ a] [_ p _ c]
)

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



type RegName int8
const NilX RegName = -1
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
