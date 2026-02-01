package main

import (
	"fmt"
)

// Notes:
// S V and Z are mutually exclusive.
// They might always appear at the same spot, unsure.
//
// If S is present, W must be present.
//
// Byte Consumption:
// Opcode Bytes (+1..2)
// ModRM byte   (+1   )
// Displacement (+0..2)
// Immediates   (+0..2)

type Instruction struct {
	opCode   OpCode
	sbfs     SingleBitFields
	mode     Mode
	reg_ext  RegisterOrExtension
	reg_mem  RegisterOrMemory
	disp     Displacement
	imm      Immediate
	regIsExt bool
	regOnly  bool
}

type (
	OpCode				byte

	SingleBitFields struct {
		S, W, D, V, Z bool
	}

	Mode				byte

	RegisterOrExtension byte

	RegisterOrMemory	byte

	Displacement struct {
		Value int16
		BytesConsumed int
	}

	Immediate struct {
		Value int16
		BytesConsumed int
	}
)


// Start: First Byte
func parseInstruction(idx int, binary []byte, ) (*Instruction, int, error) {
	b := binary[idx]
	inst := &Instruction{}
	switch {
	case OpCode(b)&MASK_MOV_RM == OP_MOV_RM:
		inst.opCode = OP_MOV_RM
		inst.sbfs.D = BitIsSet(b, D_MASK)
		inst.sbfs.W = BitIsSet(b, W_MASK_DEFAULT)
		idx++
		idx += getModExtRM(idx, binary, inst)
		idx += getDisplacement(idx, binary, inst)
		inst.regIsExt = false
		return inst, idx, nil

	//MOV Imm -> Reg Pattern: 1011 wreg
	case OpCode(b)&MASK_MOV_IR == OP_MOV_IR:
		inst.opCode = OP_MOV_IR
		inst.sbfs.W = BitIsSet(b, BIT_3)
		inst.reg_ext = RegisterOrExtension(b)&0b0000_0111
		idx++
		idx += getImmediate(idx, binary, inst)
		inst.regIsExt = false
		inst.regOnly = true
		return inst, idx, nil

	default:
		return &Instruction{}, idx+1, nil
	}
}


// Second Byte
func getModExtRM(idx int, binary []byte, inst *Instruction) (int) {
	b := binary[idx]
	inst.mode = (Mode(b)&MODE_MASK) >> 6
	inst.reg_ext = (RegisterOrExtension(b)&REG_OR_EXT_MASK) >> 3
	inst.reg_mem = RegisterOrMemory(b)&REG_OR_MEM_MASK
	return 1
}


// Disp bytes
func getDisplacement(idx int, binary []byte, inst *Instruction) (int) {
	mode := inst.mode
	rm := inst.reg_mem

	var dispValue int16
	byteCount := 0

	// TODO(): 
	switch mode {
	case MemMode_0:
		// Direct Address (2x disp bytes, 16bit address)
		if rm == 0b110 {
			dispValue = int16(binary[idx]) | (int16(binary[idx+1]) << 8)
			byteCount = 2
		}
	case MemMode_8:
		dispValue = int16(int8(binary[idx]))
		byteCount = 1

	case MemMode_16:
		dispValue = int16(binary[idx]) | (int16(binary[idx+1]) << 8)
		byteCount = 2
	}

	inst.disp = Displacement{
		Value: dispValue,
		BytesConsumed: byteCount,
	}
	return byteCount
}


// Data bytes
func getImmediate(idx int, binary []byte, inst *Instruction) (int) {
	word := inst.sbfs.W
	signExtend := inst.sbfs.S

	switch {
	case !signExtend && !word:  
		inst.imm = Immediate{
			Value: int16(int8(binary[idx])),
			BytesConsumed: 1,
		}
		return 1

	case !signExtend && word:  
		inst.imm = Immediate{
			Value: int16(binary[idx]) | (int16(binary[idx+1]) << 8),
			BytesConsumed: 2,
		}
		return 2

	case signExtend && word:
		val := binary[idx]
		var valExtended int16
		if (val & BIT_7) != 0 {
			valExtended = int16(val | 0b11111111)
		} else {
			valExtended = int16(int8(val))
		}
		inst.imm = Immediate{
			Value: valExtended,
			BytesConsumed: 1,
		}
		return 1

	case signExtend && !word:
		panic(
			fmt.Sprintf("invalid case in binary: s:w == 0b10\nFrom idx(%d-1): -1:[%08b], datalo:[%08b]",
				idx, binary[idx-1], binary[idx]),
			)
	}

	return idx
}

const (
	MemMode_0   Mode = 0b00
	MemMode_8   Mode = 0b01
	MemMode_16  Mode = 0b10
	RegToReg    Mode = 0b11
)

//////////////////
// Mask Helpers
//////////////////
const (
	MODE_MASK       = 0b11000000
	REG_OR_EXT_MASK = 0b00111000
	REG_OR_MEM_MASK = 0b00000111
	D_MASK          = 0b00000010 // Have not seen an exception

	// Many exceptions here but there is a trend
	W_MASK_DEFAULT     = 0b00000001
	// Only exception I see is "REP" which has "z" at bit 0
	EXTRA_MASK_DEFAULT = 0b00000010 

	// Generic Bit Addressing Helpers:
	BIT_7 = 0b10000000
	BIT_6 = 0b01000000
	BIT_5 = 0b00100000
	BIT_4 = 0b00010000
	BIT_3 = 0b00001000
	BIT_2 = 0b00000100
	BIT_1 = 0b00000010
	BIT_0 = 0b00000001
)

func BitIsSet(b, mask byte) bool {
	return b&mask == mask
}



/////////////////////
// String Functions
/////////////////////

func (r RegisterOrExtension) StringWithW(word bool) string {
	if !word {
		return WReg(r).String()
	}
	return WReg(r | 0b1000).String()
}

func (r RegisterOrMemory) StringWithW(word bool) string {
	if !word {
		return WReg(r).String()
	}
	return WReg(r | 0b1000).String()
}

func (o OpCode) StringWithExt(ext RegisterOrExtension) string {
	// TODO(Johan):
	return ""
}
