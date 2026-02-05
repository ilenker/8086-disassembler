#!/bin/sh

c2="[38;5;208m"
c1="[38;5;103m"
cr="[39m"

args=(--color -y)

header() {
	printf "%-79s %s\n" "${c2}[Reference]${cr}" "${c2}[Output]${cr}"
}

header
go run . binaries/37_single_register_mov > simOut.asm
diff "${args[@]}" asm/no_comments/37_single_register_mov.asm simOut.asm
	printf "${c2}λ> Next?${cr}\n"
	read reply

header
go run . binaries/38_many_register_mov > simOut.asm
diff "${args[@]}" asm/no_comments/38_many_register_mov.asm simOut.asm
	printf "${c2}λ> Next?${cr}\n"
	read reply

header
go run . binaries/39_more_movs > simOut.asm
diff "${args[@]}" asm/no_comments/39_more_movs.asm simOut.asm
	printf "${c2}λ> Next?${cr}\n"
	read reply

header
go run . binaries/40_challenge_movs > simOut.asm
diff "${args[@]}" asm/no_comments/40_challenge_movs.asm simOut.asm
	printf "${c2}λ> Next?${cr}\n"
	read reply

header
go run . binaries/41_add_sub_cmp_jnz > simOut.asm
diff "${args[@]}" asm/no_comments/41_add_sub_cmp_jnz.asm simOut.asm
	printf "${c2}λ> Next?${cr}\n"
	read reply
