package main

import (
	"fmt"
	"math"
)

type CPU struct {
	// Registers
	PC     uint16
	SP     uint8
	A      uint8
	X      uint8
	Y      uint8
	P      uint8
	memory [65536]uint8

	Cycles uint16
	// TODO: Add fields for flags
}

// Addressing Modes
/*
Returns value, address of a Zero Page of memory, the first 256 bits
*/
func (cpu *CPU) ZeroPage() (uint8, uint16) {
	address := cpu.memory[cpu.PC+1]

	value := cpu.memory[address]

	cpu.PC += 2
	cpu.Cycles += 3
	return value, uint16(address)
}

/*
Returns (value, address)+x of a Zero Page of memory, the first 256 bits
*/
func (cpu *CPU) ZeroPageX() (uint8, uint16) {
	zeroAddress := cpu.memory[cpu.PC+1]
	effectiveAddress := uint8(zeroAddress + cpu.X)
	value := cpu.memory[effectiveAddress]
	cpu.PC += 2
	cpu.Cycles += 4
	return value, uint16(effectiveAddress)
}

func (cpu *CPU) IndexedIndirect() (uint8, uint16) {
	zeroAddress := cpu.memory[cpu.PC+1]
	effectiveAddress := zeroAddress + cpu.X&0xFF
	value := cpu.memory[effectiveAddress]
	return value, uint16(effectiveAddress)
}

func (cpu *CPU) IndirectIndex() (uint8, uint16) {
	zeroAddress := cpu.memory[cpu.PC+1]
	effectiveAddress := zeroAddress + cpu.Y&0xFF
	value := cpu.memory[effectiveAddress]
	cpu.Cycles += 5
	cpu.PC += 2
	return value, uint16(effectiveAddress)
}

func (cpu *CPU) Relative() uint16 {
	offset := int8(cpu.memory[cpu.PC+1])

	targetAddress := cpu.PC + 2 + uint16(offset)
	return targetAddress
}

func (cpu *CPU) BEQRelative() {
	targetAddress := cpu.Relative()

	// if the zero flag is flase then keep it moving
	if cpu.P&0x02 != 0 {
		cpu.PC = targetAddress
	} else {
		cpu.PC += 2
	}
}

func (cpu *CPU) ZeroPageY() uint8 {
	zeroAddress := cpu.memory[cpu.PC+1]
	effectiveAddress := uint8(zeroAddress + cpu.Y)
	value := cpu.memory[effectiveAddress]
	cpu.PC += 2
	cpu.Cycles += 4
	return value
}

func (cpu *CPU) Immediate() uint8 {
	value := cpu.memory[cpu.PC+1]
	cpu.PC += 2
	cpu.Cycles += 2
	return value
}

// Increment all usages of this by 1
func (cpu *CPU) Absolute() uint16 {
	lowByte := uint16(cpu.memory[cpu.PC+1])
	highByte := uint16(cpu.memory[cpu.PC+2])

	absoluteAddress := (highByte << 8) | lowByte
	cpu.PC += 3
	cpu.Cycles += 4
	return uint16(absoluteAddress)
}

// Increment all usages of this by 1
func (cpu *CPU) AbsoluteX() uint16 {
	lowByte := uint16(cpu.memory[cpu.PC+1])
	highByte := uint16(cpu.memory[cpu.PC+2])
	absoluteAddress := (highByte << 8) | lowByte
	absoluteAddress += uint16(cpu.X)
	cpu.Cycles += 4
	cpu.PC += 3
	return uint16(absoluteAddress)
}

// Increment all usages of this by 1
func (cpu *CPU) AbsoluteY() uint16 {
	lowByte := uint16(cpu.memory[cpu.PC+1])
	highByte := uint16(cpu.memory[cpu.PC+2])
	absoluteAddress := (highByte << 8) | lowByte
	absoluteAddress += uint16(cpu.Y)
	cpu.Cycles += 4
	cpu.PC += 3
	return uint16(absoluteAddress)
}

