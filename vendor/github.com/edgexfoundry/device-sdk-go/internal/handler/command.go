// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
// Copyright (C) 2018-2020 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package handler

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/edgexfoundry/device-sdk-go/internal/cache"
	"github.com/edgexfoundry/device-sdk-go/internal/common"
	"github.com/edgexfoundry/device-sdk-go/internal/transformer"
	dsModels "github.com/edgexfoundry/device-sdk-go/pkg/models"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
)

// Note, every HTTP request to ServeHTTP is made in a separate goroutine, which
// means care needs to be taken with respect to shared data accessed through *Server.
func CommandHandler(vars map[string]string, body string, method string, queryParams string) (*dsModels.Event, common.AppError) {
	dKey := vars[common.IdVar]
	cmd := vars[common.CommandVar]

	var ok bool
	var d contract.Device
	if dKey != "" {
		d, ok = cache.Devices().ForId(dKey)
	} else {
		dKey = vars[common.NameVar]
		d, ok = cache.Devices().ForName(dKey)
	}
	if !ok {
		msg := fmt.Sprintf("Device: %s not found; %s", dKey, method)
		common.LoggingClient.Error(msg)
		return nil, common.NewNotFoundError(msg, nil)
	}

	if d.AdminState == contract.Locked {
		msg := fmt.Sprintf("%s is locked; %s", d.Name, method)
		common.LoggingClient.Error(msg)
		return nil, common.NewLockedError(msg, nil)
	}

	if d.OperatingState == contract.Disabled {
		msg := fmt.Sprintf("%s is disabled; %s", d.Name, method)
		common.LoggingClient.Error(msg)
		return nil, common.NewLockedError(msg, nil)
	}

	// TODO: need to mark device when operation in progress, so it can't be removed till completed

	cmdExists, err := cache.Profiles().CommandExists(d.Profile.Name, cmd, method)

	// TODO: once cache locking has been implemented, this should never happen
	if err != nil {
		msg := fmt.Sprintf("internal error; Device: %s searching %s in cache failed; %s", d.Name, cmd, method)
		common.LoggingClient.Error(msg)
		return nil, common.NewServerError(msg, err)
	}

	var evt *dsModels.Event = nil
	var appErr common.AppError
	if !cmdExists {
		dr, drExists := cache.Profiles().DeviceResource(d.Profile.Name, cmd)
		if !drExists {
			msg := fmt.Sprintf("%s for Device: %s not found; %s", cmd, d.Name, method)
			common.LoggingClient.Error(msg)
			return nil, common.NewNotFoundError(msg, nil)
		}

		if strings.ToLower(method) == common.GetCmdMethod {
			evt, appErr = execReadDeviceResource(&d, &dr, queryParams)
		} else {
			appErr = execWriteDeviceResource(&d, &dr, body)
		}
	} else {
		if strings.ToLower(method) == common.GetCmdMethod {
			evt, appErr = execReadCmd(&d, cmd, queryParams)
		} else {
			appErr = execWriteCmd(&d, cmd, body)
		}
	}

	go common.UpdateLastConnected(d.Name)
	return evt, appErr
}

func execReadDeviceResource(device *contract.Device, dr *contract.DeviceResource, queryParams string) (*dsModels.Event, common.AppError) {
	var reqs []dsModels.CommandRequest
	var req dsModels.CommandRequest
	common.LoggingClient.Debug(fmt.Sprintf("Handler - execReadCmd: deviceResource: %s", dr.Name))

	req.DeviceResourceName = dr.Name
	req.Attributes = dr.Attributes
	if queryParams != "" {
		if len(req.Attributes) <= 0 {
			req.Attributes = make(map[string]string)
		}
		m := common.FilterQueryParams(queryParams)
		req.Attributes[common.URLRawQuery] = m.Encode()
	}
	req.Type = dsModels.ParseValueType(dr.Properties.Value.Type)
	reqs = append(reqs, req)

	results, err := common.Driver.HandleReadCommands(device.Name, device.Protocols, reqs)
	if err != nil {
		msg := fmt.Sprintf("Handler - execReadCmd: error for Device: %s DeviceResource: %s, %v", device.Name, dr.Name, err)
		return nil, common.NewServerError(msg, err)
	}

	return cvsToEvent(device, results, dr.Name)
}

