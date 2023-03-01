package main

import (
	"errors"
	"fmt"

	"github.com/simonvetter/modbus"
)

func GetDeviceModbusData(mc *modbus.ModbusClient, registerConfigs []RegisterConfig) ([]RegisterValue, error) {
	var rva []RegisterValue
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

	for _, register := range registerConfigs {
		rv := RegisterValue{
			ParamName: register.ParamName,
		}

		switch register.Size {
		case "UINT16":
			reg16, err = mc.ReadRegister(uint16(register.HoldingRegister), modbus.HOLDING_REGISTER)
			if err == nil {
				rv.Value = fmt.Sprintf(register.Format, float32(reg16)*register.Multiplier)
			} else {
				return nil, errors.New("Modbus error: " + err.Error())
			}
		case "SINT16":
			reg16, err = mc.ReadRegister(uint16(register.HoldingRegister), modbus.HOLDING_REGISTER)
			if err == nil {
				regs16 = int16(reg16)
				rv.Value = fmt.Sprintf(register.Format, float32(regs16)*register.Multiplier)
			} else {
				return nil, errors.New("Modbus error: " + err.Error())
			}
		case "UINT32":
			reg32, err = mc.ReadUint32(uint16(register.HoldingRegister), modbus.HOLDING_REGISTER)
			if err == nil {
				rv.Value = fmt.Sprintf(register.Format, float32(reg32)*register.Multiplier)
			} else {
				return nil, errors.New("Modbus error: " + err.Error())
			}
		case "SINT32":
			reg32, err = mc.ReadUint32(uint16(register.HoldingRegister), modbus.HOLDING_REGISTER)
			if err == nil {
				regs32 = int32(reg32)
				rv.Value = fmt.Sprintf(register.Format, float32(regs32)*register.Multiplier)
			} else {
				return nil, errors.New("Modbus error: " + err.Error())
			}
		case "UINT64":
			reg64, err = mc.ReadUint64(uint16(register.HoldingRegister), modbus.HOLDING_REGISTER)
			if err == nil {
				rv.Value = fmt.Sprintf(register.Format, float64(reg64)*float64(register.Multiplier))
			} else {
				return nil, errors.New("Modbus error: " + err.Error())
			}
		case "SINT64":
			reg64, err = mc.ReadUint64(uint16(register.HoldingRegister), modbus.HOLDING_REGISTER)
			if err == nil {
				regs64 = int64(reg64)
				rv.Value = fmt.Sprintf(register.Format, float64(regs64)*float64(register.Multiplier))
			} else {
				return nil, errors.New("Modbus error: " + err.Error())
			}
		}

		rva = append(rva, rv)
	}

	return rva, err
}