func (cpu *CPU) setNegativeFlag(value uint8) {
	if getBit(value, 7) {
		cpu.P = setBit(cpu.P, 7)
	} else {
		cpu.P = clearBit(cpu.P, 7)
	}
}

func (cpu *CPU) setCarryFlag(value1 uint8, value2 uint8) {
	if value1 > math.MaxUint8-value2 {
		cpu.P = setBit(cpu.P, 0)
	} else {
		cpu.P = clearBit(cpu.P, 0)
	}
}

func (cpu *CPU) setZeroFlag(value uint8) {
	if value == 0 {
		setBit(cpu.P, 1)
	} else {
		clearBit(cpu.P, 1)
	}
}

func clearBit(n uint8, pos uint8) uint8 {
	return n &^ (1 << pos)
}

func getBit(n uint8, pos uint8) bool {
	return (n & (1 << pos)) != 0
}

func toggleBit(n uint8, pos uint) uint8 {
	return n ^ (1 << pos)
}

func setBit(n uint8, pos uint8) uint8 {
	return n | (1 << pos)
}

func (cpu *CPU) LDAImmediate() {
	value := cpu.Immediate()
	cpu.A = value
	cpu.setNegativeFlag(cpu.A)
	cpu.setZeroFlag(cpu.A)
	fmt.Printf("LDA #$%02X\n", value)
}

func (cpu *CPU) LDAZeroPage() {
	value, _ := cpu.ZeroPage()
	cpu.A = value
	cpu.setNegativeFlag(cpu.A)
	cpu.setZeroFlag(cpu.A)
	fmt.Printf("LDA #$%02X\n", value)
}

func (cpu *CPU) LDAZeroPageX() {
	value, _ := cpu.ZeroPageX()
	cpu.A = value
	cpu.setNegativeFlag(cpu.A)
	cpu.setZeroFlag(cpu.A)
	fmt.Printf("LDA #$%02X\n", value)
}

func (cpu *CPU) LDAIndexIndirect() {
	value, _ := cpu.IndexedIndirect()
	cpu.A = value
	cpu.setNegativeFlag(cpu.A)
	cpu.setZeroFlag(cpu.A)
	fmt.Printf("LDA #$%02X\n", value)
}

func (cpu *CPU) LDAIndirectIndex() {
	value, _ := cpu.IndirectIndex()
	cpu.A = value
	cpu.setNegativeFlag(cpu.A)
	cpu.setZeroFlag(cpu.A)
	fmt.Printf("LDA #$%02X\n", value)
}

func (cpu *CPU) LDAAbsolute() {
	address := cpu.Absolute()
	value := cpu.memory[address]
	cpu.A = value
	cpu.setNegativeFlag(cpu.A)
	cpu.setZeroFlag(cpu.A)
	fmt.Printf("LDA #$%02X\n", value)
}

func (cpu *CPU) LDXImmediate() {
	value := cpu.Immediate()
	cpu.X = value
	cpu.setNegativeFlag(cpu.X)
	cpu.setZeroFlag(cpu.X)
	fmt.Printf("LDX #$%02X\n", value)
}

func (cpu *CPU) LDXAbsolute() {
	address := cpu.Absolute()
	value := cpu.memory[address]
	cpu.X = value
	cpu.setNegativeFlag(cpu.X)
	cpu.setZeroFlag(cpu.X)
	fmt.Printf("LDX #$%02X\n", value)
}

func (cpu *CPU) LDXZeroPage() {
	value, _ := cpu.ZeroPage()
	cpu.X = value
	cpu.setNegativeFlag(cpu.X)
	cpu.setZeroFlag(cpu.X)
	fmt.Printf("LDX #$%02X\n", value)
}