func cvsToEvent(device *contract.Device, cvs []*dsModels.CommandValue, cmd string) (*dsModels.Event, common.AppError) {
	readings := make([]contract.Reading, 0, common.CurrentConfig.Device.MaxCmdOps)
	var transformsOK = true
	var err error

	for _, cv := range cvs {
		// get the device resource associated with the rsp.RO
		dr, ok := cache.Profiles().DeviceResource(device.Profile.Name, cv.DeviceResourceName)
		if !ok {
			msg := fmt.Sprintf("Handler - execReadCmd: no deviceResource: %s for dev: %s in Command Result %v", cv.DeviceResourceName, device.Name, cv)
			common.LoggingClient.Error(msg)
			return nil, common.NewServerError(msg, nil)
		}

		if common.CurrentConfig.Device.DataTransform {
			err = transformer.TransformReadResult(cv, dr.Properties.Value)
			if err != nil {
				common.LoggingClient.Error(fmt.Sprintf("Handler - execReadCmd: CommandValue (%s) transformed failed: %v", cv.String(), err))
				transformsOK = false
			}
		}

		err = transformer.CheckAssertion(cv, dr.Properties.Value.Assertion, device)
		if err != nil {
			common.LoggingClient.Error(fmt.Sprintf("Handler - execReadCmd: Assertion failed for device resource: %s, with value: %v", cv.String(), err))
			cv = dsModels.NewStringValue(cv.DeviceResourceName, cv.Origin, fmt.Sprintf("Assertion failed for device resource, with value: %s and assertion: %s", cv.String(), dr.Properties.Value.Assertion))
		}

		ro, err := cache.Profiles().ResourceOperation(device.Profile.Name, cv.DeviceResourceName, common.GetCmdMethod)
		if err != nil {
			common.LoggingClient.Debug(fmt.Sprintf("getting resource operation failed: %s", err.Error()))
		} else if len(ro.Mappings) > 0 {
			newCV, ok := transformer.MapCommandValue(cv, ro.Mappings)
			if ok {
				cv = newCV
			} else {
				common.LoggingClient.Warn(fmt.Sprintf("Handler - execReadCmd: Resource Operation (%s) mapping value (%s) failed with the mapping table: %v", ro.DeviceCommand, cv.String(), ro.Mappings))
				//transformsOK = false  // issue #89 will discuss how to handle there is no mapping matched
			}
		}

		// TODO: the Java SDK supports a RO secondary device resource.
		// If defined, then a RO result will generate a reading for the
		// secondary device resource. As this use case isn't defined and/or used in
		// any of the existing Java device services, this concept hasn't
		// been implemened in gxds. TBD at the devices f2f whether this
		// be killed completely.

		reading := common.CommandValueToReading(cv, device.Name, dr.Properties.Value.MediaType, dr.Properties.Value.FloatEncoding)
		readings = append(readings, *reading)

		if cv.Type == dsModels.Binary {
			common.LoggingClient.Debug(fmt.Sprintf("Handler - execReadCmd: device: %s DeviceResource: %v reading: binary value", device.Name, cv.DeviceResourceName))
		} else {
			common.LoggingClient.Debug(fmt.Sprintf("Handler - execReadCmd: device: %s DeviceResource: %v reading: %v", device.Name, cv.DeviceResourceName, reading))
		}
	}

	if !transformsOK {
		msg := fmt.Sprintf("Transform failed for dev: %s cmd: %s method: GET", device.Name, cmd)
		common.LoggingClient.Error(msg)
		common.LoggingClient.Debug(fmt.Sprintf("Readings: %v", readings))
		return nil, common.NewServerError(msg, nil)
	}

	// push to Core Data
	cevent := contract.Event{Device: device.Name, Readings: readings}
	event := &dsModels.Event{Event: cevent}
	event.Origin = common.GetUniqueOrigin()

	// TODO: enforce config.MaxCmdValueLen; need to include overhead for
	// the rest of the reading JSON + Event JSON length?  Should there be
	// a separate JSON body max limit for retvals & command parameters?

	return event, nil
}

