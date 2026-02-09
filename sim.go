package main

import (
	"fmt"
	"github.com/ilenker/8086-disassembler/internal/sim"
)

func (inst *Instruction) ExecInstruction(cpu *sim.CPUContext) {
	_ = fmt.Sprintf("")
	dest := inst.dest
	src := inst.src
	w := inst.sbfs.W

	switch inst.category {
	case DATA_TRANSFER:
		val := cpu.ValueOf(&src, w)
		cpu.WriteNoFlags(&dest, val, w)

	case ARITHMETIC:
		lhs := cpu.ValueOf(&dest, w)
		rhs := cpu.ValueOf(&src, w)

		switch ALUFunction(inst.ext) {
		case EXT_ADD:
			cpu.UpdateCOFlags(lhs, rhs, w, true)
			cpu.WriteWithFlags(&dest, lhs+rhs, w)
		case EXT_SUB:
			cpu.UpdateCOFlags(lhs, rhs, w, false)
			cpu.WriteWithFlags(&dest, lhs-rhs, w)
		case EXT_CMP:
			cpu.UpdateSZFlags(lhs-rhs)
		}
	case CONTROL_TRANSFER:
		switch inst.opCode {
		case OP_JNZ:
			if cpu.Flags&sim.FlagZero == 0 {
				nextIndex := cpu.IP + inst.size + int(src.ImmValue)
				cpu.IP = nextIndex
				return
			}
		case OP_JE:
			if cpu.Flags&sim.FlagZero != 0 {
				nextIndex := cpu.IP + inst.size + int(src.ImmValue)
				cpu.IP = nextIndex
				return
			}
		case OP_LOOP:
			cx := sim.Argument{}
			cx.Type = sim.ArgReg
			cx.Reg.Name = sim.CX
			cx.Reg.Subset = sim.SubsetX
			imm := sim.Argument{}
			imm.Type = sim.ArgImm
			imm.ImmValue = 1
			lhs := cpu.ValueOf(&cx, true)
			rhs := cpu.ValueOf(&imm, true)
			cpu.UpdateCOFlags(lhs, rhs, true, false)
			cpu.UpdateSZFlags(lhs-rhs)
			cpu.WriteWithFlags(&cx, lhs-rhs, true)
			if cpu.Flags&sim.FlagZero == 0 {
				nextIndex := cpu.IP + inst.size + int(src.ImmValue)
				cpu.IP = nextIndex
				return
			}

		}
	}
	cpu.IP = inst.binIndex + inst.size
}

func SeekNextInstructionIndex(cpu *sim.CPUContext, currentIdx int, insts []Instruction) int {
	if insts[currentIdx].binIndex == cpu.IP {
		return currentIdx
	}

	if insts[currentIdx].binIndex > cpu.IP {
		for cpu.IP != insts[currentIdx].binIndex {
			currentIdx--
		}
		return currentIdx
	}

	for currentIdx < len(insts) && cpu.IP != insts[currentIdx].binIndex {
		currentIdx++
	}
	return currentIdx
}

const (
	maskLo uint16 = 0x00ff
	maskHi uint16 = 0xff00
)
