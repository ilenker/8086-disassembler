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
// Mod + r/m is one unit, there's never an r/m without mod
// and vice versa.
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
	ext      byte
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
	//────────────────────────────────────
	//               MOV
	//────────────────────────────────────
	switch {
	case OpCode(b)&MASK_MOV_RM == OP_MOV_RM:
		inst.opCode = OP_MOV_RM
		inst.sbfs.D = BitIsSet(b, D_MASK)
		inst.sbfs.W = BitIsSet(b, W_MASK_DEFAULT)
		idx++
		idx += getModRegExtRM(idx, binary, inst)
		idx += getDisplacement(idx, binary, inst)
		inst.regIsExt = false
		return inst, idx, nil

	// MOV Imm -> Reg Pattern: 1011 wreg
	case OpCode(b)&MASK_MOV_IR == OP_MOV_IR:
		inst.opCode = OP_MOV_IR
		inst.sbfs.W = BitIsSet(b, BIT_3)
		inst.reg_ext = RegisterOrExtension(b)&0b0000_0111
		idx++
		idx += getImmediate(idx, binary, inst)
		inst.regIsExt = false
		inst.regOnly = true

		return inst, idx, nil
	
	// MOV Imm -> Reg/Mem
	// 1100011[w] [mod 0 0 0 r/m] [disp lo] [disp hi] [data] [data if w]
	case OpCode(b)&MASK_MOV_IRM == OP_MOV_IRM:
		inst.opCode = OP_MOV_IRM
		inst.sbfs.W = BitIsSet(b, W_MASK_DEFAULT)
		inst.sbfs.D = true
		idx++
		idx += getModRegExtRM(idx, binary, inst)
		if inst.reg_ext == EXT_MOV_IRM {
			inst.regIsExt = true
		} else {
			// TODO(Johan): What to do on unlikely event of these bits not being 000?
			// Just carry on for now
		}
		idx += getDisplacement(idx, binary, inst)
		idx += getImmediate(idx, binary, inst)
		return inst, idx, nil

	// MOV Mem <-> Ax
	// Pattern: 1010 00dw [addr lo] [addr hi]
	case OpCode(b)&MASK_MOV_M_TF_Ax == OP_MOV_M_TF_Ax:
		inst.opCode = OP_MOV_M_TF_Ax
		// Flipping this seems to work, but noting
		// for possible future ramifications
		inst.sbfs.D = !BitIsSet(b, D_MASK) 
		inst.sbfs.W = BitIsSet(b, W_MASK_DEFAULT)
		idx++
		idx += getAddress(idx, binary, inst)
		//inst.reg_ext = 0b000
		inst.mode = 0b00
		inst.reg_mem = 0b110
		return inst, idx, nil


	//────────────────────────────────────
	//     100000SW ADD / SUB / CMP
	//────────────────────────────────────
	// Pattern: [1000_00sw][mod ext rm][disp l][disp h][data l][data h]
	case OpCode(b)&MASK_GROUP_1 == OP_GROUP_1:
		inst.opCode = OP_GROUP_1
		inst.sbfs.S = BitIsSet(b, EXTRA_MASK_DEFAULT)
		inst.sbfs.W = BitIsSet(b, W_MASK_DEFAULT)
		inst.sbfs.D = true
		idx++
		idx += getModRegExtRM(idx, binary, inst)
		switch ALUFunction(inst.reg_ext) {
		case EXT_ADD,
		     EXT_SUB,
			 EXT_CMP:
			inst.regIsExt = true
			inst.ext = byte(inst.reg_ext)
		}
		idx += getDisplacement(idx, binary, inst)
		idx += getImmediate(idx, binary, inst)
		return inst, idx, nil

	//────────────────────────────────────
	//	ALU "Subroup": [00 "ext" IsImm D W] [mod reg r/m]? [disp/data][disp/data]
	//────────────────────────────────────
	// Pattern: [1000_00sw][mod ext rm][disp l][disp h][data l][data h]
	case OpCode(b)&MASK_SUBGROUP_ALU == OP_SUBGROUP_ALU:
		// "Extension" is handled slightly differently here
		// because it's not an extension per say, but I'm treating
		// it like one. Main Group [00] -> extended -> [00 000] = ADD
		inst.opCode = OP_SUBGROUP_ALU
		inst.ext = b&REG_OR_EXT_MASK
		inst.sbfs.W = BitIsSet(b, W_MASK_DEFAULT)
		if b&OP_ALU_IsToAx != 0 {
			idx++
			idx += getImmediate(idx, binary, inst)
			inst.reg_mem = 0b000 // A
			inst.mode = 0b11
			return inst, idx, nil
		}
		inst.sbfs.D = BitIsSet(b, D_MASK)
		idx++
		idx += getModRegExtRM(idx, binary, inst)
		idx += getDisplacement(idx, binary, inst)
		return inst, idx, nil

	default:
		return &Instruction{}, idx+1, nil
	}
}


// Second Byte
func getModRegExtRM(idx int, binary []byte, inst *Instruction) (int) {
	b := binary[idx]
	inst.mode = (Mode(b)&MODE_MASK) >> 6
	if inst.mode == RegToReg {
		inst.regOnly = true
	}
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

func getAddress(idx int, binary []byte, inst *Instruction) (int) {
	w := inst.sbfs.W
	if w {
		inst.disp = Displacement{
			Value: int16(binary[idx]) | (int16(binary[idx+1]) << 8),
			BytesConsumed: 2,
		}
		return 2
	}
	inst.disp = Displacement{
		Value: int16(int8(binary[idx])),
		BytesConsumed: 1,
	}
	return 1
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