func execReadCmd(device *contract.Device, cmd string, queryParams string) (*dsModels.Event, common.AppError) {
	// make ResourceOperations
	ros, err := cache.Profiles().ResourceOperations(device.Profile.Name, cmd, common.GetCmdMethod)
	if err != nil {
		common.LoggingClient.Error(err.Error())
		return nil, common.NewNotFoundError(err.Error(), err)
	}

	if len(ros) > common.CurrentConfig.Device.MaxCmdOps {
		msg := fmt.Sprintf("Handler - execReadCmd: MaxCmdOps (%d) execeeded for dev: %s cmd: %s method: GET",
			common.CurrentConfig.Device.MaxCmdOps, device.Name, cmd)
		common.LoggingClient.Error(msg)
		return nil, common.NewServerError(msg, nil)
	}

	reqs := make([]dsModels.CommandRequest, len(ros))

	for i, op := range ros {
		drName := op.DeviceResource
		common.LoggingClient.Debug(fmt.Sprintf("Handler - execReadCmd: deviceResource: %s", drName))

		// TODO: add recursive support for resource command chaining. This occurs when a
		// deviceprofile resource command operation references another resource command
		// instead of a device resource (see BoschXDK for reference).

		dr, ok := cache.Profiles().DeviceResource(device.Profile.Name, drName)
		common.LoggingClient.Debug(fmt.Sprintf("Handler - execReadCmd: deviceResource: %v", dr))
		if !ok {
			msg := fmt.Sprintf("Handler - execReadCmd: no deviceResource: %s for dev: %s cmd: %s method: GET", drName, device.Name, cmd)
			common.LoggingClient.Error(msg)
			return nil, common.NewServerError(msg, nil)
		}

		reqs[i].DeviceResourceName = dr.Name
		reqs[i].Attributes = dr.Attributes
		if queryParams != "" {
			if len(reqs[i].Attributes) <= 0 {
				reqs[i].Attributes = make(map[string]string)
			}
			m := common.FilterQueryParams(queryParams)
			reqs[i].Attributes[common.URLRawQuery] = m.Encode()
		}
		reqs[i].Type = dsModels.ParseValueType(dr.Properties.Value.Type)
	}

	results, err := common.Driver.HandleReadCommands(device.Name, device.Protocols, reqs)
	if err != nil {
		msg := fmt.Sprintf("Handler - execReadCmd: error for Device: %s cmd: %s, %v", device.Name, cmd, err)
		return nil, common.NewServerError(msg, err)
	}

	return cvsToEvent(device, results, cmd)
}

func execWriteDeviceResource(device *contract.Device, dr *contract.DeviceResource, params string) common.AppError {
	paramMap, err := parseParams(params)
	if err != nil {
		msg := fmt.Sprintf("Handler - execWriteDeviceResource: Put parameters parsing failed: %s", params)
		common.LoggingClient.Error(msg)
		return common.NewBadRequestError(msg, err)
	}

	v, ok := paramMap[dr.Name]
	if !ok && dr.Properties.Value.DefaultValue != "" {
		v = dr.Properties.Value.DefaultValue
	} else if !ok {
		msg := fmt.Sprintf("there is no %s in parameters and no default value in DeviceResource", dr.Name)
		common.LoggingClient.Error(msg)
		return common.NewBadRequestError(msg, fmt.Errorf(msg))
	}

	cv, err := createCommandValueFromDR(dr, v)
	if err != nil {
		msg := fmt.Sprintf("Handler - execWriteDeviceResource: Put parameters parsing failed: %s", params)
		common.LoggingClient.Error(msg)
		return common.NewBadRequestError(msg, err)
	}

	reqs := make([]dsModels.CommandRequest, 1)
	common.LoggingClient.Debug(fmt.Sprintf("Handler - execWriteDeviceResource: putting deviceResource: %s", dr.Name))
	reqs[0].DeviceResourceName = cv.DeviceResourceName
	reqs[0].Attributes = dr.Attributes
	reqs[0].Type = cv.Type

	if common.CurrentConfig.Device.DataTransform {
		err = transformer.TransformWriteParameter(cv, dr.Properties.Value)
		if err != nil {
			msg := fmt.Sprintf("Handler - execWriteDeviceResource: CommandValue (%s) transformed failed: %v", cv.String(), err)
			common.LoggingClient.Error(msg)
			return common.NewServerError(msg, err)
		}
	}

	err = common.Driver.HandleWriteCommands(device.Name, device.Protocols, reqs, []*dsModels.CommandValue{cv})
	if err != nil {
		msg := fmt.Sprintf("Handler - execWriteDeviceResource: error for Device: %s Device Resource: %s, %v", device.Name, dr.Name, err)
		return common.NewServerError(msg, err)
	}

	return nil
}

