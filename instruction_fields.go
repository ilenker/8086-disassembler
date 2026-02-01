package main

type WReg uint8
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


