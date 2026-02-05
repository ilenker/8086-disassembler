package main

import (
	"github.com/ilenker/8086-disassembler/internal/sim"
)

func (inst *Instruction) ExecInstruction(cpu *sim.CPUContext) {
	dest := inst.dest
	src := inst.src

	switch {
	case dest.Type == sim.ArgMem:
		addrDest := cpu.GetEffectiveAddress(dest)
		switch src.Type {
		case sim.ArgReg:
			switch src.Reg.Subset {
			case sim.SubsetX:
				cpu.Memory[addrDest]   = byte(cpu.Registers[src.Reg.Name] >> 8)
				cpu.Memory[addrDest+1] = byte(cpu.Registers[src.Reg.Name])
			case sim.SubsetLo:
				cpu.Memory[addrDest] = byte(cpu.Registers[src.Reg.Name])
			case sim.SubsetHi:
				cpu.Memory[addrDest] = byte(cpu.Registers[src.Reg.Name] >> 8)
			}
		case sim.ArgImm:
			// TODO(Johan): check width
			cpu.Memory[addrDest] = byte(src.ImmValue >> 8)
			cpu.Memory[addrDest+1] = byte(src.ImmValue)
		}

	case dest.Type == sim.ArgReg &&
	     src.Type  == sim.ArgMem:
		addrSrc := cpu.GetEffectiveAddress(src)
		switch dest.Reg.Subset {
		case sim.SubsetX:
			byteHi := uint16(cpu.Memory[addrSrc])
			byteLo := uint16(cpu.Memory[addrSrc+1])
			cpu.Registers[dest.Reg.Name] = (byteHi<<8) | byteLo
		case sim.SubsetLo:
			regx := cpu.Registers[dest.Reg.Name]
			byteLo := cpu.Memory[addrSrc]
			cpu.Registers[dest.Reg.Name] = (regx & 0xff00) | uint16(byteLo)
		case sim.SubsetHi:
			regx := cpu.Registers[dest.Reg.Name]
			byteHi := cpu.Memory[addrSrc]
			cpu.Registers[dest.Reg.Name] = (regx & 0x00ff) | (uint16(byteHi)<<8)
		}

	case dest.Type == sim.ArgReg &&
		 src.Type == sim.ArgImm:
		cpu.Registers[dest.Reg.Name] = uint16(src.ImmValue)

	case dest.Type == sim.ArgReg &&
	     src.Type  == sim.ArgReg:
		switch {

		case dest.Reg.Subset == sim.SubsetX &&
			 src.Reg.Subset  == sim.SubsetX:
			cpu.Registers[dest.Reg.Name] = cpu.Registers[src.Reg.Name]

		case dest.Reg.Subset == sim.SubsetX &&
			 src.Reg.Subset  == sim.SubsetLo:
			cpu.Registers[dest.Reg.Name] =
			cpu.Registers[dest.Reg.Name]&0xff00 |
			cpu.Registers[src.Reg.Name ]&0x00ff

		case dest.Reg.Subset == sim.SubsetX &&
			 src.Reg.Subset  == sim.SubsetHi:
			cpu.Registers[dest.Reg.Name] =
			cpu.Registers[dest.Reg.Name]&0x00ff |
			cpu.Registers[src.Reg.Name ]&0xff00


		case dest.Reg.Subset == sim.SubsetLo &&
			 src.Reg.Subset  == sim.SubsetX:
			cpu.Registers[dest.Reg.Name] =
			cpu.Registers[dest.Reg.Name]&0xff00 |
			cpu.Registers[src.Reg.Name ]&0x00ff

		case dest.Reg.Subset == sim.SubsetLo &&
			 src.Reg.Subset  == sim.SubsetLo:
			cpu.Registers[dest.Reg.Name] =
			cpu.Registers[dest.Reg.Name]&0x00ff |
			cpu.Registers[src.Reg.Name ]&0x00ff

		case dest.Reg.Subset == sim.SubsetLo &&
			 src.Reg.Subset  == sim.SubsetHi:
			cpu.Registers[dest.Reg.Name] =
			cpu.Registers[dest.Reg.Name]&0x00ff |
			cpu.Registers[src.Reg.Name ] >> 8


		case dest.Reg.Subset == sim.SubsetHi &&
			 src.Reg.Subset  == sim.SubsetX:
			cpu.Registers[dest.Reg.Name] =
			cpu.Registers[dest.Reg.Name]&0x00ff |
			cpu.Registers[src.Reg.Name ]&0xff00

		case dest.Reg.Subset == sim.SubsetHi &&
			 src.Reg.Subset  == sim.SubsetLo:
			cpu.Registers[dest.Reg.Name] =
			cpu.Registers[dest.Reg.Name]&0x00ff |
			cpu.Registers[src.Reg.Name ] << 8

		case dest.Reg.Subset == sim.SubsetHi &&
			 src.Reg.Subset  == sim.SubsetHi:
			cpu.Registers[dest.Reg.Name] =
			cpu.Registers[dest.Reg.Name]&0x00ff |
			cpu.Registers[src.Reg.Name ]
		}
	}

}