func execWriteCmd(device *contract.Device, cmd string, params string) common.AppError {
	ros, err := cache.Profiles().ResourceOperations(device.Profile.Name, cmd, common.SetCmdMethod)
	if err != nil {
		msg := fmt.Sprintf("Handler - execWriteCmd: can't find ResrouceOperations in Profile(%s) and Command(%s), %v", device.Profile.Name, cmd, err)
		common.LoggingClient.Error(msg)
		return common.NewBadRequestError(msg, err)
	}

	if len(ros) > common.CurrentConfig.Device.MaxCmdOps {
		msg := fmt.Sprintf("Handler - execWriteCmd: MaxCmdOps (%d) execeeded for dev: %s cmd: %s method: PUT",
			common.CurrentConfig.Device.MaxCmdOps, device.Name, cmd)
		common.LoggingClient.Error(msg)
		return common.NewServerError(msg, nil)
	}

	cvs, err := parseWriteParams(device.Profile.Name, ros, params)
	if err != nil {
		msg := fmt.Sprintf("Handler - execWriteCmd: Put parameters parsing failed: %s", params)
		common.LoggingClient.Error(msg)
		return common.NewBadRequestError(msg, err)
	}

	reqs := make([]dsModels.CommandRequest, len(cvs))
	for i, cv := range cvs {
		drName := cv.DeviceResourceName
		common.LoggingClient.Debug(fmt.Sprintf("Handler - execWriteCmd: putting deviceResource: %s", drName))

		// TODO: add recursive support for resource command chaining. This occurs when a
		// deviceprofile resource command operation references another resource command
		// instead of a device resource (see BoschXDK for reference).

		dr, ok := cache.Profiles().DeviceResource(device.Profile.Name, drName)
		common.LoggingClient.Debug(fmt.Sprintf("Handler - execWriteCmd: putting deviceResource: %s", drName))
		if !ok {
			msg := fmt.Sprintf("Handler - execWriteCmd: no deviceResource: %s for dev: %s cmd: %s method: GET", drName, device.Name, cmd)
			common.LoggingClient.Error(msg)
			return common.NewServerError(msg, nil)
		}

		reqs[i].DeviceResourceName = cv.DeviceResourceName
		reqs[i].Attributes = dr.Attributes
		reqs[i].Type = cv.Type

		if common.CurrentConfig.Device.DataTransform {
			err = transformer.TransformWriteParameter(cv, dr.Properties.Value)
			if err != nil {
				msg := fmt.Sprintf("Handler - execWriteCmd: CommandValue (%s) transformed failed: %v", cv.String(), err)
				common.LoggingClient.Error(msg)
				return common.NewServerError(msg, err)
			}
		}
	}

	err = common.Driver.HandleWriteCommands(device.Name, device.Protocols, reqs, cvs)
	if err != nil {
		msg := fmt.Sprintf("Handler - execWriteCmd: error for Device: %s cmd: %s, %v", device.Name, cmd, err)
		return common.NewServerError(msg, err)
	}

	return nil
}

