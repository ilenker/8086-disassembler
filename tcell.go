package main

import (
	"fmt"
	"time"
	"strings"
	"github.com/gdamore/tcell/v3/color"
	tc "github.com/gdamore/tcell/v3"
	"github.com/ilenker/8086-disassembler/internal/sim"
)

type srcLine struct {
	line string
	idx  int
}

var sourceMap = make(map[int]srcLine)

var scr tc.Screen
var stdin chan tc.Event

func (ctx *CTX) setup() {
	ctx.args.getArgs()
	if ctx.args.tui {
		scr, _ = tc.NewScreen()
		scr.Init()
		stdin = scr.EventQ()
	}
}

func (ctx *CTX) drawCPU() {
	if !ctx.args.tui {
		fmt.Printf("Flags:[%s]\n", ctx.cpu.Flags.String())
		fmt.Println(ctx.cpu.String())
		return
	}
	reg := ctx.cpu.Registers
	x := 3
	y := 4
	w := 20
	h := 13
	st := tc.StyleDefault.Foreground(color.Cornsilk)
	st2 := tc.StyleDefault.Foreground(color.DimGrey)
	box(x, y, w, h)
	scr.PutStrStyled(x-1, y-1, "Registers", st.Foreground(color.AntiqueWhite))

	scr.PutStrStyled(x, y,   fmt.Sprintf("Ax:[%04x]%8d", reg[sim.AX], reg[sim.AX]), st.Foreground(color.MediumVioletRed))
	scr.PutStrStyled(x, y+1, fmt.Sprintf("Bx:[%04x]%8d", reg[sim.BX], reg[sim.BX]), st.Foreground(color.Chocolate))
	scr.PutStrStyled(x, y+2, fmt.Sprintf("Cx:[%04x]%8d", reg[sim.CX], reg[sim.CX]), st.Foreground(color.Bisque))
	scr.PutStrStyled(x, y+3, fmt.Sprintf("Dx:[%04x]%8d", reg[sim.DX], reg[sim.DX]), st.Foreground(color.LimeGreen))

	scr.PutStrStyled(x, y+4, "   ---   ", st)

	scr.PutStrStyled(x, y+5, fmt.Sprintf("SP:[%04x]%8d", reg[sim.SP], reg[sim.SP]), st2)
	scr.PutStrStyled(x, y+6, fmt.Sprintf("BP:[%04x]%8d", reg[sim.BP], reg[sim.BP]), st2)
	scr.PutStrStyled(x, y+7, fmt.Sprintf("SI:[%04x]%8d", reg[sim.SI], reg[sim.SI]), st2)
	scr.PutStrStyled(x, y+8, fmt.Sprintf("DI:[%04x]%8d", reg[sim.DI], reg[sim.DI]), st2)

	scr.PutStrStyled(x, y+10, "Flags: " + ctx.cpu.Flags.String(), st2)
}

func (ctx *CTX) drawBinary() {
	if !ctx.args.tui {
		fmt.Printf(ctx.inst.BinaryString())
		return
	}
}

func (ctx *CTX) drawStruct() {
	if !ctx.args.tui {
		ctx.inst.printStruct()
		return
	}
}

func (ctx *CTX) drawASMLine() {
	if !ctx.args.tui {
		fmt.Println(ctx.inst.renderAsASM86())
		return
	}
	x := 26
	scr.PutStr(x, 4+ctx.inst.srcIndex, ctx.inst.renderAsASM86())
	for i := range 50 {
		scr.PutStr(x-1, 4+i, " ")
	}
	scr.PutStr(x-1, 4+ctx.inst.srcIndex, ">")
}

func (ctx *CTX) drawMemoryRegionRGBA(offset, posx, posy, w, h int) {
	box(posx, posy, w, h/2)
	mem := &ctx.cpu.Memory

	addr := offset
	for y := 0; y < h/2; y++ {
		for x := 0; x < w; x++ {
			rt := int32(mem[addr])
			gt := int32(mem[addr+1])
			bt := int32(mem[addr+2])
			_ =  mem[addr+3]

			rb := int32(mem[(w*4)+addr])
			gb := int32(mem[(w*4)+addr+1])
			bb := int32(mem[(w*4)+addr+2])

			ct := tc.NewRGBColor(rt, gt, bt)
			cb := tc.NewRGBColor(rb, gb, bb)

			scr.SetContent(
				posx+x, posy+y, '▀', nil,
				tc.StyleDefault.Foreground(ct).Background(cb))
			addr += 4
		}
		addr += w*4
	}
}

func box(x, y, w, h int) {
	x--;
	top := "┌" + strings.Repeat("─", w) + "┐"
	lin := "│" + strings.Repeat(" ", w) + "│"
	bot := "└" + strings.Repeat("─", w) + "┘"
	// Top/Bot
	st := tc.StyleDefault.Foreground(color.Gray)
	scr.PutStrStyled(x, y-1, top, st)

	for i := range h {
		scr.PutStrStyled(x, y+i, lin, st)
	}
	scr.PutStrStyled(x, y+h, bot, st)
}

func tcEnd() {
	scr.Fini()
}

func (ctx *CTX) getInput() bool {
	select {
	case ev := <-stdin:
		if key, ok := ev.(*tc.EventKey); ok {
			switch key.Str(){
			case "q":
				return false
			case "a":
				ctx.t.Reset(time.Microsecond * 100000)
			case "r":
				ctx.t.Reset(time.Microsecond * 50000)
			case "s":
				ctx.t.Reset(time.Microsecond * 5000)
			case "t":
				ctx.t.Reset(time.Microsecond * 100)
			case "g":
				ctx.t.Reset(time.Microsecond * 10)
			case "w":
				ctx.t.Stop()
				ctx.stepCh <-0
			}
		}
		default:
		return true
	}
	return true
}