func (cpu *CPU) LDXZeroPageX() {
	value, _ := cpu.ZeroPageX()
	cpu.X = value
	cpu.setNegativeFlag(cpu.X)
	cpu.setZeroFlag(cpu.X)
	fmt.Printf("LDX #$%02X\n", value)
}

func (cpu *CPU) LDYImmediate() {
	value := cpu.Immediate()
	cpu.Y = value
	cpu.setNegativeFlag(cpu.Y)

	cpu.setZeroFlag(cpu.Y)
	fmt.Printf("LDX #$%02X\n", value)
}

func (cpu *CPU) LDYAbsolute() {
	address := cpu.Absolute()
	value := cpu.memory[address]
	cpu.Y = value
	cpu.setNegativeFlag(cpu.Y)

	cpu.setZeroFlag(cpu.Y)
	fmt.Printf("LDX #$%02X\n", value)
}

func (cpu *CPU) LDYAbsoluteX() {
	address := cpu.AbsoluteX()
	value := cpu.memory[address]
	cpu.Y = value
	cpu.setNegativeFlag(cpu.Y)

	cpu.setZeroFlag(cpu.Y)
	fmt.Printf("LDX #$%02X\n", value)
}

func (cpu *CPU) LDYZeroPage() {
	value, _ := cpu.ZeroPage()
	cpu.Y = value
	cpu.setNegativeFlag(cpu.Y)

	cpu.setZeroFlag(cpu.Y)
	fmt.Printf("LDX #$%02X\n", value)
}

func (cpu *CPU) LDYZeroPageX() {
	value, _ := cpu.ZeroPageX()
	cpu.Y = value
	cpu.setNegativeFlag(cpu.Y)

	cpu.setZeroFlag(cpu.Y)
	fmt.Printf("LDX #$%02X\n", value)
}

func (cpu *CPU) STAAbsolute() {
	address := cpu.Absolute()
	cpu.memory[address] = cpu.A
}

func (cpu *CPU) STAAbsoluteX() {
	address := cpu.AbsoluteX()
	cpu.memory[address] = cpu.A
}

func (cpu *CPU) STAAbsoluteY() {
	address := cpu.AbsoluteY()
	cpu.memory[address] = cpu.A
}

func (cpu *CPU) STAZeroPage() {
	_, address := cpu.ZeroPage()
	cpu.memory[address] = cpu.A
}

func (cpu *CPU) STAZeroPageX() {
	_, address := cpu.ZeroPageX()
	cpu.memory[address] = cpu.A
}

func (cpu *CPU) STAIndexIndirect() {
	_, address := cpu.IndexedIndirect()
	cpu.memory[address] = cpu.A
}

func (cpu *CPU) STAIndirectIndex() {
	_, address := cpu.IndirectIndex()
	cpu.memory[address] = cpu.A
}

func (cpu *CPU) STXAbsolute() {
	address := cpu.Absolute()
	cpu.memory[address] = cpu.X
}

func (cpu *CPU) STXXZeroPageX() {
	_, address := cpu.ZeroPageX()
	cpu.memory[address] = cpu.X
}

func (cpu *CPU) STXZeroPage() {
	_, address := cpu.ZeroPage()
	cpu.memory[address] = cpu.X
}

func (cpu *CPU) STYAbsolute() {
	address := cpu.Absolute()
	cpu.memory[address] = cpu.Y
}

func (cpu *CPU) STYZeroPageX() {
	_, address := cpu.ZeroPageX()
	cpu.memory[address] = cpu.Y
}

func (cpu *CPU) STYZeroPage() {
	_, address := cpu.ZeroPage()
	cpu.memory[address] = cpu.Y
}

func (cpu *CPU) TAX() {
	cpu.X = cpu.A
	cpu.setZeroFlag(cpu.X)
	cpu.setNegativeFlag(cpu.X)
	cpu.Cycles += 2
	cpu.PC++
}