func parseWriteParams(profileName string, ros []contract.ResourceOperation, params string) ([]*dsModels.CommandValue, error) {
	paramMap, err := parseParams(params)
	if err != nil {
		return []*dsModels.CommandValue{}, err
	}

	result := make([]*dsModels.CommandValue, 0, len(paramMap))
	for _, ro := range ros {
		common.LoggingClient.Debug(fmt.Sprintf("looking for %s in the request parameters", ro.DeviceResource))
		p, ok := paramMap[ro.DeviceResource]
		if !ok {
			dr, ok := cache.Profiles().DeviceResource(profileName, ro.DeviceResource)
			if !ok {
				err := fmt.Errorf("the parameter %s does not match any DeviceResource in DeviceProfile", ro.DeviceResource)
				return []*dsModels.CommandValue{}, err
			}

			if ro.Parameter != "" {
				common.LoggingClient.Debug(fmt.Sprintf("there is no %s in the request parameters, retrieving value from the Parameter field from the ResourceOperation", ro.DeviceResource))
				p = ro.Parameter
			} else if dr.Properties.Value.DefaultValue != "" {
				common.LoggingClient.Debug(fmt.Sprintf("there is no %s in the request parameters, retrieving value from the DefaultValue field from the ValueProperty", ro.DeviceResource))
				p = dr.Properties.Value.DefaultValue
			} else {
				err := fmt.Errorf("the parameter %s is not defined in the request body and there is no default value", ro.DeviceResource)
				return []*dsModels.CommandValue{}, err
			}
		}

		if len(ro.Mappings) > 0 {
			newP, ok := ro.Mappings[p]
			if ok {
				p = newP
			} else {
				msg := fmt.Sprintf("parseWriteParams: Resource (%s) mapping value (%s) failed with the mapping table: %v", ro.DeviceResource, p, ro.Mappings)
				common.LoggingClient.Warn(msg)
				//return result, fmt.Errorf(msg) // issue #89 will discuss how to handle there is no mapping matched
			}
		}

		cv, err := createCommandValueFromRO(profileName, &ro, p)
		if err == nil {
			result = append(result, cv)
		} else {
			return result, err
		}
	}

	return result, nil
}

func parseParams(params string) (paramMap map[string]string, err error) {
	err = json.Unmarshal([]byte(params), &paramMap)
	if err != nil {
		common.LoggingClient.Error(fmt.Sprintf("parsing Write parameters failed %s, %v", params, err))
		return
	}

	if len(paramMap) == 0 {
		err = fmt.Errorf("no parameters specified")
		return
	}
	return
}

func createCommandValueFromRO(profileName string, ro *contract.ResourceOperation, v string) (*dsModels.CommandValue, error) {
	dr, ok := cache.Profiles().DeviceResource(profileName, ro.DeviceResource)
	if !ok {
		msg := fmt.Sprintf("createCommandValueForParam: no deviceResource: %s", ro.DeviceResource)
		common.LoggingClient.Error(msg)
		return nil, fmt.Errorf(msg)
	}

	return createCommandValueFromDR(&dr, v)
}

