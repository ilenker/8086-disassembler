package main

import (
	"fmt"
	"github.com/ilenker/8086-disassembler/internal/sim"
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

var iCache = make(map[int]Instruction)
var INST_COUNTER int

// NOTE(Johan): Struct is currently in logical-order, not in efficiently-packed-order
type Instruction struct {
	opCode   OpCode
	sbfs     SingleBitFields
	mode     Mode
	reg_ext  RegOrExt
	reg_mem  RegOrMem
	disp     Displacement
	imm      Immediate
	regIsExt bool
	ext      byte
	category Category
	srOp	 bool
	binIndex int
	srcIndex int
	size	 int

	// "intermediate representation" 
	dest sim.Argument
	src  sim.Argument

	debug struct {
		raw   string
		byte1 string
		byte2 string
		disp1 string
		disp2 string
		data1 string
		data2 string
	}
}

type (
	OpCode          byte
	SingleBitFields struct { S, W, D, V, Z bool }
	Mode            byte
	RegOrExt        byte
	RegOrMem        byte
	Displacement    struct { Value int16; BytesConsumed int }
	Immediate       struct { Value int16; BytesConsumed int }
	Category        int
)

const (
	MemMode_0  Mode = 0b00
	MemMode_8  Mode = 0b01
	MemMode_16 Mode = 0b10
	RegToReg   Mode = 0b11
	ImmToReg_IMPLIED Mode = 0b11111111
)

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


// Start: First Byte
func parseInstruction(idx int, binary []byte, ) (*Instruction, int, error) {
	if inst, ok := iCache[idx]; ok {
		return &inst, 0, nil
	} 

	b := binary[idx]
	inst := &Instruction{}
	inst.binIndex = idx
	defer func() {
		inst.srcIndex = INST_COUNTER
		INST_COUNTER++
		inst.size = idx - inst.binIndex
		iCache[inst.binIndex] = *inst
	}()
	inst.debug.byte1 = fmt.Sprintf("%08b", b)
	// TODO(Johan): Handle NO-OP
	//────────────────────────────────────
	//               MOV
	//────────────────────────────────────
	// MOV	r/m	<->	reg
	// [1000 10dw]	[mod reg r/m]	[disp lo]	[disp hi]
	switch {
	case OpCode(b)&MASK_MOV_RM == OP_MOV_RM:
		inst.opCode = OP_MOV_RM
		inst.sbfs.D = BitIsSet(b, D_MASK)
		inst.sbfs.W = BitIsSet(b, W_MASK_DEFAULT)
		idx++
		idx += getModRegExtRM(idx, binary, inst)
		idx += getDisplacement(idx, binary, inst)
		inst.regIsExt = false
		inst.category = DATA_TRANSFER

		return inst, idx, nil

	// MOV Imm -> Reg Pattern: 1011 wreg
	// [1011 wreg] [data lo] [data hi]
	case OpCode(b)&MASK_MOV_IR == OP_MOV_IR:
		inst.opCode = OP_MOV_IR
		inst.sbfs.W = BitIsSet(b, BIT_3)
		reg := (b)&0b0000_0111 // Rare case of Byte1 using this mask
		inst.reg_mem = RegOrMem(reg)
		inst.mode = ImmToReg_IMPLIED
		idx++
		idx += getImmediate(idx, binary, inst)
		inst.regIsExt = false
		inst.category = DATA_TRANSFER
		// --- IntRep ---
		inst.dest.Type = sim.ArgReg
		inst.dest.Reg.Name = sim.RegName(reg)
		if inst.sbfs.W {
			inst.dest.Reg.Name = sim.RegName(reg)
			inst.dest.Reg.Subset = sim.SubsetX
		} else
		if reg >= byte(AL) && reg < byte(AH) {
			inst.dest.Reg.Name = sim.RegName(reg)
			inst.dest.Reg.Subset = sim.SubsetLo
		} else
		if reg >= byte(AH) && reg < byte(AX) {
			inst.dest.Reg.Name = sim.RegName(reg-4)
			inst.dest.Reg.Subset = sim.SubsetHi
		}

		return inst, idx, nil
	
	// MOV Imm -> Reg/Mem
	// 1100011[w] [mod 0 0 0 r/m] [disp lo] [disp hi] [data] [data if w]
	case OpCode(b)&MASK_MOV_IRM == OP_MOV_IRM:
		inst.opCode = OP_MOV_IRM
		inst.sbfs.W = BitIsSet(b, W_MASK_DEFAULT)
		//inst.sbfs.D = true
		idx++
		idx += getModRegExtRM(idx, binary, inst)
		if inst.reg_ext == EXT_MOV_IRM {
			inst.regIsExt = true
		} else {
			// TODO(Johan): What to do on unlikely event of these bits not being 000?
			fmt.Println(inst.BinaryString())
			panic(cRed("unknown opcode:"))
		}
		idx += getDisplacement(idx, binary, inst)
		idx += getImmediate(idx, binary, inst)
		inst.category = DATA_TRANSFER

		return inst, idx, nil

	// MOV Mem <-> Ax ("Direct address")
	// Pattern: 1010 00dw [addr lo] [addr hi]
	case OpCode(b)&MASK_MOV_M_TF_Ax == OP_MOV_M_TF_Ax:
		inst.opCode = OP_MOV_M_TF_Ax
		inst.mode = 0b00     // Implied "Mem Mode"
		inst.reg_mem = 0b110 // Implied "Direct Address"
		// Flipping this seems to work, but noting
		// for possible future ramifications
		inst.sbfs.D = !BitIsSet(b, D_MASK) 
		inst.sbfs.W = BitIsSet(b, W_MASK_DEFAULT)
		idx++
		idx += getAddress(idx, binary, inst)
		//inst.reg_ext = 0b000
		inst.category = DATA_TRANSFER

		return inst, idx, nil

	// MOV r/m <-> sr (segment register)
	// [1000 11d0] [mod 0sr rm] [disp lo] [disp hi]
	case OpCode(b)&MASK_MOV_SR == OP_MOV_SR:
		inst.opCode = OP_MOV_SR
		inst.sbfs.D = BitIsSet(b, D_MASK)
		idx++
		idx += getModRegExtRM(idx, binary, inst)
		idx += getDisplacement(idx, binary, inst)
		inst.regIsExt = false
		inst.category = DATA_TRANSFER
		inst.srOp = true

		return inst, idx, nil

	//────────────────────────────────────
	//     100000SW ADD / SUB / CMP
	//────────────────────────────────────
	// Pattern: [1000_00sw][mod ext rm][disp l][disp h][data l][data h]
	case OpCode(b)&MASK_GROUP_1 == OP_GROUP_1:
		inst.opCode = OP_GROUP_1
		inst.sbfs.S = BitIsSet(b, EXTRA_MASK_DEFAULT)
		inst.sbfs.W = BitIsSet(b, W_MASK_DEFAULT)
		idx++
		// TODO(Johan): Did not check this thoroughly for IntRep setting,
		// trusting it works but note this case for first
		// check when things go wrong
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
		inst.category = ARITHMETIC

		return inst, idx, nil

	//────────────────────────────────────────────────────────────────────────────
	//	ALU "Subroup": [00 "ext" IsImm D W] [mod reg r/m]? [disp/data][disp/data]
	//────────────────────────────────────────────────────────────────────────────
	case OpCode(b)&MASK_SUBGROUP_ALU == OP_SUBGROUP_ALU:
		// "Extension" is handled slightly differently here
		// because it's not an extension per say, but I'm treating
		// it like one. Main Group [00] -> extended -> [00 000] = ADD
		inst.opCode = OP_SUBGROUP_ALU
		inst.ext = (b&REG_OR_EXT_MASK)>>3
		inst.sbfs.W = BitIsSet(b, W_MASK_DEFAULT)
		if b&OP_ALU_IsToAx != 0 {
			idx++
			idx += getImmediate(idx, binary, inst)
			inst.reg_mem = 0b000 // A
			inst.mode = ImmToReg_IMPLIED
			// IntRep
			inst.dest.Type = sim.ArgReg
			inst.dest.Reg.Name = sim.AX
			if inst.sbfs.W {
				inst.dest.Reg.Subset = sim.SubsetX
			} else {
				// AH is exempt from "to accumulator" instructions
				inst.dest.Reg.Subset = sim.SubsetLo
			}
			inst.category = ARITHMETIC

			return inst, idx, nil
		}
		inst.sbfs.D = BitIsSet(b, D_MASK)
		idx++
		idx += getModRegExtRM(idx, binary, inst)
		idx += getDisplacement(idx, binary, inst)
		inst.category = ARITHMETIC

		return inst, idx, nil

	default:
		_, ok := Jxxx_STRING_MAP[OpCode(b)]
		if ok {
			inst.opCode = OpCode(b)
			inst.sbfs.W = false
			idx++
			inst.disp = Displacement{
				Value: int16(int8(binary[idx])),
				BytesConsumed: 1,
			}
			inst.src.Type = sim.ArgImm
			inst.src.ImmValue = inst.disp.Value
			idx++
			inst.category = CONTROL_TRANSFER

			return inst, idx, nil
		}
		return &Instruction{}, idx+1, nil
	}
}


// Second Byte
func getModRegExtRM(idx int, binary []byte, inst *Instruction) (int) {
	b := binary[idx]
	{
		inst.debug.byte2 = fmt.Sprintf("%08b", b)
	}
	bytesConsumed := 1

	inst.mode = (Mode(b)&MODE_MASK) >> 6
	inst.reg_ext = (RegOrExt(b)&REG_OR_EXT_MASK) >> 3
	inst.reg_mem = RegOrMem(b)&REG_OR_MEM_MASK

	// --- Intermediate Representation ---:
	w := inst.sbfs.W
	destIsReg := inst.sbfs.D

	// Determine RM Reg-or-Mem status
	argRM  := sim.Argument{}
	rm     := byte(inst.reg_mem)
	regext := byte(inst.reg_ext)
	mod    := inst.mode
	// R/M is Reg
	if inst.mode == RegToReg {
		argRM.Type = sim.ArgReg
		if w {
			argRM.Reg.Name = sim.RegName(rm)
			argRM.Reg.Subset = sim.SubsetX
		} else
		if rm >= byte(AL) && rm < byte(AH) {
			argRM.Reg.Name = sim.RegName(rm)
			argRM.Reg.Subset = sim.SubsetLo
		} else
		if rm >= byte(AH) && rm < byte(AX) {
			argRM.Reg.Name = sim.RegName(rm-4)
			argRM.Reg.Subset = sim.SubsetHi
		}
	} else {
		// R/M is Mem
		argRM.Type = sim.ArgMem
		// Can't access Displacement...
		argRM.Mem.Disp = 105
		switch rm {
		case 0b000:
			argRM.Mem.BaseReg  = sim.BX
			argRM.Mem.IndexReg = sim.SI
		case 0b001:
			argRM.Mem.BaseReg = sim.BX
			argRM.Mem.IndexReg = sim.DI
		case 0b010:
			argRM.Mem.BaseReg = sim.BP
			argRM.Mem.IndexReg = sim.SI
		case 0b011:
			argRM.Mem.BaseReg = sim.BP
			argRM.Mem.IndexReg = sim.DI
		case 0b100:
			argRM.Mem.IndexReg = sim.SI
			argRM.Mem.BaseReg = sim.NilX
		case 0b101:
			argRM.Mem.IndexReg = sim.DI
			argRM.Mem.BaseReg = sim.NilX
		case 0b110:
			// Exception
			if mod != 0b00 {
				argRM.Mem.BaseReg = sim.BP
			} else {
				argRM.Mem.BaseReg = sim.NilX
			}
			argRM.Mem.IndexReg = sim.NilX
		case 0b111:
			argRM.Mem.BaseReg = sim.BX
			argRM.Mem.IndexReg = sim.NilX
		}
	}

	// Determine RegExt Reg-or-Ext status
	argReg := sim.Argument{}
	// TODO(Johan): Make sure to add the rest
	// of the relevant groups to this check
	if inst.opCode == OP_GROUP_1 || inst.opCode == OP_MOV_IRM {
		// Reg is Ext
		argReg.Type = sim.ArgNone
	} else {
		// Reg is Reg
		argReg.Type = sim.ArgReg
		if w {
			argReg.Reg.Subset = sim.SubsetX
			argReg.Reg.Name = sim.RegName(regext)
		} else
		if regext >= byte(AL) && regext < byte(AH) {
			argReg.Reg.Name = sim.RegName(regext)
			argReg.Reg.Subset = sim.SubsetLo
		} else
		if regext >= byte(AH) && regext < byte(AX) {
			argReg.Reg.Name = sim.RegName(regext-4)
			argReg.Reg.Subset = sim.SubsetHi
		}
	}

	// Check Direction
	if destIsReg {
		inst.dest = argReg
		inst.src  = argRM
	} else {
		inst.dest = argRM
		inst.src  = argReg
	}

	return bytesConsumed
}


// Disp bytes
func getDisplacement(idx int, binary []byte, inst *Instruction) (int) {
	mode := inst.mode
	rm := inst.reg_mem

	var dispValue int16
	bytesConsumed := 0

	switch mode {
	case MemMode_0:
		// Direct Address (2x disp bytes, 16bit address)
		if rm == 0b110 {
			dispValue = int16(binary[idx]) | (int16(binary[idx+1]) << 8)
			bytesConsumed = 2
			{
				inst.debug.disp1 = fmt.Sprintf("%08b", binary[idx])
				inst.debug.disp2 = fmt.Sprintf("%08b", binary[idx+1])
			}
		}
	case MemMode_8:
		dispValue = int16(int8(binary[idx]))
		{
			inst.debug.disp1 = fmt.Sprintf("%08b", binary[idx])
		}
		bytesConsumed = 1

	case MemMode_16:
		dispValue = int16(binary[idx]) | (int16(binary[idx+1]) << 8)
		bytesConsumed = 2
		{
			inst.debug.disp1 = fmt.Sprintf("%08b", binary[idx])
			inst.debug.disp2 = fmt.Sprintf("%08b", binary[idx+1])
		}
	}

	inst.disp = Displacement{
		Value: dispValue,
		BytesConsumed: bytesConsumed,
	}

	// --- Intermediate Representation ---:
	if inst.dest.Type == sim.ArgMem {
		inst.dest.Mem.Disp = dispValue
	}
	if inst.src.Type == sim.ArgMem {
		inst.src.Mem.Disp = dispValue
	}

	return bytesConsumed
}


// Data bytes
func getImmediate(idx int, binary []byte, inst *Instruction) (int) {
	var value int16
	var bytesConsumed int
	word := inst.sbfs.W
	signExtend := inst.sbfs.S

	// Set the intermediate representation after immediate fields are set
	defer func() {
		if inst.src.Type != sim.ArgNone {
			inst.printStruct()
			panic("Assertion failed(inst.src.Type == sim.ArgNone):\n" +
				"Source field set on immediate instruction before parsing of immediate bytes")
		}
		inst.src.Type = sim.ArgImm
		inst.src.ImmValue = inst.imm.Value

		if bytesConsumed == 1 {
			inst.debug.data1 = fmt.Sprintf("%08b", binary[idx])
		} else {
			inst.debug.data1 = fmt.Sprintf("%08b", binary[idx])
			inst.debug.data2 = fmt.Sprintf("%08b", binary[idx+1])
		}
	}()

	switch {
	case !signExtend && !word:  
		value = int16(int8(binary[idx]))
		bytesConsumed = 1

	case !signExtend && word:  
		value = int16(binary[idx]) | (int16(binary[idx+1]) << 8)
		bytesConsumed = 2

	case signExtend && word:
		val := binary[idx]
		if (val & BIT_7) != 0 {
			value = int16(uint16(val) | 0b11111111_00000000)
		} else {
			value = int16(int8(val))
		}
		bytesConsumed = 1

	case signExtend && !word:
		panic(
			fmt.Sprintf("invalid case in binary: s:w == 0b10\nFrom idx(%d-1): -1:[%08b], datalo:[%08b]",
				idx, binary[idx-1], binary[idx]),
			)
	}

	inst.imm = Immediate{
		Value: value,
		BytesConsumed: bytesConsumed,
	}

	return bytesConsumed
}

func getAddress(idx int, binary []byte, inst *Instruction) (int) {
	bytesConsumed := 2
	w := inst.sbfs.W
	destIsMem := inst.sbfs.D
	inst.disp = Displacement{
		Value: int16(binary[idx]) | (int16(binary[idx+1]) << 8),
		BytesConsumed: bytesConsumed,
	}

	// --- Intermediate Representation ---:
	if !destIsMem {
		inst.dest.Type = sim.ArgMem
		inst.dest.Mem.Disp = inst.disp.Value
		inst.dest.Mem.BaseReg = sim.NilX
		inst.dest.Mem.IndexReg = sim.NilX
		inst.src.Type = sim.ArgReg
		if w {
			inst.src.Reg.Subset = sim.SubsetX
		} else {
			// AH is exempt from "to accumulator" instructions
			inst.src.Reg.Subset = sim.SubsetLo
		}
	} else {
		inst.dest.Type = sim.ArgReg
		if w {
			inst.dest.Reg.Subset = sim.SubsetX
		} else {
			// AH is exempt from "to accumulator" instructions
			inst.dest.Reg.Subset = sim.SubsetLo
		}
		inst.src.Type = sim.ArgMem
		inst.src.Mem.Disp = inst.disp.Value
		inst.src.Mem.BaseReg = sim.NilX
		inst.src.Mem.IndexReg = sim.NilX
	}

	{
		inst.debug.disp1 = fmt.Sprintf("(addr)%08b", binary[idx])
		inst.debug.disp2 = fmt.Sprintf("(addr)%08b", binary[idx+1])
	}
	return bytesConsumed
}

//━━━━━━━━━━━━━━
// Mask Helpers
//━━━━━━━━━━━━━━

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


//━━━━━━━━━━━━━━━━━━
// String Functions
//━━━━━━━━━━━━━━━━━━

func (r RegOrExt) StringWithW(word bool) string {
	if !word {
		return WReg(r).String()
	}
	return WReg(r | 0b1000).String()
}

func (r RegOrMem) StringWithW(word bool) string {
	if !word {
		return WReg(r).String()
	}
	return WReg(r | 0b1000).String()
}

func (i *Instruction) BinaryString() string {
	var (
		byte1,
		byte2,
		disp1,
		disp2,
		data1,
		data2 string
	)

	cbyte1 := "[48;2;32;18;77m"
	cbyte2 := "[48;2;9;71;41m"

	cdisp1 := "[48;2;1;63;64m"
	cdisp2 := "[48;2;7;81;82m"

	cdata1 := "[48;2;67;60;30m"
	cdata2 := "[48;2;86;77;70m"

	cmod   := "[38;2;30;255;162m"
	creg   := "[38;2;255;211;87m"
	crm    := "[38;2;91;126;129m"
	cr := "[49m[39m"

	if i.debug.byte1 == "" {
		byte1 = "NO OPCODE"
	} else {
		byte1 = i.debug.byte1
	}

	if i.debug.byte2 == "" {
		byte2 = "NO MOD RM"
	} else {
		byte2 = (
		cmod + i.debug.byte2[0:2] + " " +
		creg + i.debug.byte2[2:5] + " " +
		crm  + i.debug.byte2[5: ] + cr)
	}

	disp1 = i.debug.disp1
	disp2 = i.debug.disp2
	data1 = i.debug.data1
	data2 = i.debug.data2

	return fmt.Sprintln(
		cbyte1,
		byte1,
		cbyte2,
		byte2,
		cdisp1,
		disp1,
		cdisp2,
		disp2,
		cdata1,
		data1,
		cdata2,
		data2,
		cr,
		)
}

func (m Mode) String() string {
	switch m {
	case MemMode_0:  return "00 (Mem 0 byte disp)"
	case MemMode_8:  return "01 (Mem 1 byte disp)"
	case MemMode_16: return "10 (Mem 2 byte disp)"
	case RegToReg:   return "11 (Reg->Reg)"
	case ImmToReg_IMPLIED:   return "xxxx(Imm->Reg Implied)"
	}
	return "unknown mode"
}
