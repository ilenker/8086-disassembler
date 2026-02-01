package main

const (
    // MOV Reg <-> Reg/Mem
    // Pattern: 1000 10dw
	OP_MOV_RM   OpCode = 0b1000_1000
    MASK_MOV_RM OpCode = 0b1111_1100

    // MOV Imm -> Reg/Mem
    // Pattern: 1100 011w
	OP_MOV_IRM   OpCode = 0b1100_0110
    MASK_MOV_IRM OpCode = 0b1111_1110

    // MOV Imm -> Reg
    // Pattern: 1011 wreg
	OP_MOV_IR   OpCode = 0b1011_0000
    MASK_MOV_IR OpCode = 0b1111_0000

    // MOV Mem -> Ax
    // Pattern: 1010 000w
	OP_MOV_M_Ax   OpCode = 0b1010_0000
    MASK_MOV_M_Ax OpCode = 0b1111_1110

    // MOV Ax -> Mem
    // Pattern: 1010 001w
	OP_MOV_Ax_M   OpCode = 0b1010_0010
    MASK_MOV_Ax_M OpCode = 0b1111_1110
)

func (o OpCode) StringNoExt() string {
	switch o {
	case OP_MOV_RM,
		OP_MOV_IRM,
		OP_MOV_IR,
		OP_MOV_M_Ax,
		OP_MOV_Ax_M:
		return "mov"
	}
	return "!UNKNOWN OPCODE!"
}
