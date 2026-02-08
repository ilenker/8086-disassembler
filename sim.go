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
	}
}

const (
	maskLo uint16 = 0x00ff
	maskHi uint16 = 0xff00
)
