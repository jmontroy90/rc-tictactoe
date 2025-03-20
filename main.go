package main

import (
	"bufio"
	"fmt"
	"golang.org/x/term"
	"log"
	"os"
	"os/exec"
	"time"
)

func main() {
	restoreFunc, err := configureTerminal()
	if err != nil {
		log.Fatalf("Error configuring terminal: %v", err)
	}
	defer restoreFunc()
	// create board, run game loop
	b := newBoard(3, 3)
	b.printState()
	reader := bufio.NewReader(os.Stdin)
	for {
		input, _, err := reader.ReadRune()
		if err != nil {
			log.Fatalf("Error reading input: %v", err)
		}
		b.processInput(input)
		b.printState()
		// check state for decision-making
		if continueGame := b.evalBoardState(input); !continueGame {
			break
		}
	}
}

func configureTerminal() (restore func(), err error) {
	fd := int(os.Stdin.Fd())
	old, err := term.MakeRaw(fd)
	if err != nil {
		return func() {}, err
	}
	fmt.Print("\033[?25l") // ANSI code for: makes cursor disappear
	// closure to enable restoring original terminal state
	return func() {
		_ = term.Restore(fd, old)
		fmt.Print("\033[?25h") // ANSI code for: make cursor appear
		clearScreen()
	}, nil
}

type board struct {
	grid          [][]rune
	width, height int
	curX, curY    int
	currPlayer    rune // this is a slightly weird design choice
	moveCount     int
}

func newBoard(width, height int) *board {
	b := board{
		width:  width,
		height: height,
		grid:   make([][]rune, height),
		curX:   1, curY: 1, // start with cursor in the middle
		currPlayer: 'X',
	}
	for i := range b.width {
		b.grid[i] = make([]rune, b.height)
	}
	return &b
}

func (b *board) evalBoardState(input rune) (continueGame bool) {
	switch {
	case input == 'q':
		fmt.Printf("\r\n\nExiting...")
		time.Sleep(300 * time.Millisecond)
	case b.checkForWin():
		// We've already switched players at this point; a bit awkward.
		fmt.Printf("\r\n\nPlayer '%c' wins!\r\n", b.otherPlayer())
		time.Sleep(500 * time.Millisecond)
	case b.isFull():
		fmt.Printf("\r\n\nDraw!")
		time.Sleep(500 * time.Millisecond)
	default:
		continueGame = true
	}
	return
}

func (b *board) checkForWin() bool {
	diag1HasWinner, diag2HasWinner := true, true
	for yi := range b.height {
		if pos := b.grid[0][0]; pos != b.grid[yi][yi] || pos == 0 {
			diag1HasWinner = false
		}
		if pos := b.grid[b.height-1][0]; pos != b.grid[b.height-1-yi][yi] || pos == 0 {
			diag2HasWinner = false
		}
		rowHasWinner, colHasWinner := true, true
		for xi := range b.width {
			if pos := b.grid[yi][0]; pos != b.grid[yi][xi] || pos == 0 {
				rowHasWinner = false
			}
			if pos := b.grid[0][yi]; pos != b.grid[xi][yi] || pos == 0 {
				colHasWinner = false
			}
		}
		if rowHasWinner || colHasWinner {
			return true
		}
	}
	return diag1HasWinner || diag2HasWinner
}

func (b *board) processInput(input rune) {
	switch input {
	case 'w', 'W':
		b.curY--
	case 'a', 'A':
		b.curX--
	case 's', 'S':
		b.curY++
	case 'd', 'D':
		b.curX++
	case ' ':
		if b.grid[b.curY][b.curX] != 0 {
			fmt.Printf("\r\n\nCell already taken!") // a lot of this API is a little awkward
			time.Sleep(500 * time.Millisecond)
			return
		}
		b.grid[b.curY][b.curX] = b.currPlayer
		b.switchPlayer()
	}
	b.normalizeXY()
}

// normalize positions in case of overflow
func (b *board) normalizeXY() {
	if b.curY >= b.height {
		b.curY = b.height - 1
	}
	if b.curY < 0 {
		b.curY = 0
	}
	if b.curX >= b.width {
		b.curX = b.width - 1
	}
	if b.curX < 0 {
		b.curX = 0
	}
}

func (b *board) switchPlayer() {
	b.currPlayer = b.otherPlayer()
}

func (b *board) otherPlayer() rune {
	if b.currPlayer == 'X' {
		return 'O'
	} else {
		return 'X'
	}
}

func (b *board) isFull() bool {
	for _, row := range b.grid {
		for _, cell := range row {
			if cell == 0 {
				return false
			}
		}
	}
	return true
}

func clearScreen() {
	cmd := exec.Command("clear") // valid for literally just my computer running macOS
	cmd.Stdout = os.Stdout
	_ = cmd.Run()
}

func (b *board) instructions() string {
	return fmt.Sprintf("- Enter 'q' to quit\r\n" +
		"- WASD to control up-left-down-right\r\n" +
		"- SPACE to input move\r\n\n")
}

func (b *board) printState() {
	clearScreen()
	fmt.Printf(b.instructions())
	for rowIx, row := range b.grid {
		fmt.Print("       ") // tabs don't work in raw mode
		for colIx, cell := range row {
			out := string(cell)
			if b.curX == colIx && b.curY == rowIx && cell == 0 {
				out = "█" // literally just a block UTF-8 character
			} else if b.curX == colIx && b.curY == rowIx && cell != 0 {
				out = fmt.Sprintf("\033[7m%c\033[0m", cell) // highlight existing character
			} else if cell == 0 {
				out = " "
			}
			if colIx+1 != b.width {
				out += "|"
			}
			fmt.Printf("%s", out)
		}
		fmt.Printf("\r\n       ") // tabs yay
		if rowIx+1 != b.height {
			for range b.width*2 - 1 {
				fmt.Printf("-")
			}
			fmt.Printf("\r\n")
		}
	}
	fmt.Printf("\r\n\nCurrent Player: %c", b.currPlayer)
}
