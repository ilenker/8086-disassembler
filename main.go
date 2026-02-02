package main

import (
	"os"
	"fmt"
)

/*
────────────────────────────────
Example 8086 instruction anatomy
────────────────────────────────
 MOV [Reg/Mem] [To/From] Reg 
 10001001 11011001            [DISP-LO] [DISP-HI]
          ||├─┘└─┴───001
 10001001 11└---011  R/M
 OP----DW MODE  REG         

────────────────────────────────
 MOV Immediate to register
 [1011 w reg] [data] [data if w]
────────────────────────────────
*/

const (
	DATA_TRANSFER Category = iota
	ARITHMETIC
	LOGIC
	STRING_MANIPULATION
	CONTROL_TRANSFER
	PROCESSOR_CONTROL
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("provide binary path: <path/to/binary>")
		return
	}

	binary, err := os.ReadFile(os.Args[1])
	if err != nil {
		fmt.Printf("Failed to read file: %s\n", os.Args[0])
		return
	}

	i := 0
	var inst *Instruction
	var instructions = make([]Instruction, 0, 32)
	for i < len(binary) {
		inst, i, err = parseInstruction(i, binary)
		instructions = append(instructions, *inst)
	}

	for _, inst := range instructions {
		s := inst.renderAsASM86()
		fmt.Println(s)
	}
}

func (i *Instruction) renderAsASM86() string {
	empty := Instruction{}
	if  *i == empty {
		return "!Empty Instruction!"
	}
	var (
		src,
		dest,
		rm,
		reg,
		size,
		mnemonic string
	)

	if i.category == CONTROL_TRANSFER {
		offset := i.disp.Value
		mnemonic, ok := Jxxx_STRING_MAP[i.opCode]
		if ok {
			return fmt.Sprintf("%s %d", mnemonic, offset)
		}
		return "!Unknown Control Transfer Instruction!"
	}

	if i.regIsExt {
		mnemonic = i.StringWithExt()
		reg = ""
	} else {
		mnemonic = i.StringNoExt()
		reg = i.reg_ext.StringWithW(i.sbfs.W)
	}
	rm = i.RegMemString()
	size = ""

	// ──────────────
	// Has Immediate
	// ──────────────
	if i.imm.BytesConsumed != 0 {
		src = fmt.Sprintf("%d", i.imm.Value)
		dest = rm
		// ─────────────────────────
		// Check Ambiguity Heuristic
		// ─────────────────────────
		if i.mode != 0b11 {
			if i.sbfs.W {
				size = "word "
			} else {
				size = "byte "
			}
		}
	}

	// ──────────────
	// No Immediate
	// ──────────────
	if i.imm.BytesConsumed == 0 {
		if i.sbfs.D {
			dest = reg
			src = rm
		} else {
			dest = rm
			src = reg
		}
	}

	return fmt.Sprintf("%s %s%s, %s", mnemonic, size, dest, src)
}

// MOD + RM table
func (i *Instruction) RegMemString() string {
	mod  := i.mode
	rm   := i.reg_mem
	w    := i.sbfs.W
	disp := i.disp
	var rmString string

	// DIRECT ADDRESS
	// whichever 2 bytes we saved, they should now be interpreted as unsigned
	// which is why we cast to uint16
	if mod == MemMode_0 && rm == 0b110 {
		return fmt.Sprintf("[%d]", uint16(disp.Value))
	}

	if mod == RegToReg {
		return rm.StringWithW(w)
	}

	switch rm {
	case 0b000: rmString = "[bx + si"
	case 0b001: rmString = "[bx + di"
	case 0b010: rmString = "[bp + si"
	case 0b011: rmString = "[bp + di"
	case 0b100: rmString = "[si"
	case 0b101: rmString = "[di"
	case 0b110: rmString = "[bp"
	case 0b111: rmString = "[bx"
	}

	switch mod {
	case MemMode_0:
		return rmString + "]"
	case MemMode_8, MemMode_16:
		if disp.Value == 0 {
			// TODO(Johan): Unsure if "+ 0" should be included
			return rmString + " + 0]"
		}
		if disp.Value > 0 {
			return fmt.Sprintf("%s + %d]", rmString, disp.Value)
		}
		return fmt.Sprintf("%s - %d]", rmString, -disp.Value)
	}
	return "!R/M PARSE ERROR!"
}