func (cpu *CPU) TAY() {
	cpu.Y = cpu.A
	cpu.setZeroFlag(cpu.Y)
	cpu.setNegativeFlag(cpu.Y)
	cpu.Cycles += 2
	cpu.PC++
}

func (cpu *CPU) TSX() {
	cpu.X = cpu.SP
	cpu.setZeroFlag(cpu.X)
	cpu.setNegativeFlag(cpu.X)
	cpu.Cycles += 2
	cpu.PC++
}

func (cpu *CPU) TXA() {
	cpu.A = cpu.SP
	cpu.setZeroFlag(cpu.A)
	cpu.setNegativeFlag(cpu.A)
	cpu.Cycles += 2
	cpu.PC++
}

func (cpu *CPU) TXS() {
	cpu.SP = cpu.X
	cpu.Cycles += 2
	cpu.PC++
}

func (cpu *CPU) TYA() {
	cpu.A = cpu.Y
	cpu.Cycles += 2
	cpu.setNegativeFlag(cpu.A)
	cpu.setZeroFlag(cpu.A)
	cpu.PC++
}

func (cpu *CPU) Push(value uint8) {
	cpu.memory[0x0100+uint16(cpu.SP)] = value
	cpu.SP--
	cpu.Cycles += 3
}

func (cpu *CPU) Pull() uint8 {
	cpu.SP++
	return cpu.memory[0x0100+uint16(cpu.SP)]
}

func (cpu *CPU) PHA() {
	cpu.Push(cpu.A)
	cpu.PC++
}

func (cpu *CPU) PHP() {
	cpu.Push(cpu.P)
}

func (cpu *CPU) PLA() {
	cpu.A = cpu.Pull()
	cpu.setZeroFlag(cpu.A)
	cpu.setNegativeFlag(cpu.A)
	cpu.PC++
}

func (cpu *CPU) PLP() {
	cpu.P = cpu.Pull()
	cpu.PC++
}

func (cpu *CPU) ANDImmediate() {
	val := cpu.Immediate()
	cpu.A = val & cpu.A
	cpu.setZeroFlag(cpu.A)
	cpu.setNegativeFlag(cpu.A)
}

func (cpu *CPU) ANDAbsolute() {
	address := cpu.Absolute()

	val := cpu.memory[address]
	cpu.A = val & cpu.A
	cpu.setZeroFlag(cpu.A)
	cpu.setNegativeFlag(cpu.A)
}

func (cpu *CPU) ANDAbsoluteX() {
	address := cpu.AbsoluteX()

	val := cpu.memory[address]
	cpu.A = val & cpu.A
	cpu.setZeroFlag(cpu.A)
	cpu.setNegativeFlag(cpu.A)
}

func (cpu *CPU) ANDAbsoluteY() {
	address := cpu.AbsoluteY()

	val := cpu.memory[address]
	cpu.A = val & cpu.A
	cpu.setZeroFlag(cpu.A)
	cpu.setNegativeFlag(cpu.A)
}

func (cpu *CPU) ANDZeroPage() {
	_, address := cpu.ZeroPage()

	val := cpu.memory[address]
	cpu.A = val & cpu.A
	cpu.setZeroFlag(cpu.A)
	cpu.setNegativeFlag(cpu.A)
}

func (cpu *CPU) ANDZeroPageX() {
	_, address := cpu.ZeroPageX()

	val := cpu.memory[address]
	cpu.A = val & cpu.A
	cpu.setZeroFlag(cpu.A)
	cpu.setNegativeFlag(cpu.A)
}

func (cpu *CPU) ANDIndexIndirect() {
	_, address := cpu.IndexedIndirect()

	val := cpu.memory[address]
	cpu.A = val & cpu.A
	cpu.setZeroFlag(cpu.A)
	cpu.setNegativeFlag(cpu.A)
}

func (cpu *CPU) ANDIndirectIndex() {
	_, address := cpu.IndirectIndex()
	val := cpu.memory[address]
	cpu.A = val & cpu.A
	cpu.setZeroFlag(cpu.A)
	cpu.setNegativeFlag(cpu.A)
}

