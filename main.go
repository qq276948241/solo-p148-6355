package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
	"strings"
	"syscall"
	"unsafe"
)

const (
	MaxUndo = 10

	ColorReset  = "\033[0m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorCyan   = "\033[36m"
	ColorMagenta = "\033[35m"
	ColorRed    = "\033[31m"

	MoveUp    = "up"
	MoveDown  = "down"
	MoveLeft  = "left"
	MoveRight = "right"
)

type Level struct {
	Name string   `json:"name"`
	Map  []string `json:"map"`
}

type Position struct {
	X, Y int
}

type MoveRecord struct {
	Player Position
	Boxes  []Position
}

type GameState struct {
	levelName string
	player    Position
	boxes     []Position
	targets   []Position
	walls     map[Position]bool
	floor     map[Position]bool
	width     int
	height    int
	undoStack []MoveRecord
	steps     int
	startTime time.Time
}

func loadLevel(filename string) (*Level, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var level Level
	if err := json.Unmarshal(data, &level); err != nil {
		return nil, err
	}
	return &level, nil
}

func parseLevel(level *Level) *GameState {
	gs := &GameState{
		levelName: level.Name,
		walls:      make(map[Position]bool),
		floor:      make(map[Position]bool),
		startTime:  time.Now(),
	}

	maxWidth := 0
	for _, row := range level.Map {
		if len(row) > maxWidth {
			maxWidth = len(row)
		}
	}
	gs.width = maxWidth
	gs.height = len(level.Map)

	for y, row := range level.Map {
		for x, ch := range row {
			pos := Position{x, y}
			switch ch {
			case '#':
				gs.walls[pos] = true
			case '.', ' ':
				if ch == '.' {
					gs.floor[pos] = true
				}
			case '@':
				gs.floor[pos] = true
				gs.player = pos
			case '$':
				gs.floor[pos] = true
				gs.boxes = append(gs.boxes, pos)
			case 'X':
				gs.floor[pos] = true
				gs.targets = append(gs.targets, pos)
			case '*':
				gs.floor[pos] = true
				gs.boxes = append(gs.boxes, pos)
				gs.targets = append(gs.targets, pos)
			case '+':
				gs.floor[pos] = true
				gs.player = pos
				gs.targets = append(gs.targets, pos)
			}
		}
	}

	return gs
}

func (gs *GameState) isWall(pos Position) bool {
	return gs.walls[pos]
}

func (gs *GameState) isFloor(pos Position) bool {
	return gs.floor[pos]
}

func (gs *GameState) getBoxIndex(pos Position) int {
	for i, box := range gs.boxes {
		if box == pos {
			return i
		}
	}
	return -1
}

func (gs *GameState) isTarget(pos Position) bool {
	for _, t := range gs.targets {
		if t == pos {
			return true
		}
	}
	return false
}

func (gs *GameState) saveUndoState() {
	boxesCopy := make([]Position, len(gs.boxes))
	copy(boxesCopy, gs.boxes)
	record := MoveRecord{
		Player: gs.player,
		Boxes:  boxesCopy,
	}
	gs.undoStack = append(gs.undoStack, record)
	if len(gs.undoStack) > MaxUndo {
		gs.undoStack = gs.undoStack[1:]
	}
}

func (gs *GameState) undo() bool {
	if len(gs.undoStack) == 0 {
		return false
	}
	record := gs.undoStack[len(gs.undoStack)-1]
	gs.undoStack = gs.undoStack[:len(gs.undoStack)-1]
	gs.player = record.Player
	gs.boxes = record.Boxes
	return true
}

func (gs *GameState) move(direction string) bool {
	var dx, dy int
	switch direction {
	case MoveUp:
		dy = -1
	case MoveDown:
		dy = 1
	case MoveLeft:
		dx = -1
	case MoveRight:
		dx = 1
	}

	newPlayer := Position{gs.player.X + dx, gs.player.Y + dy}

	if gs.isWall(newPlayer) || !gs.isFloor(newPlayer) {
		return false
	}

	boxIdx := gs.getBoxIndex(newPlayer)
	if boxIdx != -1 {
		newBox := Position{newPlayer.X + dx, newPlayer.Y + dy}
		if gs.isWall(newBox) || !gs.isFloor(newBox) || gs.getBoxIndex(newBox) != -1 {
			return false
		}
		gs.saveUndoState()
		gs.boxes[boxIdx] = newBox
		gs.player = newPlayer
		gs.steps++
		return true
	}

	gs.saveUndoState()
	gs.player = newPlayer
	gs.steps++
	return true
}

func (gs *GameState) isWin() bool {
	for _, box := range gs.boxes {
		if !gs.isTarget(box) {
			return false
		}
	}
	return true
}

func (gs *GameState) countOnTarget() int {
	count := 0
	for _, box := range gs.boxes {
		if gs.isTarget(box) {
			count++
		}
	}
	return count
}

func (gs *GameState) checkUnsolvable() (bool, string) {
	for _, box := range gs.boxes {
		if gs.isTarget(box) {
			continue
		}
		if gs.isCornered(box) {
			return true, fmt.Sprintf("箱子在位置 (%d,%d) 被推到死角，无法移动！", box.X+1, box.Y+1)
		}
	}
	return false, ""
}

func (gs *GameState) isCornered(box Position) bool {
	up := Position{box.X, box.Y - 1}
	down := Position{box.X, box.Y + 1}
	left := Position{box.X - 1, box.Y}
	right := Position{box.X + 1, box.Y}

	wallUp := gs.isWall(up) || !gs.isFloor(up)
	wallDown := gs.isWall(down) || !gs.isFloor(down)
	wallLeft := gs.isWall(left) || !gs.isFloor(left)
	wallRight := gs.isWall(right) || !gs.isFloor(right)

	if (wallUp || wallDown) && (wallLeft || wallRight) {
		return true
	}

	return false
}

func (gs *GameState) remainingUndo() int {
	return len(gs.undoStack)
}

func clearScreen() {
	fmt.Print("\033[2J\033[H")
}

func (gs *GameState) render(message string) {
	clearScreen()

	onTarget := gs.countOnTarget()
	totalTargets := len(gs.targets)
	elapsed := time.Since(gs.startTime).Seconds()
	minutes := int(elapsed) / 60
	seconds := int(elapsed) % 60

	status := fmt.Sprintf(" %s | 步数: %d | 时间: %02d:%02d | 撤销: %d/%d | 目标: %d/%d ",
		gs.levelName, gs.steps, minutes, seconds, gs.remainingUndo(), MaxUndo, onTarget, totalTargets)
	fmt.Println(strings.Repeat("=", len(status)))
	fmt.Println(status)
	fmt.Println(strings.Repeat("=", len(status)))
	fmt.Println()

	for y := 0; y < gs.height; y++ {
		for x := 0; x < gs.width; x++ {
			pos := Position{x, y}

			if gs.player == pos {
				fmt.Printf("%s@%s", ColorGreen, ColorReset)
			} else if idx := gs.getBoxIndex(pos); idx != -1 {
				if gs.isTarget(pos) {
					fmt.Printf("%s$%s", ColorMagenta, ColorReset)
				} else {
					fmt.Printf("%s$%s", ColorYellow, ColorReset)
				}
			} else if gs.isTarget(pos) {
				fmt.Printf("%sX%s", ColorCyan, ColorReset)
			} else if gs.isWall(pos) {
				fmt.Print("#")
			} else if gs.isFloor(pos) {
				fmt.Print(".")
			} else {
				fmt.Print(" ")
			}
		}
		fmt.Println()
	}

	fmt.Println()
	if message != "" {
		fmt.Printf("%s%s%s\n\n", ColorRed, message, ColorReset)
	}
	fmt.Print("操作: 方向键移动 | u 撤销 | r 重开 | q 退出")
	if message != "" {
		fmt.Print(" | 按任意键继续")
	}
	fmt.Println()
}

type KeyType int

const (
	KeyUnknown KeyType = iota
	KeyUp
	KeyDown
	KeyLeft
	KeyRight
	KeyU
	KeyR
	KeyQ
	KeyOther
)

func readByte() (byte, error) {
	var buf [1]byte
	_, err := os.Stdin.Read(buf[:])
	return buf[0], err
}

func readKey() (KeyType, rune) {
	first, err := readByte()
	if err != nil {
		return KeyUnknown, 0
	}

	if first == 0x1b {
		second, err := readByte()
		if err != nil {
			return KeyUnknown, 0
		}
		if second == '[' {
			third, err := readByte()
			if err != nil {
				return KeyUnknown, 0
			}
			switch third {
			case 'A':
				return KeyUp, 0
			case 'B':
				return KeyDown, 0
			case 'C':
				return KeyRight, 0
			case 'D':
				return KeyLeft, 0
			}
		}
		return KeyUnknown, 0
	}

	switch rune(first) {
	case 'u', 'U':
		return KeyU, rune(first)
	case 'r', 'R':
		return KeyR, rune(first)
	case 'q', 'Q':
		return KeyQ, rune(first)
	default:
		return KeyOther, rune(first)
	}
}

func waitForKey() {
	readByte()
}

func loadAllLevels() ([]*Level, error) {
	var levels []*Level
	for i := 1; i <= 5; i++ {
		filename := fmt.Sprintf("levels/level%d.json", i)
		level, err := loadLevel(filename)
		if err != nil {
			return nil, err
		}
		levels = append(levels, level)
	}
	return levels, nil
}

var (
	kernel32 = syscall.NewLazyDLL("kernel32.dll")
	procGetConsoleMode = kernel32.NewProc("GetConsoleMode")
	procSetConsoleMode = kernel32.NewProc("SetConsoleMode")

	originalInMode uint32
	originalOutMode uint32
	stdinHandle syscall.Handle
	stdoutHandle syscall.Handle
)

const (
	ENABLE_ECHO_INPUT uint32 = 0x0004
	ENABLE_LINE_INPUT uint32 = 0x0002
	ENABLE_PROCESSED_INPUT uint32 = 0x0001
	ENABLE_VIRTUAL_TERMINAL_INPUT uint32 = 0x0200
	ENABLE_VIRTUAL_TERMINAL_PROCESSING uint32 = 0x0004
	ENABLE_PROCESSED_OUTPUT uint32 = 0x0001
)

func enableRawMode() error {
	stdinHandle = syscall.Handle(os.Stdin.Fd())
	_, _, err := procGetConsoleMode.Call(uintptr(stdinHandle), uintptr(unsafe.Pointer(&originalInMode)))
	if err != syscall.Errno(0) {
		return err
	}

	newInMode := originalInMode & ^(ENABLE_ECHO_INPUT | ENABLE_LINE_INPUT | ENABLE_PROCESSED_INPUT)
	newInMode |= ENABLE_VIRTUAL_TERMINAL_INPUT

	_, _, err = procSetConsoleMode.Call(uintptr(stdinHandle), uintptr(newInMode))
	if err != syscall.Errno(0) {
		return err
	}

	stdoutHandle = syscall.Handle(os.Stdout.Fd())
	_, _, err = procGetConsoleMode.Call(uintptr(stdoutHandle), uintptr(unsafe.Pointer(&originalOutMode)))
	if err == syscall.Errno(0) {
		newOutMode := originalOutMode | ENABLE_VIRTUAL_TERMINAL_PROCESSING | ENABLE_PROCESSED_OUTPUT
		procSetConsoleMode.Call(uintptr(stdoutHandle), uintptr(newOutMode))
	}

	return nil
}

func disableRawMode() {
	procSetConsoleMode.Call(uintptr(stdinHandle), uintptr(originalInMode))
	procSetConsoleMode.Call(uintptr(stdoutHandle), uintptr(originalOutMode))
}

func main() {
	if err := enableRawMode(); err != nil {
		fmt.Printf("初始化终端失败: %v\n", err)
		return
	}
	defer disableRawMode()

	levels, err := loadAllLevels()
	if err != nil {
		fmt.Printf("加载关卡失败: %v\n", err)
		return
	}

	totalStartTime := time.Now()
	var gs *GameState
	currentLevel := 0
	message := ""
	waitingForKey := false

	for currentLevel < len(levels) {
		if gs == nil {
			gs = parseLevel(levels[currentLevel])
			message = ""
			waitingForKey = false
		}

		gs.render(message)

		if waitingForKey {
			waitForKey()
			message = ""
			waitingForKey = false
			continue
		}

		key, _ := readKey()

		switch key {
		case KeyQ:
			fmt.Println("\n游戏退出。")
			return
		case KeyR:
			gs = parseLevel(levels[currentLevel])
			message = "已重开当前关卡"
			waitingForKey = false
		case KeyU:
			if gs.undo() {
				message = ""
			} else {
				message = "没有可撤销的步骤"
				waitingForKey = true
			}
		case KeyUp:
			gs.move(MoveUp)
		case KeyDown:
			gs.move(MoveDown)
		case KeyLeft:
			gs.move(MoveLeft)
		case KeyRight:
			gs.move(MoveRight)
		default:
			continue
		}

		if gs.isWin() {
			gs.render("恭喜过关！")
			fmt.Println("\n按任意键进入下一关...")
			waitForKey()
			currentLevel++
			gs = nil
			continue
		}

		if unsolvable, reason := gs.checkUnsolvable(); unsolvable {
			message = reason + " 按 u 撤销或 r 重开"
			waitingForKey = true
		}
	}

	clearScreen()
	totalTime := time.Since(totalStartTime)
	minutes := int(totalTime.Seconds()) / 60
	seconds := int(totalTime.Seconds()) % 60
	fmt.Println(strings.Repeat("=", 50))
	fmt.Println("恭喜通关！")
	fmt.Printf("总用时: %02d:%02d\n", minutes, seconds)
	fmt.Println(strings.Repeat("=", 50))
}

