package main

import (
	"errors"
	"fmt"

	"github.com/simonvetter/modbus"
)

func GetDeviceModbusData(mc *modbus.ModbusClient, registerConfigs []RegisterConfig) ([]RegisterValue, error) {
	var rva []RegisterValue
	var err error

	var u16Register uint16
	var u32Register uint32
	var u64Register uint64
	var s16Register int16
	var s32Register int32
	var s64Register int64

	if mc == nil {
		return nil, errors.New("Modbus client not connected")
	}

	for _, register := range registerConfigs {
		rv := RegisterValue{
			ParamName: register.ParamName,
		}

		switch register.Size {
		case "UINT16":
			u16Register, err = mc.ReadRegister(uint16(register.HoldingRegister), modbus.HOLDING_REGISTER)
			if err == nil {
				rv.Value = fmt.Sprintf(register.Format, float32(u16Register)*register.Multiplier)
			} else {
				return nil, errors.New("Modbus error: " + err.Error())
			}
		case "SINT16":
			u16Register, err = mc.ReadRegister(uint16(register.HoldingRegister), modbus.HOLDING_REGISTER)
			if err == nil {
				s16Register = int16(u16Register)
				rv.Value = fmt.Sprintf(register.Format, float32(s16Register)*register.Multiplier)
			} else {
				return nil, errors.New("Modbus error: " + err.Error())
			}
		case "UINT32":
			u32Register, err = mc.ReadUint32(uint16(register.HoldingRegister), modbus.HOLDING_REGISTER)
			if err == nil {
				rv.Value = fmt.Sprintf(register.Format, float32(u32Register)*register.Multiplier)
			} else {
				return nil, errors.New("Modbus error: " + err.Error())
			}
		case "SINT32":
			u32Register, err = mc.ReadUint32(uint16(register.HoldingRegister), modbus.HOLDING_REGISTER)
			if err == nil {
				s32Register = int32(u32Register)
				rv.Value = fmt.Sprintf(register.Format, float32(s32Register)*register.Multiplier)
			} else {
				return nil, errors.New("Modbus error: " + err.Error())
			}
		case "UINT64":
			u64Register, err = mc.ReadUint64(uint16(register.HoldingRegister), modbus.HOLDING_REGISTER)
			if err == nil {
				rv.Value = fmt.Sprintf(register.Format, float64(u64Register)*float64(register.Multiplier))
			} else {
				return nil, errors.New("Modbus error: " + err.Error())
			}
		case "SINT64":
			u64Register, err = mc.ReadUint64(uint16(register.HoldingRegister), modbus.HOLDING_REGISTER)
			if err == nil {
				s64Register = int64(u64Register)
				rv.Value = fmt.Sprintf(register.Format, float64(s64Register)*float64(register.Multiplier))
			} else {
				return nil, errors.New("Modbus error: " + err.Error())
			}
		}

		rva = append(rva, rv)
	}

	return rva, err
}
