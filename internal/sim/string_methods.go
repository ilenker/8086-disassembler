package sim

import (
	"fmt"
	"strings"
)

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

func (cpu *CPUContext) String() string {
	str := strings.Builder{}
	str.WriteString(cRed(fmt.Sprintf("Ax(%5d): %#02x %#02x", cpu.Registers[AX], cpu.Registers[AX]>>8, byte(cpu.Registers[AX]))))
	str.WriteString(fmt.Sprintf(" | SP(%5d): %#04x\n", cpu.Registers[SP], cpu.Registers[SP]))
	str.WriteString(cBlu(fmt.Sprintf("Cx(%5d): %#02x %#02x", cpu.Registers[CX], cpu.Registers[CX]>>8, byte(cpu.Registers[CX]))))
	str.WriteString(fmt.Sprintf(" | BP(%5d): %#04x\n", cpu.Registers[BP], cpu.Registers[BP]))
	str.WriteString(cGrn(fmt.Sprintf("Dx(%5d): %#02x %#02x", cpu.Registers[DX], cpu.Registers[DX]>>8, byte(cpu.Registers[DX]))))
	str.WriteString(fmt.Sprintf(" | SI(%5d): %#04x\n", cpu.Registers[SI], cpu.Registers[SI]))
	str.WriteString(cYel(fmt.Sprintf("Bx(%5d): %#02x %#02x", cpu.Registers[BX], cpu.Registers[BX]>>8, byte(cpu.Registers[BX]))))
	str.WriteString(fmt.Sprintf(" | DI(%5d): %#04x\n", cpu.Registers[DI], cpu.Registers[DI]))

	str.WriteString(" | Flags(" + cpu.Flags.String() + ")\n")
	str.WriteString(fmt.Sprintf(" | IP: %#02x\n", cpu.IP))

	return str.String()
}

func (fs Flags) String() string {
	str := strings.Builder{}
	if fs&FlagOverflow  != 0 { str.WriteRune('O') }
	if fs&FlagDirection != 0 { str.WriteRune('D') }
	if fs&FlagInterrupt != 0 { str.WriteRune('I') }
	if fs&FlagTrap      != 0 { str.WriteRune('T') }
	if fs&FlagSign      != 0 { str.WriteRune('S') }
	if fs&FlagZero      != 0 { str.WriteRune('Z') }
	if fs&FlagAuxCarry  != 0 { str.WriteRune('A') }
	if fs&FlagParity    != 0 { str.WriteRune('P') }
	if fs&FlagCarry     != 0 { str.WriteRune('C') }
	return str.String()
}
