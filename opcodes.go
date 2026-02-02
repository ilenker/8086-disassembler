package main

const (
	//────────────────────────────────────
	//               MOV
	//────────────────────────────────────
    // MOV Reg <-> Reg/Mem
    // Pattern: 1000 10dw
	OP_MOV_RM   OpCode = 0b1000_1000
    MASK_MOV_RM OpCode = 0b1111_1100

    // MOV Imm -> Reg/Mem
    // Pattern: 1100 011w
	OP_MOV_IRM   OpCode = 0b1100_0110
    MASK_MOV_IRM OpCode = 0b1111_1110
	EXT_MOV_IRM  RegisterOrExtension = 0b000

    // MOV Imm -> Reg
    // Pattern: 1011 wreg
	OP_MOV_IR   OpCode = 0b1011_0000
    MASK_MOV_IR OpCode = 0b1111_0000

    // MOV Mem -> Ax
    // Pattern: 1010 00dw
	OP_MOV_M_TF_Ax   OpCode = 0b1010_0000
    MASK_MOV_M_TF_Ax OpCode = 0b1111_1100

	//────────────────────────────────────
	// ALU "Subgroup": [00][ext][IsImmediate][W]
	//────────────────────────────────────
	OP_SUBGROUP_ALU    OpCode = 0b0000_0000
	MASK_SUBGROUP_ALU  OpCode = 0b1100_0000
	OP_ALU_IsToAx      byte   = 0b0000_0100

	//──────────────────────────────────────
	// GROUP 1 - Data Processing Immediates
	//──────────────────────────────────────
	// Pattern: [1000_00sw][mod ext rm][disp l][disp h][data l][data h]

	OP_GROUP_1 OpCode   = 0b1000_0000
	MASK_GROUP_1 OpCode = 0b1111_1100
	EXT_ADD ALUFunction = 0b000
	EXT_SUB ALUFunction = 0b101
	EXT_CMP ALUFunction = 0b111

	// Pattern: [0000_010w][data l][data h]
	OP_ADD_I_Ax OpCode   = 0b0000_0100
	MASK_ADD_I_Ax OpCode = 0b1111_1110
)

func (i *Instruction) StringNoExt() string {
	switch i.opCode {
	case OP_MOV_RM,
		OP_MOV_IR,
		OP_MOV_M_TF_Ax:
		return "mov"
	case OP_SUBGROUP_ALU:
		switch ALUFunction(i.ext) {
		case add:
			return "add"
		case sub:
			return "sub"
		case cmp:
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
	case OP_GROUP_1:
		switch ALUFunction(i.reg_ext) {
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
