package main

import (
	"os"
	"fmt"
	"time"
	"strconv"
	"github.com/ilenker/8086-disassembler/internal/sim"
)

/*
────────────────────────────────
Example 8086 instruction anatomy
────────────────────────────────
 MOV [Reg/Mem] [To/From] Reg 
 10001001 11011001            [DISP-LO] [DISP-HI]
          ||├─┘└─┴───001
 10001001 11└───011  R/M
 OP----DW MODE  REG         

────────────────────────────────
 MOV Immediate to register
 [1011 w reg] [data] [data if w]
────────────────────────────────
*/

var debug = false
func main() {
	var binary []byte
	var err error
	if !debug {
		if len(os.Args) < 2 {
			fmt.Println("provide binary path: <path/to/binary>")
			return
		} 
		binary, err = os.ReadFile(os.Args[1])
		if err != nil {
			fmt.Println("Failed to read file: ", os.Args[0])
			return
		}
	} else {
		binary, err = os.ReadFile("temp")
		if err != nil {
			fmt.Print("Failed to read temp file")
			return
		}
	}

	cpu := sim.CPUContext{}
	ctx := CTX{
		args: Args{},
		cpu: &cpu,
		binary: binary,
	}
	ctx.args.getArgs()
	copy(cpu.Memory[:], binary)

	ctx.doPreArgs()
	ctx.timerStart = time.Now()

	if ctx.args.exec {
		for cpu.IP < len(binary) {
			ctx.inst, _, _ = parseInstruction(cpu.IP, cpu.Memory[:])
			line := ctx.inst.renderAsASM86()
			fmt.Println(line)
			ctx.inst.ExecInstruction(&cpu)
			ctx.doArgs()
			ctx.cycles++
		}
	} else {
		inst := &Instruction{}
		i := 0
		for i < len(binary) {
			inst, i, err = parseInstruction(i, binary)
			if err != nil {
				fmt.Println(cRed("parseInstruction fail: "), err)
			}
			fmt.Println(inst.renderAsASM86())
		}
	}

	ctx.doPostArgs()
	if ctx.args.dump {
		os.WriteFile("8086_memory.data", cpu.Memory[:], 0644)
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
			return rmString + " + 0]"
		}
		if disp.Value > 0 {
			return fmt.Sprintf("%s + %d]", rmString, disp.Value)
		}
		return fmt.Sprintf("%s - %d]", rmString, -disp.Value)
	}
	return "!R/M PARSE ERROR!"
}


type Args struct {
	showMem struct{
		enabled 	bool
		start       int
		nBytes      int
		bytesPerRow int
	}
	exec,
	dump,
	stats,
	showCPU,
	showBinary,
	showStruct,
	poisonMemory,
	poisonRegisters bool
}

func (args *Args) getArgs() {
	// Defaults
	args.showMem.nBytes = 32
	args.showMem.bytesPerRow = 8
	if len(os.Args) > 2 {
		for i := 2; i < len(os.Args); i++ {
			switch os.Args[i] {
			case "--stats":
				args.stats = true
			case "--show-binary", "-bin":
				args.showBinary = true
			case "--show-struct", "-struct":
				args.showStruct = true
			case "--show-cpu", "-reg":
				args.showCPU = true
			case "--show-mem", "-mem":
				if i+1 < len(os.Args) {
					args.showMem.enabled = true
					args.showMem.start, _ = strconv.Atoi(os.Args[i+1])
					if i+2 < len(os.Args) {
						args.showMem.nBytes, _ = strconv.Atoi(os.Args[i+2])
						if i+3 < len(os.Args) {
							args.showMem.bytesPerRow, _ = strconv.Atoi(os.Args[i+3])
						}
					}
				}

			case "-pm":
				args.poisonMemory = true
			case "-pr":
				args.poisonRegisters = true
			case "-prm", "-pmr":
				args.poisonMemory = true
				args.poisonRegisters = true
			case "-exec":
				args.exec = true

			}
		}
	}
}

type CTX struct {
	binary     []byte
	args       Args
	cpu        *sim.CPUContext
	inst       *Instruction
	timerStart time.Time
	cycles     int
}

func (ctx *CTX) doPreArgs() {
	if ctx.args.showBinary {
		printBinary(ctx.binary)
	}
	if ctx.args.poisonRegisters {
		poisonRegisters(ctx.cpu)
	}
	if ctx.args.poisonMemory {
		poisonMemory(ctx.cpu)
	}
}

func (ctx *CTX) doArgs() {
	if ctx.args.showCPU {
		fmt.Printf("Flags:[%s]\n", ctx.cpu.Flags.String())
		fmt.Println(ctx.cpu.String())
	}
	if ctx.args.showBinary {
		fmt.Printf(ctx.inst.BinaryString())
	}
	if ctx.args.showStruct {
		ctx.inst.printStruct()
	}
}

func (ctx *CTX) doPostArgs() {
	if ctx.args.stats {
		totalTime := time.Since(ctx.timerStart)
		fmt.Println("----------------------------------")
		fmt.Printf("Bytes Processed:\t%d\n", len(ctx.binary))
		fmt.Printf("Cycle Count:\t%d\n", ctx.cycles)
		fmt.Printf("Total Time:\t\t%v\n", totalTime)
		if ctx.cycles > 0 {
			fmt.Printf("Time per instruction:\t%v\n", totalTime/time.Duration(ctx.cycles))
		}
	}
	if ctx.args.showMem.enabled {
		fmt.Println(ctx.cpu.InspectMemory(
			ctx.args.showMem.start,
			ctx.args.showMem.nBytes,
			ctx.args.showMem.bytesPerRow,
			ctx.args.showMem.start))
	}
}
