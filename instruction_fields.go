package main

type WReg byte
const (
	AL WReg = iota 
	CL
	DL
	BL
	AH
	CH
	DH
	BH
	AX
	CX
	DX
	BX
	SP // Stack pointer
	BP // Base pointer
	SI // Source index
	DI // Destination index
)

func (value WReg) String() string {
    return [...]string{
	"al",
	"cl",
	"dl",
	"bl",
	"ah",
	"ch",
	"dh",
	"bh",
	"ax",
	"cx",
	"dx",
	"bx",
	"sp",
	"bp",
	"si",
	"di",
	}[value]
}

type ALUFunction byte

func (a ALUFunction) String() string {
	switch a {
	case EXT_ADD:
	return "EXT_ADD"
	case EXT_SUB:
	return "EXT_SUB"
	case EXT_CMP:
	return "EXT_CMP"
	}
	return "Unknown extension"
}
