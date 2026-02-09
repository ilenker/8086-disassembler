package main

import (
	"fmt"
)

func (i *Instruction) printStruct() {
	fmt.Printf("opCode:        %08b\n", i.opCode)
	fmt.Printf("Mnemonic:      %v\n", i.StringNoExt())
	fmt.Printf("   (ext):      %v\n", i.StringWithExt())
	fmt.Printf("sbfs:          S:%s W:%s D:%s V:%s Z:%s\n",
		bool2Str(i.sbfs.S),
		bool2Str(i.sbfs.W),
		bool2Str(i.sbfs.D),
		bool2Str(i.sbfs.V),
		bool2Str(i.sbfs.Z),
		)
	fmt.Printf("mode:          %v\n", i.mode)
	fmt.Printf("reg_ext:       %03b\n", i.reg_ext)
	fmt.Printf("reg_mem:       %03b\n", i.reg_mem)
	fmt.Printf("disp:          %v\n", i.disp)
	fmt.Printf("imm:           %v\n", i.imm)
	fmt.Printf("regIsExt:      %v\n", i.regIsExt)
	fmt.Printf("ext:           %v\n", i.ext)
	fmt.Printf("category:      %v\n", i.category.String())
	fmt.Printf("binIndex:      %#02x\n", i.binIndex)

	fmt.Println("  --- IntRep ---  ")
	fmt.Printf("src:           %v\n", i.src)
	fmt.Printf("dest:          %v\n", i.dest)
}

func printBinary(binary []byte) {
	fmt.Println("Binary input: ")
	for i := range binary {
		fmt.Printf("%08b ", binary[i])
		if i % 4 == 3 {
			fmt.Println()
		}
	}
	fmt.Println("\n_________________________")
}

func bool2Str(b bool) string {
    if b {
        return "1"
    } else {
        return "_"
    }
}

