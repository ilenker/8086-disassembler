; ========================================================================
;
; (C) Copyright 2023 by Molly Rocket, Inc., All Rights Reserved.
;
; This software is provided 'as-is', without any express or implied
; warranty. In no event will the authors be held liable for any damages
; arising from the use of this software.
;
; Please see https://computerenhance.com for further information
;
; ========================================================================

; ========================================================================
; LISTING 40
; ========================================================================

bits 16

; Signed displacements
; RM_TF_R 
mov ax, [bx + di - 37]
mov [si - 300], cx
mov dx, [bx - 32]

; Explicit sizes (Immediate to register/memory?) 1100011[w]   mod 0 0 0 r/m   [disp lo]  [disp hi]  [data] [data if w]
mov [bp + di], byte 7
mov [di + 901], word 347

; Direct address
; RM_TF_R 
mov bp, [5]
mov bx, [3458]

; Memory-to-accumulator test
; M_T_Ax
mov ax, [2555]
mov ax, [16]

; Accumulator-to-memory test
; Ax_T_M
mov [2554], ax
mov [15], ax