func createCommandValueFromDR(dr *contract.DeviceResource, v string) (*dsModels.CommandValue, error) {
	var result *dsModels.CommandValue
	var err error
	origin := time.Now().UnixNano()

	switch strings.ToLower(dr.Properties.Value.Type) {
	case "bool":
		value, err := strconv.ParseBool(v)
		if err != nil {
			return result, err
		}
		result, err = dsModels.NewBoolValue(dr.Name, origin, value)
	case "boolarray":
		var arr []bool
		err = json.Unmarshal([]byte(v), &arr)
		if err != nil {
			return result, err
		}
		result, err = dsModels.NewBoolArrayValue(dr.Name, origin, arr)
	case "string":
		result = dsModels.NewStringValue(dr.Name, origin, v)
	case "uint8":
		n, err := strconv.ParseUint(v, 10, 8)
		if err != nil {
			return result, err
		}
		result, err = dsModels.NewUint8Value(dr.Name, origin, uint8(n))
	case "uint8array":
		var arr []uint8
		strArr := strings.Split(strings.Trim(v, "[]"), ",")
		for _, u := range strArr {
			n, err := strconv.ParseUint(strings.Trim(u, " "), 10, 8)
			if err != nil {
				return result, err
			}
			arr = append(arr, uint8(n))
		}
		result, err = dsModels.NewUint8ArrayValue(dr.Name, origin, arr)
	case "uint16":
		n, err := strconv.ParseUint(v, 10, 16)
		if err != nil {
			return result, err
		}
		result, err = dsModels.NewUint16Value(dr.Name, origin, uint16(n))
	case "uint16array":
		var arr []uint16
		strArr := strings.Split(strings.Trim(v, "[]"), ",")
		for _, u := range strArr {
			n, err := strconv.ParseUint(strings.Trim(u, " "), 10, 16)
			if err != nil {
				return result, err
			}
			arr = append(arr, uint16(n))
		}
		result, err = dsModels.NewUint16ArrayValue(dr.Name, origin, arr)
	case "uint32":
		n, err := strconv.ParseUint(v, 10, 32)
		if err != nil {
			return result, err
		}
		result, err = dsModels.NewUint32Value(dr.Name, origin, uint32(n))
	case "uint32array":
		var arr []uint32
		strArr := strings.Split(strings.Trim(v, "[]"), ",")
		for _, u := range strArr {
			n, err := strconv.ParseUint(strings.Trim(u, " "), 10, 32)
			if err != nil {
				return result, err
			}
			arr = append(arr, uint32(n))
		}
		result, err = dsModels.NewUint32ArrayValue(dr.Name, origin, arr)
	case "uint64":
		n, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			return result, err
		}
		result, err = dsModels.NewUint64Value(dr.Name, origin, n)
	case "uint64array":
		var arr []uint64
		strArr := strings.Split(strings.Trim(v, "[]"), ",")
		for _, u := range strArr {
			n, err := strconv.ParseUint(strings.Trim(u, " "), 10, 64)
			if err != nil {
				return result, err
			}
			arr = append(arr, n)
		}
		result, err = dsModels.NewUint64ArrayValue(dr.Name, origin, arr)
	case "int8":
		n, err := strconv.ParseInt(v, 10, 8)
		if err != nil {
			return result, err
		}
		result, err = dsModels.NewInt8Value(dr.Name, origin, int8(n))
	case "int8array":
		var arr []int8
		err = json.Unmarshal([]byte(v), &arr)
		if err != nil {
			return result, err
		}
		result, err = dsModels.NewInt8ArrayValue(dr.Name, origin, arr)
	case "int16":
		n, err := strconv.ParseInt(v, 10, 16)
		if err != nil {
			return result, err
		}
		result, err = dsModels.NewInt16Value(dr.Name, origin, int16(n))
	case "int16array":
		var arr []int16
		err = json.Unmarshal([]byte(v), &arr)
		if err != nil {
			return result, err
		}
		result, err = dsModels.NewInt16ArrayValue(dr.Name, origin, arr)
	case "int32":
		n, err := strconv.ParseInt(v, 10, 32)
		if err != nil {
			return result, err
		}
		result, err = dsModels.NewInt32Value(dr.Name, origin, int32(n))
	case "int32array":
		var arr []int32
		err = json.Unmarshal([]byte(v), &arr)
		if err != nil {
			return result, err
		}
		result, err = dsModels.NewInt32ArrayValue(dr.Name, origin, arr)
	case "int64":
		n, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return result, err
		}
		result, err = dsModels.NewInt64Value(dr.Name, origin, n)
	case "int64array":
		var arr []int64
		err = json.Unmarshal([]byte(v), &arr)
		if err != nil {
			return result, err
		}
		result, err = dsModels.NewInt64ArrayValue(dr.Name, origin, arr)
	case "float32":
		n, e := strconv.ParseFloat(v, 32)
		if e == nil {
			result, err = dsModels.NewFloat32Value(dr.Name, origin, float32(n))
			break
		}
		if numError, ok := e.(*strconv.NumError); ok {
			if numError.Err == strconv.ErrRange {
				err = e
				break
			}
		}
		var decodedToBytes []byte
		decodedToBytes, err = base64.StdEncoding.DecodeString(v)
		if err == nil {
			var val float32
			val, err = float32FromBytes(decodedToBytes)
			if err != nil {
				break
			} else if math.IsNaN(float64(val)) {
				err = fmt.Errorf("fail to parse %v to float32, unexpected result %v", v, val)
			} else {
				result, err = dsModels.NewFloat32Value(dr.Name, origin, val)
			}
		}
	case "float32array":
		var arr []float32
		err = json.Unmarshal([]byte(v), &arr)
		if err != nil {
			return result, err
		}
		result, err = dsModels.NewFloat32ArrayValue(dr.Name, origin, arr)
	case "float64":
		var val float64
		val, err = strconv.ParseFloat(v, 64)
		if err == nil {
			result, err = dsModels.NewFloat64Value(dr.Name, origin, val)
			break
		}
		if numError, ok := err.(*strconv.NumError); ok {
			if numError.Err == strconv.ErrRange {
				break
			}
		}
		var decodedToBytes []byte
		decodedToBytes, err = base64.StdEncoding.DecodeString(v)
		if err == nil {
			val, err = float64FromBytes(decodedToBytes)
			if err != nil {
				break
			} else if math.IsNaN(val) {
				err = fmt.Errorf("fail to parse %v to float64, unexpected result %v", v, val)
			} else {
				result, err = dsModels.NewFloat64Value(dr.Name, origin, val)
			}
		}
	case "float64array":
		var arr []float64
		err = json.Unmarshal([]byte(v), &arr)
		if err != nil {
			return result, err
		}
		result, err = dsModels.NewFloat64ArrayValue(dr.Name, origin, arr)
	}

	if err != nil {
		common.LoggingClient.Error(fmt.Sprintf("Handler - Command: Parsing parameter value (%s) to %s failed: %v", v, dr.Properties.Value.Type, err))
		return result, err
	}

	return result, err
}

