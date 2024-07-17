package main

import (
	"fmt"
	"io"
	"math"
	"os"
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
}

// Addressing Modes
/*
Returns value, address of a Zero Page of memory, the first 256 bits base of 3 cycles
*/
func (cpu *CPU) ZeroPage() (uint8, uint16) {
	address := cpu.memory[cpu.PC+1]

	value := cpu.memory[address]

	cpu.PC += 2
	cpu.Cycles += 3
	return value, uint16(address)
}

/*
Returns (value, address)+x of a Zero Page of memory, the first 256 bits base of 3 cycles
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

func (cpu *CPU) Indirect() uint16 {
	lowByte := uint16(cpu.memory[cpu.PC+1])
	highByte := uint16(cpu.memory[cpu.PC+2])

	effectiveAddress := (highByte << 8) | lowByte
	cpu.PC += 3
	// NES/6502 bug: If the indirect vector falls on a page boundary (i.e., $xxFF where xx is any value from $00 to $FF),
	// the second byte is fetched from the beginning of that page rather than the beginning of the next page.
	if lowByte == 0xFF {
		lowAddress := uint16(cpu.memory[effectiveAddress])
		highAddress := uint16(cpu.memory[effectiveAddress&0xFF00])
		return (highAddress << 8) | lowAddress
	} else {
		lowAddress := uint16(cpu.memory[effectiveAddress])
		highAddress := uint16(cpu.memory[effectiveAddress+1])
		return (highAddress << 8) | lowAddress
	}
}

func (cpu *CPU) Relative() uint16 {
	offset := int8(cpu.memory[cpu.PC+1])
	// targetAddress := cpu.PC + 2 + uint16(offset)
	// return targetAddress
	return uint16(offset)
}

func (cpu *CPU) Relativetest() int8 {
	offset := int8(cpu.memory[cpu.PC+1])
	// targetAddress := cpu.PC + 2 + uint16(offset)
	// return targetAddress
	return (offset)
}

// Returns the first 8 bits in memory
func (cpu *CPU) ZeroPageY() (uint8, uint16) {
	zeroAddress := cpu.memory[cpu.PC+1]
	effectiveAddress := uint8(zeroAddress + cpu.Y)
	value := cpu.memory[effectiveAddress]
	cpu.PC += 2
	cpu.Cycles += 4
	return value, uint16(effectiveAddress)
}

// Returns a value immediately supplied in the command, takes base of 2 Cycles
func (cpu *CPU) Immediate() uint8 {
	value := cpu.memory[cpu.PC+1]
	cpu.PC += 2
	cpu.Cycles += 2
	return value
}

// Returns a 16 bit memory address, takes base of 4 Cycles
func (cpu *CPU) Absolute() uint16 {
	lowByte := uint16(cpu.memory[cpu.PC+1])
	highByte := uint16(cpu.memory[cpu.PC+2])

	absoluteAddress := (highByte << 8) | lowByte
	cpu.PC += 3
	cpu.Cycles += 4
	return uint16(absoluteAddress)
}

// Returns a 16 bit memory address + value in x register, takes base of 4 Cycles
func (cpu *CPU) AbsoluteX() uint16 {
	lowByte := uint16(cpu.memory[cpu.PC+1])
	highByte := uint16(cpu.memory[cpu.PC+2])
	absoluteAddress := (highByte << 8) | lowByte
	absoluteAddress += uint16(cpu.X)
	cpu.Cycles += 4
	cpu.PC += 3
	return uint16(absoluteAddress)
}

// Returns a 16 bit memory address + value in Y register, takes base of 4 Cycles
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
		cpu.P = setBit(cpu.P, 1)
	} else {
		cpu.P = clearBit(cpu.P, 1)
	}
}

// TODO
// Finish this implementation
func (cpu *CPU) setADDOverflowFlag(value1 uint, value2 uint) {
	if value1+value2 > 255 {
		fmt.Print("OVERFLOW")
		cpu.P = setBit(cpu.P, 6)
		cpu.P = setBit(cpu.P, 0)
	} else {
		cpu.P = clearBit(cpu.P, 6)
	}
}

func (cpu *CPU) setSUBOverflowFlag(value1 uint, value2 uint) {
	if int(value1)-int(value2) < 0 {
		cpu.P = setBit(cpu.P, 6)
	} else {
		cpu.P = clearBit(cpu.P, 6)
		cpu.P = clearBit(cpu.P, 0)
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

func (cpu *CPU) LDAAbsoluteX() {
	address := cpu.AbsoluteX()
	value := cpu.memory[address]
	cpu.A = value
	cpu.setNegativeFlag(cpu.A)
	cpu.setZeroFlag(cpu.A)
	fmt.Printf("LDA #$%02X\n", value)
}

func (cpu *CPU) LDAAbsoluteY() {
	address := cpu.AbsoluteY()
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

func (cpu *CPU) LDXZeroPageX() {
	_, address := cpu.ZeroPageX()
	value := cpu.memory[address]
	cpu.X = value
	cpu.setNegativeFlag(cpu.X)
	cpu.setZeroFlag(cpu.X)
	fmt.Printf("LDX #$%02X\n", value)
}

func (cpu *CPU) LDXAbsoluteY() {
	address := cpu.AbsoluteY()
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

func (cpu *CPU) LDXZeroPageY() {
	value, _ := cpu.ZeroPageY()
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

func (cpu *CPU) STXZeroPageY() {
	_, address := cpu.ZeroPageY()
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
	cpu.PC++
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

func (cpu *CPU) ADCAbsolute() {
	address := cpu.Absolute()
	value := cpu.memory[address]
	cpu.setCarryFlag(cpu.A, value)
	cpu.setADDOverflowFlag(uint(cpu.A), uint(value))
	cpu.A += value
	cpu.setZeroFlag(cpu.A)
	cpu.setNegativeFlag(cpu.A)
}

func (cpu *CPU) ADCImmediate() {
	value := cpu.Immediate()
	cpu.setCarryFlag(cpu.A, value)
	cpu.setADDOverflowFlag(uint(cpu.A), uint(value))
	cpu.A += value
	cpu.setZeroFlag(cpu.A)
	cpu.setNegativeFlag(cpu.A)
}

func (cpu *CPU) ADCAbsoluteX() {
	address := cpu.AbsoluteX()
	value := cpu.memory[address]
	cpu.setCarryFlag(cpu.A, value)
	cpu.setADDOverflowFlag(uint(cpu.A), uint(value))
	cpu.A += value
	cpu.setZeroFlag(cpu.A)
	cpu.setNegativeFlag(cpu.A)
}

func (cpu *CPU) ADCAbsoluteY() {
	address := cpu.AbsoluteY()
	value := cpu.memory[address]
	cpu.setCarryFlag(cpu.A, value)
	cpu.setADDOverflowFlag(uint(cpu.A), uint(value))
	cpu.A += value
	cpu.setZeroFlag(cpu.A)
	cpu.setNegativeFlag(cpu.A)
}

func (cpu *CPU) ADCZeroPage() {
	_, address := cpu.ZeroPage()
	value := cpu.memory[address]
	cpu.setCarryFlag(cpu.A, value)
	cpu.setADDOverflowFlag(uint(cpu.A), uint(value))
	cpu.A += value
	cpu.setZeroFlag(cpu.A)
	cpu.setNegativeFlag(cpu.A)
}

func (cpu *CPU) ADCZeroPageX() {
	_, address := cpu.ZeroPageX()
	value := cpu.memory[address]
	cpu.setCarryFlag(cpu.A, value)
	cpu.setADDOverflowFlag(uint(cpu.A), uint(value))
	cpu.A += value
	cpu.setZeroFlag(cpu.A)
	cpu.setNegativeFlag(cpu.A)
}

func (cpu *CPU) ADCIndirectIndex() {
	_, address := cpu.IndirectIndex()
	value := cpu.memory[address]
	cpu.setCarryFlag(cpu.A, value)
	cpu.setADDOverflowFlag(uint(cpu.A), uint(value))
	cpu.A += value
	cpu.setZeroFlag(cpu.A)
	cpu.setNegativeFlag(cpu.A)
}

func (cpu *CPU) ADCIndexIndirect() {
	_, address := cpu.IndexedIndirect()
	value := cpu.memory[address]
	cpu.setCarryFlag(cpu.A, value)
	cpu.setADDOverflowFlag(uint(cpu.A), uint(value))
	cpu.A += value
	cpu.setZeroFlag(cpu.A)
	cpu.setNegativeFlag(cpu.A)
}

func (cpu *CPU) SBCImmediate() {
	value := cpu.Immediate()
	oldCarry := uint8(0)
	if getBit(cpu.P, 0) {
		oldCarry = 1
	}
	value = value + (1 - oldCarry)
	cpu.setCarryFlag(cpu.A, value)
	cpu.setSUBOverflowFlag(uint(cpu.A), uint(value))
	cpu.A -= value
	cpu.setZeroFlag(cpu.A)
	cpu.setNegativeFlag(cpu.A)
}

func (cpu *CPU) SBCZeroPage() {
	value, _ := cpu.ZeroPage()
	oldCarry := uint8(0)
	if getBit(cpu.P, 0) {
		oldCarry = 1
	}
	value = value + (1 - oldCarry)
	cpu.setCarryFlag(cpu.A, value)
	cpu.setSUBOverflowFlag(uint(cpu.A), uint(value))
	cpu.A -= value
	cpu.setZeroFlag(cpu.A)
	cpu.setNegativeFlag(cpu.A)
}

func (cpu *CPU) SBCZeroPageX() {
	value, _ := cpu.ZeroPageX()
	oldCarry := uint8(0)
	if getBit(cpu.P, 0) {
		oldCarry = 1
	}
	value = value + (1 - oldCarry)
	cpu.setCarryFlag(cpu.A, value)
	cpu.setSUBOverflowFlag(uint(cpu.A), uint(value))
	cpu.A -= value
	cpu.setZeroFlag(cpu.A)
	cpu.setNegativeFlag(cpu.A)
}

func (cpu *CPU) SBCAbsolute() {
	address := cpu.Absolute()
	value := cpu.memory[address]

	oldCarry := uint8(0)
	if getBit(cpu.P, 0) {
		oldCarry = 1
	}
	value = value + (1 - oldCarry)
	cpu.setCarryFlag(cpu.A, value)
	cpu.setSUBOverflowFlag(uint(cpu.A), uint(value))
	cpu.A -= value
	cpu.setZeroFlag(cpu.A)
	cpu.setNegativeFlag(cpu.A)
}

func (cpu *CPU) SBCAbsoluteX() {
	address := cpu.AbsoluteX()
	value := cpu.memory[address]
	oldCarry := uint8(0)
	if getBit(cpu.P, 0) {
		oldCarry = 1
	}
	value = value + (1 - oldCarry)
	cpu.setCarryFlag(cpu.A, value)
	cpu.setSUBOverflowFlag(uint(cpu.A), uint(value))
	cpu.A -= value
	cpu.setZeroFlag(cpu.A)
	cpu.setNegativeFlag(cpu.A)
}

func (cpu *CPU) SBCAbsoluteY() {
	address := cpu.AbsoluteY()
	value := cpu.memory[address]
	oldCarry := uint8(0)
	if getBit(cpu.P, 0) {
		oldCarry = 1
	}
	value = value + (1 - oldCarry)
	cpu.setCarryFlag(cpu.A, value)
	cpu.setSUBOverflowFlag(uint(cpu.A), uint(value))
	cpu.A -= value
	cpu.setZeroFlag(cpu.A)
	cpu.setNegativeFlag(cpu.A)
}

func (cpu *CPU) SBCIndirectIndex() {
	value, _ := cpu.IndirectIndex()
	oldCarry := uint8(0)
	if getBit(cpu.P, 0) {
		oldCarry = 1
	}
	value = value + (1 - oldCarry)
	cpu.setCarryFlag(cpu.A, value)
	cpu.setSUBOverflowFlag(uint(cpu.A), uint(value))
	cpu.A -= value
	cpu.setZeroFlag(cpu.A)
	cpu.setNegativeFlag(cpu.A)
}

func (cpu *CPU) SBCIndexIndirect() {
	value, _ := cpu.IndexedIndirect()
	oldCarry := uint8(0)
	if getBit(cpu.P, 0) {
		oldCarry = 1
	}
	value = value + (1 - oldCarry)
	cpu.setCarryFlag(cpu.A, value)
	cpu.setSUBOverflowFlag(uint(cpu.A), uint(value))
	cpu.A -= value
	cpu.setZeroFlag(cpu.A)
	cpu.setNegativeFlag(cpu.A)
}

func (cpu *CPU) CMPImmediate() {
	if cpu.A > cpu.Immediate() {
		cpu.P = setBit(cpu.P, 0)
	}
	if cpu.A == cpu.Immediate() {
		cpu.P = setBit(cpu.P, 1)
	}
}

func (cpu *CPU) CMPZeroPage() {
	value, _ := cpu.ZeroPage()
	if cpu.A > value {
		cpu.P = setBit(cpu.P, 0)
	}
	if cpu.A == value {
		cpu.P = setBit(cpu.P, 1)
	}
}

func (cpu *CPU) CMPZeroPageX() {
	value, _ := cpu.ZeroPageX()
	if cpu.A > value {
		cpu.P = setBit(cpu.P, 0)
	}
	if cpu.A == value {
		cpu.P = setBit(cpu.P, 1)
	}
}

func (cpu *CPU) CMPAbsolute() {
	address := cpu.Absolute()
	value := cpu.memory[address]
	if cpu.A > value {
		cpu.P = setBit(cpu.P, 0)
	}
	if cpu.A == value {
		cpu.P = setBit(cpu.P, 1)
	}
}

func (cpu *CPU) CMPAbsoluteY() {
	address := cpu.AbsoluteY()
	value := cpu.memory[address]
	if cpu.A > value {
		cpu.P = setBit(cpu.P, 0)
	}
	if cpu.A == value {
		cpu.P = setBit(cpu.P, 1)
	}
}

func (cpu *CPU) CMPAbsoluteX() {
	address := cpu.AbsoluteX()
	value := cpu.memory[address]
	if cpu.A > value {
		cpu.P = setBit(cpu.P, 0)
	}
	if cpu.A == value {
		cpu.P = setBit(cpu.P, 1)
	}
}

func (cpu *CPU) CMPIndirectIndirect() {
	value, _ := cpu.IndirectIndex()
	if cpu.A > value {
		cpu.P = setBit(cpu.P, 0)
	}
	if cpu.A == value {
		cpu.P = setBit(cpu.P, 1)
	}
}

func (cpu *CPU) CMPIndexedIndirect() {
	value, _ := cpu.IndexedIndirect()
	if cpu.A > value {
		cpu.P = setBit(cpu.P, 0)
	}
	if cpu.A == value {
		cpu.P = setBit(cpu.P, 1)
	}
}

func (cpu *CPU) CPXImmediate() {
	if cpu.X > cpu.Immediate() {
		cpu.P = setBit(cpu.P, 0)
	}
	if cpu.X == cpu.Immediate() {
		cpu.P = setBit(cpu.P, 1)
	}
}

func (cpu *CPU) CPXZeroPage() {
	value, _ := cpu.ZeroPage()
	if cpu.X > value {
		cpu.P = setBit(cpu.P, 0)
	}
	if cpu.X == value {
		cpu.P = setBit(cpu.P, 1)
	}
}

func (cpu *CPU) CPXAbsolute() {
	address := cpu.Absolute()
	value := cpu.memory[address]
	if cpu.X > value {
		cpu.P = setBit(cpu.P, 0)
	}
	if cpu.A == value {
		cpu.P = setBit(cpu.P, 1)
	}
}

func (cpu *CPU) CPYImmediate() {
	if cpu.Y > cpu.Immediate() {
		cpu.P = setBit(cpu.P, 0)
	}
	if cpu.Y == cpu.Immediate() {
		cpu.P = setBit(cpu.P, 1)
	}
}

func (cpu *CPU) CPYZeroPage() {
	value, _ := cpu.ZeroPage()
	if cpu.Y > value {
		cpu.P = setBit(cpu.P, 0)
	}
	if cpu.Y == value {
		cpu.P = setBit(cpu.P, 1)
	}
}

func (cpu *CPU) CPYAbsolute() {
	address := cpu.Absolute()
	value := cpu.memory[address]
	if cpu.Y > value {
		cpu.P = setBit(cpu.P, 0)
	}
	if cpu.Y == value {
		cpu.P = setBit(cpu.P, 1)
	}
}

func (cpu *CPU) INCZeroPage() {
	_, address := cpu.ZeroPage()
	cpu.memory[address]++
	cpu.setZeroFlag(cpu.memory[address])
	cpu.setNegativeFlag(cpu.memory[address])
	cpu.Cycles += 2
}

func (cpu *CPU) INCZeroPageX() {
	_, address := cpu.ZeroPageX()
	cpu.memory[address]++
	cpu.setZeroFlag(cpu.memory[address])
	cpu.setNegativeFlag(cpu.memory[address])
	cpu.Cycles += 3
}

func (cpu *CPU) INCAbsolute() {
	address := cpu.Absolute()
	cpu.memory[address]++
	cpu.setZeroFlag(cpu.memory[address])
	cpu.setNegativeFlag(cpu.memory[address])
	cpu.Cycles += 2
}

func (cpu *CPU) INCAbsoluteX() {
	address := cpu.AbsoluteX()
	cpu.memory[address]++
	cpu.setZeroFlag(cpu.memory[address])
	cpu.setNegativeFlag(cpu.memory[address])
	cpu.Cycles += 3
}

func (cpu *CPU) INX() {
	cpu.X++
	cpu.setZeroFlag(cpu.X)
	cpu.setNegativeFlag(cpu.X)
	cpu.Cycles += 2

	cpu.PC++
}

func (cpu *CPU) INY() {
	cpu.Y++
	cpu.setZeroFlag(cpu.Y)
	cpu.setNegativeFlag(cpu.Y)
	cpu.Cycles += 2

	cpu.PC++
}

func (cpu *CPU) DECZeroPage() {
	_, address := cpu.ZeroPage()
	cpu.memory[address]--
	cpu.setZeroFlag(cpu.memory[address])
	cpu.setNegativeFlag(cpu.memory[address])
	cpu.Cycles += 3
	cpu.PC++
}

func (cpu *CPU) DECZeroPageX() {
	_, address := cpu.ZeroPageX()
	cpu.memory[address]--
	cpu.setZeroFlag(cpu.memory[address])
	cpu.setNegativeFlag(cpu.memory[address])
	cpu.Cycles += 3
	cpu.PC++
}

func (cpu *CPU) DECAbsolute() {
	address := cpu.Absolute()
	cpu.memory[address]--
	cpu.setZeroFlag(cpu.memory[address])
	cpu.setNegativeFlag(cpu.memory[address])
	cpu.Cycles += 2
	cpu.PC++
}

func (cpu *CPU) DECAbsoluteX() {
	address := cpu.AbsoluteX()
	cpu.memory[address]--
	cpu.setZeroFlag(cpu.memory[address])
	cpu.setNegativeFlag(cpu.memory[address])
	cpu.Cycles += 3
}

// func (cpu *CPU) DEX() {
// 	cpu.X--
// 	cpu.PC++
// 	cpu.updateZeroFlag(cpu.X)
// }

func (cpu *CPU) DEX() {
	cpu.X--
	cpu.setZeroFlag(cpu.X)
	cpu.setNegativeFlag(cpu.X)
	cpu.PC++
	cpu.Cycles += 2
}

func (cpu *CPU) updateZeroFlag(value uint8) {
	if value == 0 {
		cpu.P |= (1 << 1)
	} else {
		cpu.P &^= (1 << 1)
	}
}

func (cpu *CPU) DEY() {
	cpu.Y--
	cpu.setZeroFlag(cpu.Y)
	cpu.setNegativeFlag(cpu.Y)
	cpu.PC++
	cpu.Cycles += 2
}

func (cpu *CPU) ASLAccumulator() {
	leftbit := getBit(cpu.A, 7)
	if leftbit {
		cpu.P = setBit(cpu.P, 0)
	} else {
		cpu.P = clearBit(cpu.P, 0)
	}
	cpu.A = cpu.A << 1
	cpu.setZeroFlag(cpu.A)
	cpu.setNegativeFlag(cpu.A)
	cpu.PC++
	cpu.Cycles += 2
}

func (cpu *CPU) ASLZeroPage() {
	_, address := cpu.ZeroPage()

	leftbit := getBit(cpu.memory[address], 7)
	if leftbit {
		cpu.P = setBit(cpu.P, 0)
	} else {
		cpu.P = clearBit(cpu.P, 0)
	}
	cpu.memory[address] = cpu.memory[address] << 1
	cpu.setZeroFlag(cpu.memory[address])
	cpu.setNegativeFlag(cpu.memory[address])
}

func (cpu *CPU) ASLZeroPageX() {
	_, address := cpu.ZeroPageX()
	leftbit := getBit(cpu.memory[address], 7)
	if leftbit {
		cpu.P = setBit(cpu.P, 0)
	} else {
		cpu.P = clearBit(cpu.P, 0)
	}
	cpu.memory[address] = cpu.memory[address] << 1
	cpu.setZeroFlag(cpu.memory[address])
	cpu.setNegativeFlag(cpu.memory[address])
}

func (cpu *CPU) ASLAbsolute() {
	address := cpu.Absolute()
	leftbit := getBit(cpu.memory[address], 7)
	if leftbit {
		cpu.P = setBit(cpu.P, 0)
	} else {
		cpu.P = clearBit(cpu.P, 0)
	}
	cpu.memory[address] = cpu.memory[address] << 1
	cpu.setZeroFlag(cpu.memory[address])
	cpu.setNegativeFlag(cpu.memory[address])
}

func (cpu *CPU) ASLAbsoluteX() {
	address := cpu.AbsoluteX()
	leftbit := getBit(cpu.memory[address], 7)
	if leftbit {
		cpu.P = setBit(cpu.P, 0)
	} else {
		cpu.P = clearBit(cpu.P, 0)
	}
	cpu.memory[address] = cpu.memory[address] << 1
	cpu.setZeroFlag(cpu.memory[address])
	cpu.setNegativeFlag(cpu.memory[address])
}

func (cpu *CPU) LSRAccumulator() {
	leftbit := getBit(cpu.A, 0)
	if leftbit {
		cpu.P = setBit(cpu.P, 0)
	} else {
		cpu.P = clearBit(cpu.P, 0)
	}
	cpu.A = cpu.A >> 1
	cpu.setNegativeFlag(cpu.A)
	cpu.setZeroFlag(cpu.A)
	cpu.PC++
	cpu.Cycles += 2
}

func (cpu *CPU) LSRZeroPage() {
	_, address := cpu.ZeroPage()
	leftbit := getBit(cpu.memory[address], 0)
	if leftbit {
		cpu.P = setBit(cpu.P, 0)
	} else {
		cpu.P = clearBit(cpu.P, 0)
	}
	cpu.memory[address] = cpu.memory[address] >> 1
	cpu.setNegativeFlag(cpu.memory[address])
	cpu.setZeroFlag(cpu.memory[address])
}

func (cpu *CPU) LSRZeroPageX() {
	_, address := cpu.ZeroPageX()
	leftbit := getBit(cpu.memory[address], 0)
	if leftbit {
		cpu.P = setBit(cpu.P, 0)
	} else {
		cpu.P = clearBit(cpu.P, 0)
	}
	cpu.memory[address] = cpu.memory[address] >> 1
	cpu.setNegativeFlag(cpu.memory[address])
	cpu.setZeroFlag(cpu.memory[address])
}

func (cpu *CPU) LSRAbsolute() {
	address := cpu.Absolute()
	leftbit := getBit(cpu.memory[address], 0)
	if leftbit {
		cpu.P = setBit(cpu.P, 0)
	} else {
		cpu.P = clearBit(cpu.P, 0)
	}
	cpu.memory[address] = cpu.memory[address] >> 1
	cpu.setNegativeFlag(cpu.memory[address])
	cpu.setZeroFlag(cpu.memory[address])
}

func (cpu *CPU) LSRAbsoluteX() {
	address := cpu.Absolute()
	leftbit := getBit(cpu.memory[address], 0)
	if leftbit {
		cpu.P = setBit(cpu.P, 0)
	} else {
		cpu.P = clearBit(cpu.P, 0)
	}
	cpu.memory[address] = cpu.memory[address] >> 1
	cpu.setNegativeFlag(cpu.memory[address])
	cpu.setZeroFlag(cpu.memory[address])
}

func (cpu *CPU) ROLAccumulator() {
	leftbit := getBit(cpu.A, 7)
	cpu.A = cpu.A << 1
	if getBit(cpu.P, 0) {
		cpu.A = setBit(cpu.A, 0)
	} else {
		cpu.A = clearBit(cpu.A, 0)
	}
	if leftbit {
		cpu.P = setBit(cpu.P, 0)
	} else {
		cpu.P = clearBit(cpu.P, 0)
	}
	cpu.setZeroFlag(cpu.A)
	cpu.setNegativeFlag(cpu.A)
	cpu.PC++
	cpu.Cycles += 2
}

func (cpu *CPU) ROLZeroPage() {
	value, address := cpu.ZeroPage()
	leftbit := getBit(value, 7)
	cpu.memory[address] = cpu.memory[address] << 1
	if getBit(cpu.P, 0) {
		cpu.memory[address] = setBit(cpu.memory[address], 0)
	} else {
		cpu.A = clearBit(cpu.memory[address], 0)
	}
	if leftbit {
		cpu.P = setBit(cpu.P, 0)
	} else {
		cpu.P = clearBit(cpu.P, 0)
	}
	cpu.setZeroFlag(cpu.A)
	cpu.setNegativeFlag(cpu.A)
	cpu.Cycles += 2
}

func (cpu *CPU) ROLZeroPageX() {
	value, address := cpu.ZeroPageX()
	leftbit := getBit(value, 7)
	cpu.memory[address] = cpu.memory[address] << 1
	if getBit(cpu.P, 0) {
		cpu.memory[address] = setBit(cpu.memory[address], 0)
	} else {
		cpu.A = clearBit(cpu.memory[address], 0)
	}
	if leftbit {
		cpu.P = setBit(cpu.P, 0)
	} else {
		cpu.P = clearBit(cpu.P, 0)
	}
	cpu.setZeroFlag(cpu.A)
	cpu.setNegativeFlag(cpu.A)
	cpu.Cycles += 3
}

func (cpu *CPU) ROLAbsolute() {
	address := cpu.Absolute()
	value := cpu.memory[address]
	leftbit := getBit(value, 7)
	cpu.memory[address] = cpu.memory[address] << 1
	if getBit(cpu.P, 0) {
		cpu.memory[address] = setBit(cpu.memory[address], 0)
	} else {
		cpu.A = clearBit(cpu.memory[address], 0)
	}
	if leftbit {
		cpu.P = setBit(cpu.P, 0)
	} else {
		cpu.P = clearBit(cpu.P, 0)
	}
	cpu.setZeroFlag(cpu.A)
	cpu.setNegativeFlag(cpu.A)
	cpu.Cycles += 3
}

func (cpu *CPU) ROLAbsoluteX() {
	address := cpu.AbsoluteX()
	value := cpu.memory[address]
	leftbit := getBit(value, 7)
	cpu.memory[address] = cpu.memory[address] << 1
	if getBit(cpu.P, 0) {
		cpu.memory[address] = setBit(cpu.memory[address], 0)
	} else {
		cpu.A = clearBit(cpu.memory[address], 0)
	}
	if leftbit {
		cpu.P = setBit(cpu.P, 0)
	} else {
		cpu.P = clearBit(cpu.P, 0)
	}
	cpu.setZeroFlag(cpu.A)
	cpu.setNegativeFlag(cpu.A)
	cpu.Cycles += 4
}

func (cpu *CPU) RORAccumulator() {
	leftbit := getBit(cpu.A, 0)
	cpu.A = cpu.A >> 1
	if getBit(cpu.P, 0) {
		cpu.A = setBit(cpu.A, 7)
	} else {
		cpu.A = clearBit(cpu.A, 7)
	}
	if leftbit {
		cpu.P = setBit(cpu.P, 0)
	} else {
		cpu.P = clearBit(cpu.P, 0)
	}
	cpu.setZeroFlag(cpu.A)
	cpu.setNegativeFlag(cpu.A)
	cpu.PC++
	cpu.Cycles += 2
}

func (cpu *CPU) RORZeroPage() {
	value, address := cpu.ZeroPage()
	leftbit := getBit(value, 0)
	cpu.memory[address] = cpu.memory[address] >> 1
	if getBit(cpu.P, 0) {
		cpu.memory[address] = setBit(cpu.memory[address], 7)
	} else {
		cpu.memory[address] = clearBit(cpu.memory[address], 7)
	}
	if leftbit {
		cpu.P = setBit(cpu.P, 0)
	} else {
		cpu.P = clearBit(cpu.P, 0)
	}
	cpu.setZeroFlag(cpu.A)
	cpu.setNegativeFlag(cpu.A)
	cpu.Cycles += 2
}

func (cpu *CPU) RORZeroPageX() {
	value, address := cpu.ZeroPageX()
	leftbit := getBit(value, 0)
	cpu.memory[address] = cpu.memory[address] >> 1
	if getBit(cpu.P, 0) {
		cpu.memory[address] = setBit(cpu.memory[address], 7)
	} else {
		cpu.memory[address] = clearBit(cpu.memory[address], 7)
	}
	if leftbit {
		cpu.P = setBit(cpu.P, 0)
	} else {
		cpu.P = clearBit(cpu.P, 0)
	}
	cpu.setZeroFlag(cpu.A)
	cpu.setNegativeFlag(cpu.A)
	cpu.Cycles += 3
}

func (cpu *CPU) RORAbsolute() {
	address := cpu.Absolute()
	value := cpu.memory[address]
	leftbit := getBit(value, 0)
	cpu.memory[address] = cpu.memory[address] >> 1
	if getBit(cpu.P, 0) {
		cpu.memory[address] = setBit(cpu.memory[address], 7)
	} else {
		cpu.memory[address] = clearBit(cpu.memory[address], 7)
	}
	if leftbit {
		cpu.P = setBit(cpu.P, 0)
	} else {
		cpu.P = clearBit(cpu.P, 0)
	}
	cpu.setZeroFlag(cpu.A)
	cpu.setNegativeFlag(cpu.A)
	cpu.Cycles += 3
}

func (cpu *CPU) RORAbsoluteX() {
	address := cpu.AbsoluteX()
	value := cpu.memory[address]
	leftbit := getBit(value, 0)
	cpu.memory[address] = cpu.memory[address] >> 1
	if getBit(cpu.P, 0) {
		cpu.memory[address] = setBit(cpu.memory[address], 7)
	} else {
		cpu.memory[address] = clearBit(cpu.memory[address], 7)
	}
	if leftbit {
		cpu.P = setBit(cpu.P, 0)
	} else {
		cpu.P = clearBit(cpu.P, 0)
	}
	cpu.setZeroFlag(cpu.A)
	cpu.setNegativeFlag(cpu.A)
	cpu.Cycles += 4
}

func (cpu *CPU) JMPAbsolute() {
	lowByte := uint16(cpu.memory[cpu.PC+1])
	highByte := uint16(cpu.memory[cpu.PC+2])
	address := (highByte << 8) | lowByte
	cpu.PC = address
	cpu.Cycles += 3
}

func (cpu *CPU) JMPIndirect() {
	address := cpu.Indirect()
	cpu.PC = address
}

func (cpu *CPU) JSRAbsolute() {
	lowByte := uint16(cpu.memory[cpu.PC+1])
	highByte := uint16(cpu.memory[cpu.PC+2])
	targetAddress := (highByte << 8) | lowByte
	returnAddress := cpu.PC + 1
	cpu.memory[0x100|uint16(cpu.SP)] = byte((returnAddress >> 8) & 0xFF)
	cpu.SP--
	cpu.memory[0x100|uint16(cpu.SP)] = byte(returnAddress & 0xFF)
	cpu.SP--
	cpu.PC = targetAddress
	cpu.Cycles += 6
}

func (cpu *CPU) RTS() {
	cpu.SP++
	lowByte := uint16(cpu.memory[0x100|uint16(cpu.SP)])
	cpu.SP++
	highByte := uint16(cpu.memory[0x100|uint16(cpu.SP)])
	// Combine the low and high bytes to get the full return address
	returnAddress := (highByte << 8) | lowByte
	// Set the program counter to the return address + 1 (minus one is accounted for here)
	cpu.PC = returnAddress + 2
	// Increment the cycle count
	cpu.Cycles += 6
}

func (cpu *CPU) BCC() {
	offset := cpu.Relativetest()
	cpu.Cycles += 2
	if !getBit(cpu.P, 0) {

		cpu.PC = uint16(int32(cpu.PC) + 2 + int32(offset))
		cpu.Cycles++
	} else {
		cpu.PC += 2
	}
}

func (cpu *CPU) BCS() {
	offset := cpu.Relativetest()
	cpu.Cycles += 2
	if getBit(cpu.P, 0) {

		cpu.PC = uint16(int32(cpu.PC) + 2 + int32(offset))
		// cpu.PC += targetAddress
		cpu.Cycles++
	} else {
		cpu.PC += 2
	}
}

func (cpu *CPU) BEQ() {
	offset := cpu.Relativetest()
	cpu.Cycles += 2
	if getBit(cpu.P, 1) {
		cpu.PC = uint16(int32(cpu.PC) + 2 + int32(offset))
		cpu.Cycles++
	} else {
		cpu.PC += 2
	}
}

func (cpu *CPU) BMI() {
	offset := cpu.Relativetest()
	cpu.Cycles += 2
	if !getBit(cpu.P, 1) {

		cpu.PC = uint16(int32(cpu.PC) + 2 + int32(offset))
		cpu.Cycles++
	} else {
		cpu.PC += 2
	}
}

func (cpu *CPU) BNE() {
	offset := cpu.Relativetest()
	cpu.Cycles += 2
	if !getBit(cpu.P, 1) {
		// cpu.PC += uint16(targetAddress)
		cpu.PC = uint16(int32(cpu.PC) + 2 + int32(offset))
		cpu.Cycles++
	} else {
		cpu.PC += 2
	}
}

func (cpu *CPU) BPL() {
	targetAddress := cpu.Relative()
	cpu.Cycles += 2
	if !getBit(cpu.P, 7) {
		cpu.PC += targetAddress
		cpu.Cycles++
	} else {
		cpu.PC += 2
	}
}

func (cpu *CPU) BVC() {
	offset := cpu.Relativetest()
	cpu.Cycles += 2
	if !getBit(cpu.P, 6) {
		cpu.PC = uint16(int32(cpu.PC) + 2 + int32(offset))
		cpu.Cycles++
	} else {
		cpu.PC += 2
	}
}

func (cpu *CPU) BVS() {
	targetAddress := cpu.Relative()
	cpu.Cycles += 2
	if getBit(cpu.P, 6) {
		cpu.PC += targetAddress
		cpu.Cycles++
	} else {
		cpu.PC += 2
	}
}

func (cpu *CPU) CLC() {
	cpu.P = clearBit(cpu.P, 0)
	cpu.PC++
	cpu.Cycles += 2
}

func (cpu *CPU) CLD() {
	cpu.P = clearBit(cpu.P, 3)
	cpu.PC++
	cpu.Cycles += 2
}

func (cpu *CPU) CLI() {
	cpu.P = clearBit(cpu.P, 2)
	cpu.PC++
	cpu.Cycles += 2
}

func (cpu *CPU) CLV() {
	cpu.P = clearBit(cpu.P, 6)
	cpu.PC++
	cpu.Cycles += 2
}

func (cpu *CPU) SEC() {
	cpu.P = setBit(cpu.P, 0)
	cpu.PC++
	cpu.Cycles += 2
}

func (cpu *CPU) ISCAbsolute() {
	address := cpu.Absolute()
	cpu.memory[address] = cpu.memory[address] + 1
	cpu.sbc(cpu.memory[address])
	cpu.Cycles += 3
}

func (cpu *CPU) ISCAbsoluteX() {
	address := cpu.AbsoluteX()
	cpu.memory[address] = cpu.memory[address] + 1
	cpu.sbc(cpu.memory[address])
	cpu.Cycles += 3
}

func (cpu *CPU) ISCAbsoluteY() {
	address := cpu.AbsoluteY()
	cpu.memory[address] = cpu.memory[address] + 1
	cpu.sbc(cpu.memory[address])
	cpu.Cycles += 3
}

func (cpu *CPU) ISCZeroPage() {
	_, address := cpu.ZeroPage()
	cpu.memory[address] = cpu.memory[address] + 1
	cpu.sbc(cpu.memory[address])
	cpu.Cycles += 2
}

func (cpu *CPU) ISCZeroPageX() {
	_, address := cpu.ZeroPageX()
	cpu.memory[address] = cpu.memory[address] + 1
	cpu.sbc(cpu.memory[address])
	cpu.Cycles += 3
}

func (cpu *CPU) ISCIndirectIndex() {
	_, address := cpu.IndirectIndex()
	cpu.memory[address] = cpu.memory[address] + 1
	cpu.sbc(cpu.memory[address])
	cpu.Cycles += 3
}

func (cpu *CPU) ISCIndexIndirect() {
	_, address := cpu.IndexedIndirect()
	cpu.memory[address] = cpu.memory[address] + 1
	cpu.sbc(cpu.memory[address])
	cpu.Cycles += 3
}

func (cpu *CPU) sbc(value uint8) {
	oldCarry := uint8(0)
	if getBit(cpu.P, 0) {
		oldCarry = 1
	}
	value = value - (1 - oldCarry)

	cpu.setCarryFlag(cpu.A, value)
	cpu.setSUBOverflowFlag(uint(cpu.A), (uint(value)))
	cpu.A = cpu.A - value
	cpu.setZeroFlag(cpu.A)
	cpu.setNegativeFlag(cpu.A)
}

func (cpu *CPU) SED() {
	cpu.P = setBit(cpu.P, 3)
	cpu.PC++
	cpu.Cycles += 2
}

func (cpu *CPU) SEI() {
	cpu.P = setBit(cpu.P, 2)
	cpu.PC++
	cpu.Cycles += 2
}

func (cpu *CPU) BRK() {
	// Increment PC to point to the next instruction
	cpu.PC++
	// Push high byte of the PC onto the stack
	cpu.memory[0x100|uint16(cpu.SP)] = byte((cpu.PC >> 8) & 0xFF)
	cpu.SP--
	// Push low byte of the PC onto the stack
	cpu.memory[0x100|uint16(cpu.SP)] = byte(cpu.PC & 0xFF)
	cpu.SP--
	// Push the status register onto the stack with the break flag set
	cpu.memory[0x100|uint16(cpu.SP)] = cpu.P | (1 << 4)
	cpu.SP--
	// Set the break flag in the status register
	cpu.P |= (1 << 4)
	// Load the IRQ interrupt vector into the PC
	lowByte := uint16(cpu.memory[0xFFFE])
	highByte := uint16(cpu.memory[0xFFFF])
	cpu.PC = (highByte << 8) | lowByte
	// // Increment the cycle count
	cpu.Cycles += 7
}

func (cpu *CPU) NOP() {
	cpu.PC++
}

type instructionFunc func(*CPU)

func (cpu *CPU) ExecuteInstruction(opcode uint8) {
	// Map opcodes to instruction functions
	opcodeTable := map[uint8]instructionFunc{
		0xA9: (*CPU).LDAImmediate, // LDA Immediate
		0xA2: (*CPU).LDXImmediate, // LDX Immediate
		0x09: (*CPU).ORAImmediate, // ORA Immediate
		0xEA: (*CPU).NOP,          // NOP
		0x1A: (*CPU).NOP,          // NOP
		0xAD: (*CPU).LDAAbsolute,
		0xA5: (*CPU).LDAZeroPage,
		0xBD: (*CPU).LDAAbsoluteX,
		0xB9: (*CPU).LDAAbsoluteY,
		0xB5: (*CPU).LDAZeroPageX,
		0xA1: (*CPU).LDAIndexIndirect,
		0xB1: (*CPU).LDAIndirectIndex,
		0xAE: (*CPU).LDXAbsolute,  // LDX Absolute
		0xA6: (*CPU).LDXZeroPage,  // LDX Zero Page
		0xBE: (*CPU).LDXAbsoluteY, // LDX Absolute, Y
		0xB6: (*CPU).LDXZeroPageY, // LDX Zero Page, Y
		0xAC: (*CPU).LDYAbsolute,  // LDY Absolute
		0xA4: (*CPU).LDYZeroPage,  // LDY Zero Page
		0xA0: (*CPU).LDYImmediate, // LDY Immediate
		0xBC: (*CPU).LDYAbsoluteX, // LDY Absolute, X
		0xB4: (*CPU).LDYZeroPageX, // LDY Zero Page, X
		0x8D: (*CPU).STAAbsolute,
		0x85: (*CPU).STAZeroPage,
		0x9D: (*CPU).STAAbsoluteX,
		0x99: (*CPU).STAAbsoluteY,
		0x95: (*CPU).STAZeroPageX,
		0x81: (*CPU).STAIndexIndirect,
		0x91: (*CPU).STAIndirectIndex,
		0x8E: (*CPU).STXAbsolute,
		0x86: (*CPU).STXZeroPage,
		0x96: (*CPU).STXZeroPageY,
		0x8C: (*CPU).STYAbsolute,
		0x84: (*CPU).STYZeroPage,
		0x94: (*CPU).STYZeroPageX,
		0x6D: (*CPU).ADCAbsolute,
		0x65: (*CPU).ADCZeroPage,
		0x69: (*CPU).ADCImmediate,
		0x7D: (*CPU).ADCAbsoluteX,
		0x79: (*CPU).ADCAbsoluteY,
		0x75: (*CPU).ADCZeroPageX,
		0x61: (*CPU).ADCIndexIndirect,
		0x71: (*CPU).ADCIndirectIndex,
		0xED: (*CPU).SBCAbsolute,
		0xE5: (*CPU).SBCZeroPage,
		0xE9: (*CPU).SBCImmediate,
		0xFD: (*CPU).SBCAbsoluteX,
		0xF9: (*CPU).SBCAbsoluteY,
		0xF5: (*CPU).SBCZeroPageX,
		0xE1: (*CPU).SBCIndexIndirect,
		0xF1: (*CPU).SBCIndirectIndex,
		0xEE: (*CPU).INCAbsolute,
		0xE6: (*CPU).INCZeroPage,
		0xFE: (*CPU).INCAbsoluteX,
		0xF6: (*CPU).INCZeroPageX,
		0xE8: (*CPU).INX,
		0xC8: (*CPU).INY,
		0xCE: (*CPU).DECAbsolute,
		0xC6: (*CPU).DECZeroPage,
		0xDE: (*CPU).DECAbsoluteX,
		0xD6: (*CPU).DECZeroPageX,
		0xCA: (*CPU).DEX,
		0x88: (*CPU).DEY,
		0xAA: (*CPU).TAX,
		0xA8: (*CPU).TAY,
		0x8A: (*CPU).TXA,
		0x98: (*CPU).TYA,
		0x2D: (*CPU).ANDAbsolute,
		0x25: (*CPU).ANDZeroPage,
		0x29: (*CPU).ANDImmediate,
		0x3D: (*CPU).ANDAbsoluteX,
		0x39: (*CPU).ANDAbsoluteY,
		0x35: (*CPU).ANDZeroPageX,
		0x21: (*CPU).ANDIndexIndirect,
		0x31: (*CPU).ANDIndirectIndex,
		0x4D: (*CPU).EORAbsolute,
		0x45: (*CPU).EORZeroPage,
		0x49: (*CPU).EORImmediate,
		0x5D: (*CPU).EORAbsoluteX,
		0x59: (*CPU).EORAbsoluteY,
		0x55: (*CPU).EORZeroPageX,
		0x41: (*CPU).EORIndexIndirect,
		0x51: (*CPU).EORIndirectIndex,
		0x0D: (*CPU).ORAAbsolute,
		0x05: (*CPU).ORAZeroPage,
		0x1D: (*CPU).ORAAbsoluteX,
		0x19: (*CPU).ORAAbsoluteY,
		0x15: (*CPU).ORAZeroPageX,
		0x01: (*CPU).ORAIndexIndirect,
		0x11: (*CPU).ORAIndirectIndex,
		0xCD: (*CPU).CMPAbsolute,
		0xC5: (*CPU).CMPZeroPage,
		0xC9: (*CPU).CMPImmediate,
		0xDD: (*CPU).CMPAbsoluteX,
		0xD9: (*CPU).CMPAbsoluteY,
		0xD5: (*CPU).CMPZeroPageX,
		0xC1: (*CPU).CMPIndexedIndirect,
		0xD1: (*CPU).CMPIndirectIndirect,
		0xEC: (*CPU).CPXAbsolute,
		0xE4: (*CPU).CPXZeroPage,
		0xE0: (*CPU).CPXImmediate,
		0xCC: (*CPU).CPYAbsolute,
		0xC4: (*CPU).CPYZeroPage,
		0xC0: (*CPU).CPYImmediate,
		0x2C: (*CPU).BITAbsolute,
		0x24: (*CPU).BITZeroPage,
		0x0E: (*CPU).ASLAbsolute,
		0x06: (*CPU).ASLZeroPage,
		0x0A: (*CPU).ASLAccumulator,
		0x1E: (*CPU).ASLAbsoluteX,
		0x16: (*CPU).ASLZeroPageX,
		0x4E: (*CPU).LSRAbsolute,
		0x46: (*CPU).LSRZeroPage,
		0x4A: (*CPU).LSRAccumulator,
		0x5E: (*CPU).LSRAbsoluteX,
		0x56: (*CPU).LSRZeroPageX,
		0x2E: (*CPU).ROLAbsolute,
		0x26: (*CPU).ROLZeroPage,
		0x2A: (*CPU).ROLAccumulator,
		0x3E: (*CPU).ROLAbsoluteX,
		0x36: (*CPU).ROLZeroPageX,
		0x6E: (*CPU).RORAbsolute,
		0x66: (*CPU).RORZeroPage,
		0x6A: (*CPU).RORAccumulator,
		0x7E: (*CPU).RORAbsoluteX,
		0x76: (*CPU).RORZeroPageX,
		0x90: (*CPU).BCC,
		0xB0: (*CPU).BCS,
		0xF0: (*CPU).BEQ,
		0x30: (*CPU).BMI,
		0xD0: (*CPU).BNE,
		0x10: (*CPU).BPL,
		0x50: (*CPU).BVC,
		0x70: (*CPU).BVS,
		0xBA: (*CPU).TSX,
		0x9A: (*CPU).TXS,
		0x48: (*CPU).PHA,
		0x08: (*CPU).PHP,
		0x68: (*CPU).PLA,
		0x28: (*CPU).PLP,
		0x18: (*CPU).CLC,
		0xD8: (*CPU).CLD,
		0x58: (*CPU).CLI,
		0xB8: (*CPU).CLV,
		0x38: (*CPU).SEC,
		0xF8: (*CPU).SED,
		0x78: (*CPU).SEI,
		0x20: (*CPU).JSRAbsolute,
		0x60: (*CPU).RTS,
		0x00: (*CPU).BRK,
		0x4C: (*CPU).JMPAbsolute,
		0x6c: (*CPU).JMPIndirect,
		0xE7: (*CPU).ISCZeroPage,
		0xF7: (*CPU).ISCZeroPageX,
		0xEF: (*CPU).ISCAbsolute,
		0xFB: (*CPU).ISCAbsoluteX,
		0xE3: (*CPU).ISCIndirectIndex,
		0xF3: (*CPU).ISCIndexIndirect,
		// 0x40: (*CPU).RT,
		// 0xEA: (*CPU).NOP,
	}

	// Look up the instruction function for the given opcode
	if instrFunc, exists := opcodeTable[opcode]; exists {
		instrFunc(cpu) // Execute the instruction
	} else {
		panic(fmt.Sprintf("Unhandled opcode: %02X\n", opcode))
	}
}

const headerSize = 16

func (cpu *CPU) LoadProgram(program []uint8, startAddress uint16) {
	for i, data := range program {
		cpu.memory[startAddress+uint16(i)] = data
	}
	cpu.PC = startAddress // Set the program counter to the start of our program
}

func readNESFile(filename string) ([]uint8, error) {
	// Open the file
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()

	// Read the entire file
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	// Check if it's a valid NES file (should start with "NES\x1A")
	if len(data) < 16 || string(data[:4]) != "NES\x1A" {
		return nil, fmt.Errorf("not a valid NES file")
	}

	// Get the size of PRG-ROM in 16KB units
	prgRomSize := int(data[4]) * 16384

	// Check if the file is long enough to contain the PRG-ROM
	if len(data) < 16+prgRomSize {
		return nil, fmt.Errorf("file is too short to contain the expected PRG-ROM data")
	}

	// Extract the PRG-ROM data
	prgRom := data[16 : 16+prgRomSize]

	return prgRom, nil
}

func (cpu *CPU) LoadNESROM(prgROM []uint8) {
	for i, data := range prgROM {
		cpu.memory[0x8000+i] = data
		if len(prgROM) == 16384 {
			cpu.memory[0xC000+i] = data
		}
	}
}

func main() {
	cpu := &CPU{}

	program, _ := readNESFile("nestest.nes")

	cpu.LoadNESROM(program)
	cpu.PC = 0x8000
	cpu.P = 0x24
	i := 0
	for {
		opcode := cpu.memory[cpu.PC]
		fmt.Printf("Opcode: 0x%02X at PC: 0x%04X\n", opcode, cpu.PC)
		cpu.ExecuteInstruction(opcode)
		// if i == 500 {
		// 	break
		// }
		fmt.Printf("Step %d: PC: 0x%04X, A: %d, X: 0x%02X, Y: 0x%02X, P: 0x%02X \n",
			i+1, cpu.PC, cpu.A, cpu.X, cpu.Y, cpu.P)
		i++
		if opcode == 0x00 {
			fmt.Println("BREAK")
			break
		}
		fmt.Println("---")
	}
}
