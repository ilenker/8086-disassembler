# 8086 Simulator (work in progress)

This project is part of Casey Muratori's excellent Performance-Aware Programming series, check it out! https://www.computerenhance.com/p/table-of-contents

## How to use
##### 1. Write some .asm:
```asm
; example.asm
bits 16
	
mov cx, 42
plus_10:
	add ax, 10
	loop plus_10
```
##### 2. Assemble using nasm:
`nasm example.asm` -> `example` binary file
##### 3. Run the simulator:
`go run . example -exec` -> prints out list of instructions executed

### Options
```
-exec
	Simulate the instructions
	
-bin, --show-binary
	Print formatted representation of each instruction's source binary
	
-struct, --show-struct
	Print instruction information
	
-reg, --show-cpu
	Print registers and flags after each instruction
	
-mem <address> [n-bytes] [n-columns], --show-mem
	Starting from address, print n-bytes grouped by n-columns as hex.
	
-stats
	Print stats
	
-p<mr>
	**p**oison **m**emory / **r**egisters: seed memory/registers with junk
```


### Instruction progress

| Category          | Total | #Parsing | #Simming |
|-------------------|-------|----------|----------|
| Data Transfer     | 14    | 5        | 4        |
| Arithmetic        | 20    | 3        | 3        |
| Logic             | 12    | 0        | 0        |
| String Manip.     | 5     | 0        | 0        |
| Control Transfer  | 28    | 20       | 3        |
| Processor Control | 11    | 0        | 0        |

The goal is not to build a fully compliant 8086 emulator necessarily, but rather to build enough to understand how CPUs work (instruction encoding, registers, memory etc.) and most importantly, how to read assembly.
