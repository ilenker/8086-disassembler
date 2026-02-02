package main

const (
	OP_JE     OpCode = 0b01110100
	OP_JL     OpCode = 0b01111100
	OP_JLE    OpCode = 0b01111110
	OP_JB     OpCode = 0b01110010
	OP_JBE    OpCode = 0b01110110
	OP_JP     OpCode = 0b01111010
	OP_JO     OpCode = 0b01110000
	OP_JS     OpCode = 0b01111000
	OP_JNE    OpCode = 0b01110101
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
	OP_JE     :"je",
	OP_JL     :"jl",
	OP_JLE    :"jle",
	OP_JB     :"jb",
	OP_JBE    :"jbe",
	OP_JP     :"jp",
	OP_JO     :"jo",
	OP_JS     :"js",
	OP_JNE    :"jne",
	OP_JNL    :"jnl",
	OP_JNLE   :"jnle",
	OP_JNB    :"jnb",
	OP_JNBE   :"jnbe",
	OP_JNP    :"jnp",
	OP_JNO    :"jno",
	OP_JNS    :"jns",
	OP_LOOP   :"loop",
	OP_LOOPZ  :"loopz",
	OP_LOOPNZ :"loopnz",
	OP_JCXZ   :"jcxz",
}