func float64FromBytes(numericValue []byte) (res float64, err error) {
	reader := bytes.NewReader(numericValue)
	err = binary.Read(reader, binary.BigEndian, &res)
	return
}

func float32FromBytes(numericValue []byte) (res float32, err error) {
	reader := bytes.NewReader(numericValue)
	err = binary.Read(reader, binary.BigEndian, &res)
	return
}

func CommandAllHandler(cmd string, body string, method string, queryParams string) ([]*dsModels.Event, common.AppError) {
	common.LoggingClient.Debug(fmt.Sprintf("Handler - CommandAll: execute the %s command %s from all operational devices", method, cmd))
	devices := filterOperationalDevices(cache.Devices().All())

	devCount := len(devices)
	var waitGroup sync.WaitGroup
	waitGroup.Add(devCount)
	cmdResults := make(chan struct {
		event  *dsModels.Event
		appErr common.AppError
	}, devCount)

	for i, _ := range devices {
		go func(device *contract.Device) {
			defer waitGroup.Done()
			var event *dsModels.Event = nil
			var appErr common.AppError = nil
			if strings.ToLower(method) == common.GetCmdMethod {
				event, appErr = execReadCmd(device, cmd, queryParams)
			} else {
				appErr = execWriteCmd(device, cmd, body)
			}
			cmdResults <- struct {
				event  *dsModels.Event
				appErr common.AppError
			}{event, appErr}
		}(devices[i])
	}
	waitGroup.Wait()
	close(cmdResults)

	errCount := 0
	getResults := make([]*dsModels.Event, 0, devCount)
	var appErr common.AppError
	for r := range cmdResults {
		if r.appErr != nil {
			errCount++
			common.LoggingClient.Error("Handler - CommandAll: " + r.appErr.Message())
			appErr = r.appErr // only the last error will be returned
		} else if r.event != nil {
			getResults = append(getResults, r.event)
		}
	}

	if errCount < devCount {
		common.LoggingClient.Info("Handler - CommandAll: part of commands executed successfully, returning 200 OK")
		appErr = nil
	}

	return getResults, appErr

}

func filterOperationalDevices(devices []contract.Device) []*contract.Device {
	result := make([]*contract.Device, 0, len(devices))
	for i, d := range devices {
		if (d.AdminState == contract.Locked) || (d.OperatingState == contract.Disabled) {
			continue
		}
		result = append(result, &devices[i])
	}
	return result
}
