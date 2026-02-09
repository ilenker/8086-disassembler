package main

const (
	//───────────────────────────────────────
	//                 MOV
	//───────────────────────────────────────
    // Reg <-> Reg/Mem
	OP_MOV_RM          OpCode   = 0b1000_1000
    MASK_MOV_RM        OpCode   = 0b1111_1100

    // Imm -> Reg/Mem
	OP_MOV_IRM         OpCode   = 0b1100_0110
    MASK_MOV_IRM       OpCode   = 0b1111_1110
	EXT_MOV_IRM 	   RegOrExt = 0b000

    // Imm -> Reg
	OP_MOV_IR          OpCode   = 0b1011_0000
    MASK_MOV_IR        OpCode   = 0b1111_0000

    // Mem <-> Ax
	OP_MOV_M_TF_Ax     OpCode   = 0b1010_0000
    MASK_MOV_M_TF_Ax   OpCode   = 0b1111_1100

	// r/m <-> sr
	OP_MOV_SR          OpCode   = 0b1000_1100
    MASK_MOV_SR        OpCode   = 0b1111_1101

	//───────────────────────────────────────
	//			  ALU "Subgroup"
	//───────────────────────────────────────
	// [00][ext][IsImmediate][W]
	OP_SUBGROUP_ALU    OpCode   = 0b0000_0000
	MASK_SUBGROUP_ALU  OpCode   = 0b1100_0000
	OP_ALU_IsToAx      byte     = 0b0000_0100

	//──────────────────────────────────────
	//				 GROUP 1
	//	    Data Processing Immediates
	//──────────────────────────────────────
	// [1000_00sw][mod ext rm][disp][data]
	OP_GROUP_1         OpCode   = 0b1000_0000
	MASK_GROUP_1       OpCode   = 0b1111_1100
	EXT_ADD            ALUFunction = 0b000
	EXT_SUB            ALUFunction = 0b101
	EXT_CMP            ALUFunction = 0b111

	// [0000_010w][data]
	OP_ADD_I_Ax        OpCode   = 0b0000_0100
	MASK_ADD_I_Ax      OpCode   = 0b1111_1110


	//───────────────────────────────────────
	//             Short Jumps
	//───────────────────────────────────────

	OP_JE     OpCode = 0b01110100
	OP_JL     OpCode = 0b01111100
	OP_JLE    OpCode = 0b01111110
	OP_JB     OpCode = 0b01110010
	OP_JBE    OpCode = 0b01110110
	OP_JP     OpCode = 0b01111010
	OP_JO     OpCode = 0b01110000
	OP_JS     OpCode = 0b01111000
	OP_JNZ    OpCode = 0b01110101
	OP_JNL    OpCode = 0b01111101
	OP_JNLE   OpCode = 0b01111111
	OP_JNB    OpCode = 0b01110011
	OP_JNBE   OpCode = 0b01110111
	OP_JNP    OpCode = 0b01111011
	OP_JNO    OpCode = 0b01110001
	OP_JNS    OpCode = 0b01111001
	OP_LOOP   OpCode = 0b11100010
	OP_LOOPZ  OpCode = 0b11100001
	OP_LOOPNZ OpCode = 0b11100000
	OP_JCXZ   OpCode = 0b11100011
)

var Jxxx_STRING_MAP = map[OpCode]string{
	// ZF
	OP_JE     :"je",
	OP_JNZ    :"jnz",

	// CF
	OP_JB     :"jb",
	OP_JNB    :"jnb",

	// CF / ZF
	OP_JBE    :"jbe",
	OP_JNBE   :"jnbe",

	OP_JL     :"jl",
	OP_JLE    :"jle",
	OP_JP     :"jp",
	OP_JO     :"jo",
	OP_JS     :"js",
	OP_JNL    :"jnl",
	OP_JNLE   :"jnle",
	OP_JNP    :"jnp",
	OP_JNO    :"jno",
	OP_JNS    :"jns",
	OP_LOOP   :"loop",
	OP_LOOPZ  :"loopz",
	OP_LOOPNZ :"loopnz",
	OP_JCXZ   :"jcxz",
}


func (i *Instruction) StringNoExt() string {
	switch i.opCode {
	case OP_MOV_RM,
		OP_MOV_IR,
		OP_MOV_M_TF_Ax:
		return "mov"
	case OP_SUBGROUP_ALU:
		switch ALUFunction(i.ext) {
		case EXT_ADD:
			return "add"
		case EXT_SUB:
			return "sub"
		case EXT_CMP:
			return "cmp"
		}
	}
	return "!Unknown OpCode!"
}

func (i *Instruction) StringWithExt() string {
	// TODO(Johan):
	switch i.opCode {
	case OP_MOV_IRM:
		if i.reg_ext == EXT_MOV_IRM {
		return "mov"
		}
		return "!Unknown OpCode!"
	case OP_GROUP_1:
		switch ALUFunction(i.reg_ext) {
		case EXT_ADD:
			return "add"
		case EXT_SUB:
			return "sub"
		case EXT_CMP:
			return "cmp"
		}
		return "!Unknown OpCode!"
	}
	return "!Unknown OpCode!"
}