func (cpu *CPU) AddWCarryImmediate() {
	value := cpu.Immediate()
	cpu.setCarryFlag(cpu.A, value)
	cpu.A += value
	cpu.setZeroFlag(cpu.A)
	cpu.setNegativeFlag(cpu.A)
}

func (cpu *CPU) AddWCarryAbsolute() {
	address := cpu.Absolute()
	value := cpu.memory[address]
	cpu.setCarryFlag(cpu.A, value)
	cpu.A += value
	cpu.setZeroFlag(cpu.A)
	cpu.setNegativeFlag(cpu.A)
}

func (cpu *CPU) EORImmediate() {
	value := cpu.Immediate()
	cpu.A = value ^ cpu.A
	cpu.setZeroFlag(cpu.A)
	cpu.setNegativeFlag(cpu.A)
}

func (cpu *CPU) EORZeroPage() {
	_, address := cpu.ZeroPage()
	value := cpu.memory[address]
	cpu.A = value ^ cpu.A
	cpu.setZeroFlag(cpu.A)
	cpu.setNegativeFlag(cpu.A)
}

func (cpu *CPU) EORZeroPageX() {
	_, address := cpu.ZeroPageX()
	value := cpu.memory[address]
	cpu.A = value ^ cpu.A
	cpu.setZeroFlag(cpu.A)
	cpu.setNegativeFlag(cpu.A)
}

func (cpu *CPU) EORAbsolute() {
	address := cpu.Absolute()
	value := cpu.memory[address]
	cpu.A = value ^ cpu.A
	cpu.setZeroFlag(cpu.A)
	cpu.setNegativeFlag(cpu.A)
}

func (cpu *CPU) EORAbsoluteX() {
	address := cpu.AbsoluteX()
	value := cpu.memory[address]
	cpu.A = value ^ cpu.A
	cpu.setZeroFlag(cpu.A)
	cpu.setNegativeFlag(cpu.A)
}

func (cpu *CPU) EORAbsoluteY() {
	address := cpu.AbsoluteY()
	value := cpu.memory[address]
	cpu.A = value ^ cpu.A
	cpu.setZeroFlag(cpu.A)
	cpu.setNegativeFlag(cpu.A)
}

func (cpu *CPU) EORIndirectIndex() {
	_, address := cpu.IndirectIndex()
	value := cpu.memory[address]
	cpu.A = value ^ cpu.A
	cpu.setZeroFlag(cpu.A)
	cpu.setNegativeFlag(cpu.A)
}

func (cpu *CPU) EORIndexIndirect() {
	_, address := cpu.IndexedIndirect()
	value := cpu.memory[address]
	cpu.A = value ^ cpu.A
	cpu.setZeroFlag(cpu.A)
	cpu.setNegativeFlag(cpu.A)
}

func (cpu *CPU) ORAImmediate() {
	value := cpu.Immediate()
	cpu.A = value | cpu.A

	cpu.setZeroFlag(cpu.A)
	cpu.setNegativeFlag(cpu.A)
}

func (cpu *CPU) ORAZeroPage() {
	_, address := cpu.ZeroPage()
	value := cpu.memory[address]
	cpu.A = value | cpu.A

	cpu.setZeroFlag(cpu.A)
	cpu.setNegativeFlag(cpu.A)
}

func (cpu *CPU) ORAZeroPageX() {
	_, address := cpu.ZeroPageX()
	value := cpu.memory[address]
	cpu.A = value | cpu.A

	cpu.setZeroFlag(cpu.A)
	cpu.setNegativeFlag(cpu.A)
}

func (cpu *CPU) ORAAbsolute() {
	address := cpu.Absolute()
	value := cpu.memory[address]
	cpu.A = value | cpu.A

	cpu.setZeroFlag(cpu.A)
	cpu.setNegativeFlag(cpu.A)
}

