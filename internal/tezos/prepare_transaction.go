package tezos

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"

	"blockwatch.cc/tzgo/codec"
	"blockwatch.cc/tzgo/contract"
	"blockwatch.cc/tzgo/micheline"
	"blockwatch.cc/tzgo/rpc"
	"blockwatch.cc/tzgo/tezos"
	"github.com/hyperledger/firefly-common/pkg/log"
	"github.com/hyperledger/firefly-transaction-manager/pkg/ffcapi"
	jsonschema "github.com/xeipuuv/gojsonschema"
)

const (
	jsonBooleanType = "boolean"
	jsonIntegerType = "integer"
	jsonNumberType  = "number"
	jsonStringType  = "string"
	jsonArrayType   = "array"
	jsonObjectType  = "object"
)

// TransactionPrepare validates transaction inputs against the supplied schema/Michelson and performs any binary serialization required (prior to signing) to encode a transaction from JSON into the native blockchain format
func (c *tezosConnector) TransactionPrepare(ctx context.Context, req *ffcapi.TransactionPrepareRequest) (*ffcapi.TransactionPrepareResponse, ffcapi.ErrorReason, error) {
	var methodName string
	err := req.Method.Unmarshal(ctx, &methodName)
	if err != nil {
		return nil, "", err
	}

	params, err := processArgs(req, methodName)
	if err != nil {
		return nil, "", err
	}

	txArgs := contract.TxArgs{}
	txArgs.WithParameters(params)
	txArgs.WithSource(tezos.MustParseAddress(req.From))
	txArgs.WithDestination(tezos.MustParseAddress(req.To))

	op := codec.NewOp()
	op.WithSource(tezos.MustParseAddress(req.From))
	op.WithContents(txArgs.Encode())

	// Just used to do op serialization correctly.
	// The real last block hash must be set up just before broadcasting to the chain to not be outdated.
	hash, _ := c.client.GetBlockHash(ctx, rpc.Head)
	op.WithBranch(hash)

	// TODO: add gas esimation

	callData := op.Bytes()

	log.L(ctx).Infof("Prepared transaction method=%s dataLen=%d gas=%s", methodName, len(callData), req.Gas.Int())

	return &ffcapi.TransactionPrepareResponse{
		Gas:             req.Gas,
		TransactionData: hex.EncodeToString(callData),
	}, "", nil
}

func processArgs(req *ffcapi.TransactionPrepareRequest, methodName string) (micheline.Parameters, error) {
	params := micheline.Parameters{
		Entrypoint: methodName,
		Value:      micheline.NewPrim(micheline.D_UNIT),
	}

	argsMap := req.Args
	if argsMap == nil {
		return params, fmt.Errorf("must specify args")
	}

	schemaMap := req.PayloadSchema
	if schemaMap == nil {
		return params, errors.New("no payload schema provided")
	}

	rootType := schemaMap["type"]
	if rootType.(string) != "array" {
		return params, fmt.Errorf("payload schema must define a root type of \"array\"")
	}
	// we require the schema to use "prefixItems" to define the ordered array of arguments
	pitems := schemaMap["prefixItems"]
	if pitems == nil {
		return params, fmt.Errorf("payload schema must define a root type of \"array\" using \"prefixItems\"")
	}

	items := pitems.([]interface{})

	// If entrypoint doesn't accept parameters - send micheline.D_UNIT param (represents the absence of a meaningful value)
	if len(items) == 0 {
		return params, nil
	}
	if len(items) == 1 {
		michelineVal, err := convertFFIParamToMicheltonParam(argsMap, items[0])
		if err != nil {
			return params, err
		}
		params.Value = michelineVal
	} else {
		seq := micheline.NewSeq()
		for _, item := range items {
			michelineVal, err := convertFFIParamToMicheltonParam(argsMap, item)
			if err != nil {
				return params, err
			}
			seq.Args = append(seq.Args, michelineVal)
		}
		params.Value = seq
	}

	return params, nil
}

func convertFFIParamToMicheltonParam(argsMap map[string]interface{}, arg interface{}) (micheline.Prim, error) {
	resp := micheline.Prim{}
	argDef := arg.(map[string]interface{})
	name := argDef["name"]
	if name == nil {
		return resp, fmt.Errorf("property definitions of the \"prefixItems\" in the payload schema must have a \"name\"")
	}
	entry := argsMap[name.(string)]

	propType := argDef["type"].(string)

	details := argDef["details"].(map[string]interface{})
	internalType := details["internalType"].(string)

	entryStrValue, ok := entry.(string)
	if !ok {
		return resp, errors.New("invalid object passed")
	}

	err := json.Unmarshal([]byte(entryStrValue), &entry)
	if err != nil {
		return resp, err
	}

	if propType == jsonArrayType {
		resp = micheline.NewSeq()
		for _, item := range entry.([]interface{}) {
			prop, err := processPrimitive(item, internalType)
			if err != nil {
				return resp, err
			}

			resp.Args = append(resp.Args, prop)
		}
	} else {
		if internalType == "" {
			internalType = propType
		}
		resp, err = processPrimitive(entry, internalType)
		if err != nil {
			return resp, err
		}
	}

	propKind := details["kind"].(string)
	resp = applyKind(resp, propKind)

	return resp, nil
}

func processPrimitive(entry interface{}, propType string) (micheline.Prim, error) {
	resp := micheline.Prim{}
	switch propType {
	case "integer", "nat":
		entryValue, ok := entry.(float64)
		if !ok {
			return resp, errors.New("invalid object passed")
		}

		resp = micheline.NewInt64(int64(entryValue))
	case "string":
		arg, ok := entry.(string)
		if !ok {
			return resp, errors.New("invalid object passed")
		}

		resp = micheline.NewString(arg)
	case "bytes":
		entryValue, ok := entry.(string)
		if !ok {
			return resp, errors.New("invalid object passed")
		}

		resp = micheline.NewBytes([]byte(entryValue))
	case "boolean":
		entryValue, ok := entry.(bool)
		if !ok {
			return resp, errors.New("invalid object passed")
		}

		opCode := micheline.D_FALSE
		if entryValue {
			opCode = micheline.D_TRUE
		}

		resp = micheline.NewPrim(opCode)
	case "address":
		entryValue, ok := entry.(string)
		if !ok {
			return resp, errors.New("invalid object passed")
		}

		address, err := tezos.ParseAddress(entryValue)
		if err != nil {
			return resp, err
		}

		resp = micheline.NewAddress(address)
	}

	return resp, nil
}

func applyKind(param micheline.Prim, kind string) micheline.Prim {
	switch kind {
	case "option":
		return micheline.NewOption(param)
	}
	return param
}

func validate(def map[string]interface{}, name string, value interface{}) error {
	// let's use the schema to validate the body args payload first
	document := jsonschema.NewGoLoader(value)
	schemaloader := jsonschema.NewGoLoader(def)
	result, err := jsonschema.Validate(schemaloader, document)
	if err != nil {
		return fmt.Errorf("failed to validate argument \"%s\": %s", name, err)
	}
	if !result.Valid() {
		errorMsg := ""
		for _, desc := range result.Errors() {
			errorMsg = fmt.Sprintf("%s- %s\n", errorMsg, desc)
		}
		return fmt.Errorf("failed to validate argument \"%s\": %s", name, errorMsg)
	}
	return nil
}
