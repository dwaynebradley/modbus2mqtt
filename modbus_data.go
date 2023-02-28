package main

import (
	"errors"
	"fmt"

	"github.com/simonvetter/modbus"
)

func GetDeviceModbusData(mc *modbus.ModbusClient, readings []ReadingInfo) ([]ReadingInfo, error) {
	var ria []ReadingInfo
	var err error

	var reg16 uint16
	var reg32 uint32
	var reg64 uint64
	var regs16 int16
	var regs32 int32
	var regs64 int64

	if mc == nil {
		return nil, errors.New("Modbus client not connected")
	}

	for _, reading := range readings {
		ri := reading

		switch reading.Size {
		case "UINT16":
			reg16, err = mc.ReadRegister(uint16(reading.HoldingRegister), modbus.HOLDING_REGISTER)
			if err == nil {
				ri.Value = fmt.Sprintf(reading.Format, float32(reg16)*reading.Multiplier)
			} else {
				return nil, errors.New("Modbus error: " + err.Error())
			}
		case "SINT16":
			reg16, err = mc.ReadRegister(uint16(reading.HoldingRegister), modbus.HOLDING_REGISTER)
			if err == nil {
				regs16 = int16(reg16)
				ri.Value = fmt.Sprintf(reading.Format, float32(regs16)*reading.Multiplier)
			} else {
				return nil, errors.New("Modbus error: " + err.Error())
			}
		case "UINT32":
			reg32, err = mc.ReadUint32(uint16(reading.HoldingRegister), modbus.HOLDING_REGISTER)
			if err == nil {
				ri.Value = fmt.Sprintf(reading.Format, float32(reg32)*reading.Multiplier)
			} else {
				return nil, errors.New("Modbus error: " + err.Error())
			}
		case "SINT32":
			reg32, err = mc.ReadUint32(uint16(reading.HoldingRegister), modbus.HOLDING_REGISTER)
			if err == nil {
				regs32 = int32(reg32)
				ri.Value = fmt.Sprintf(reading.Format, float32(regs32)*reading.Multiplier)
			} else {
				return nil, errors.New("Modbus error: " + err.Error())
			}
		case "UINT64":
			reg64, err = mc.ReadUint64(uint16(reading.HoldingRegister), modbus.HOLDING_REGISTER)
			if err == nil {
				ri.Value = fmt.Sprintf(reading.Format, float64(reg64)*float64(reading.Multiplier))
			} else {
				return nil, errors.New("Modbus error: " + err.Error())
			}
		case "SINT64":
			reg64, err = mc.ReadUint64(uint16(reading.HoldingRegister), modbus.HOLDING_REGISTER)
			if err == nil {
				regs64 = int64(reg64)
				ri.Value = fmt.Sprintf(reading.Format, float64(regs64)*float64(reading.Multiplier))
			} else {
				return nil, errors.New("Modbus error: " + err.Error())
			}
		}

		ria = append(ria, ri)
	}

	return ria, err
}