func (cpu *CPU) ORAAbsoluteX() {
	address := cpu.AbsoluteX()
	value := cpu.memory[address]
	cpu.A = value | cpu.A

	cpu.setZeroFlag(cpu.A)
	cpu.setNegativeFlag(cpu.A)
}

func (cpu *CPU) ORAAbsoluteY() {
	address := cpu.AbsoluteY()
	value := cpu.memory[address]
	cpu.A = value | cpu.A

	cpu.setZeroFlag(cpu.A)
	cpu.setNegativeFlag(cpu.A)
}

func (cpu *CPU) ORAIndirectIndex() {
	_, address := cpu.IndirectIndex()
	value := cpu.memory[address]
	cpu.A = value | cpu.A

	cpu.setZeroFlag(cpu.A)
	cpu.setNegativeFlag(cpu.A)
}

func (cpu *CPU) ORAIndexIndirect() {
	_, address := cpu.IndexedIndirect()
	value := cpu.memory[address]
	cpu.A = value | cpu.A

	cpu.setZeroFlag(cpu.A)
	cpu.setNegativeFlag(cpu.A)
}

func (cpu *CPU) BITZeroPage() {
	value, _ := cpu.ZeroPage()

	if cpu.A&value == 0 {
		cpu.P = setBit(cpu.P, 1)
	} else {
		cpu.P = clearBit(cpu.P, 1)
	}

	if getBit(value, 6) {
		cpu.P = setBit(cpu.P, 6)
	} else {
		cpu.P = clearBit(cpu.P, 6)
	}

	if getBit(value, 7) {
		cpu.P = setBit(cpu.P, 7)
	} else {
		cpu.P = clearBit(cpu.P, 7)
	}
}

func (cpu *CPU) BITAbsolute() {
	address := cpu.Absolute()
	value := cpu.memory[address]

	if cpu.A&value == 0 {
		cpu.P = setBit(cpu.P, 1)
	} else {
		cpu.P = clearBit(cpu.P, 1)
	}

	if getBit(value, 6) {
		cpu.P = setBit(cpu.P, 6)
	} else {
		cpu.P = clearBit(cpu.P, 6)
	}

	if getBit(value, 7) {
		cpu.P = setBit(cpu.P, 7)
	} else {
		cpu.P = clearBit(cpu.P, 7)
	}
}

func (cpu *CPU) NOP() {
	cpu.PC = cpu.PC + 1
	fmt.Println("NOP")
}

func (cpu *CPU) LoadProgram(program []uint8, startAddress uint16) {
	for i, data := range program {
		cpu.memory[startAddress+uint16(i)] = data
	}
	cpu.PC = startAddress // Set the program counter to the start of our program
}

func (cpu *CPU) ExecuteInstructions(opcode uint8) {
	switch opcode {
	case 0xA9: // LDA Immediate
		cpu.LDAImmediate()
	case 0xA2: // LDX Immediate
		cpu.LDXImmediate()
	case 0xA0: // LDY Immediate
		cpu.LDYImmediate()
	case 0xEA: // NOP
		cpu.NOP()
	default:
		panic(fmt.Sprintf("Unhandled opcode: %02X\n", opcode))
	}
}

func main() {
	cpu := &CPU{}
	// Example program: LDA #$05, LDX #$0A, NOP
	program := []uint8{0xA9, 0x05, 0xA2, 0x0A, 0xEA}

	cpu.LoadProgram(program, 0x8000) // Load at address 0x8000

	cpu.setNegativeFlag(0x80)
	// Now run the emulator
	for {
		opcode := cpu.memory[cpu.PC]
		cpu.ExecuteInstructions(opcode)

		fmt.Printf("CPU status: %08b\n", cpu.P)
		// Add any necessary checks or breaks here, such as stopping the loop
		// based on a certain condition or user input
	}
}
