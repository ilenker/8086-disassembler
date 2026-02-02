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
const (
	add ALUFunction = 0b00_000_000	
	or  ALUFunction = 0b00_001_000	
	adc ALUFunction = 0b00_010_000	
	sbb ALUFunction = 0b00_011_000	
	and ALUFunction = 0b00_100_000	
	sub ALUFunction = 0b00_101_000	
	xor ALUFunction = 0b00_110_000	
	cmp ALUFunction = 0b00_111_000	
)
