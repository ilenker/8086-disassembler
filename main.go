package main

import (
	"os"
	"fmt"
	"time"
	"github.com/ilenker/8086-disassembler/internal/sim"
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
	UNCATEGORIZED Category = iota
	DATA_TRANSFER
	ARITHMETIC
	LOGIC
	STRING_MANIPULATION
	CONTROL_TRANSFER
	PROCESSOR_CONTROL
)

func (c Category) String() string {
    return [...]string{
		"UNCATEGORIZED",
		"DATA_TRANSFER",
		"ARITHMETIC",
		"LOGIC",
		"STRING_MANIPULATION",
		"CONTROL_TRANSFER",
		"PROCESSOR_CONTROL",
	}[c]
}

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

	getArgs()

	i := 0
	var inst *Instruction
	var instructions = make([]Instruction, 0, 32)
	start := time.Now()
	for i < len(binary) {
		inst, i, err = parseInstruction(i, binary)
		instructions = append(instructions, *inst)
	}

	if args.showBinary {
		fmt.Println("Binary input: ")
		for i := range binary {
			fmt.Printf("%08b ", binary[i])
			if i % 4 == 3 {
				fmt.Println()
			}
		}
		fmt.Println("\n_________________________")
	}

	cpu := sim.CPUContext{}

	fmt.Print("bits 16\n\n")
	for _, inst := range instructions {
		s := inst.renderAsASM86()
		fmt.Println(s)
		inst.ExecInstruction(&cpu)
		fmt.Println(cpu.String())
		//if args.showBinary {
		//	fmt.Printf(inst.BinaryString())
		//}
		if args.showStruct {
			inst.printStruct()
		}
	}

	if args.verbose {
		totalTime := time.Since(start)
		fmt.Println("----------------------------------")
		fmt.Printf("Bytes Processed:\t%d\n", len(binary))
		fmt.Printf("Instruction Count:\t%d\n", len(instructions))
		fmt.Printf("Total Time:\t\t%v\n", totalTime)
		fmt.Printf("Time per instruction:\t%v\n", totalTime/time.Duration(len(instructions)))
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
		offset := int8(i.disp.Value)
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
		dest = reg
		dest = rm
		// ─────────────────────────
		// Check Ambiguity Heuristic
		// ─────────────────────────
		if i.mode != RegToReg && i.mode != ImmToReg_IMPLIED {
			if i.sbfs.W {
				size = "word "
			} else {
				size = "byte "
			}
		}
		goto out
	}

	// ──────────────
	// No Immediate
	// ──────────────
	if i.sbfs.D {
		dest = reg
		src = rm
	} else {
		dest = rm
		src = reg
	}

	out:
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

	if mod == RegToReg || mod == ImmToReg_IMPLIED {
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


var args struct {
	verbose    bool
	showBinary bool
	showStruct bool
}

func getArgs() {
	if len(os.Args) > 2 {
		for i := 2; i < len(os.Args); i++ {
			switch os.Args[i] {
			case "-v":
				args.verbose = true
			case "--show-binary":
				args.showBinary = true
			case "--show-struct":
				args.showStruct = true
			}
		}
	}
}

func Bool2Str(b bool) string {
    if b {
        return "1"
    } else {
        return "_"
    }
}

func (i *Instruction) printStruct() {
	fmt.Println("────────────────────────────")
	fmt.Printf("opCode:        %08b\n", i.opCode)
	fmt.Printf("Mnemonic:      %v\n", i.StringNoExt())
	fmt.Printf("   (ext):      %v\n", i.StringWithExt())
	fmt.Printf("sbfs:          S:%s W:%s D:%s V:%s Z:%s\n",
		Bool2Str(i.sbfs.S),
		Bool2Str(i.sbfs.W),
		Bool2Str(i.sbfs.D),
		Bool2Str(i.sbfs.V),
		Bool2Str(i.sbfs.Z),
		)
	fmt.Printf("mode:          %v\n", i.mode)
	fmt.Printf("reg_ext:       %03b\n", i.reg_ext)
	fmt.Printf("reg_mem:       %03b\n", i.reg_mem)
	fmt.Printf("disp:          %v\n", i.disp)
	fmt.Printf("imm:           %v\n", i.imm)
	fmt.Printf("regIsExt:      %v\n", i.regIsExt)
	fmt.Printf("regOnly:       %v\n", i.regOnly)
	fmt.Printf("ext:           %v\n", i.ext)
	fmt.Printf("category:      %v\n", i.category.String())
	fmt.Println("────────────────────────────")
}
