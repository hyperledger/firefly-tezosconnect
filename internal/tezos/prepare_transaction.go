package tezos

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"blockwatch.cc/tzgo/codec"
	"blockwatch.cc/tzgo/contract"
	"blockwatch.cc/tzgo/micheline"
	"blockwatch.cc/tzgo/rpc"
	"blockwatch.cc/tzgo/tezos"
	"github.com/OneOf-Inc/firefly-tezosconnect/internal/msgs"
	"github.com/hyperledger/firefly-common/pkg/fftypes"
	"github.com/hyperledger/firefly-common/pkg/i18n"
	"github.com/hyperledger/firefly-transaction-manager/pkg/ffcapi"
)

// FFI schema types
const (
	_jsonBoolean = "boolean"
	_jsonInteger = "integer"
	_jsonNumber  = "number"
	_jsonString  = "string"
	_jsonArray   = "array"
	_jsonObject  = "object"
)

// Tezos data types
const (
	_internalBoolean = "boolean"
	_internalList    = "list"
	_internalStruct  = "struct"
	_internalInteger = "integer"
	_internalNat     = "nat"
	_internalString  = "string"
	_internalVariant = "variant"
	_internalOption  = "option"
	_internalAddress = "address"
	_internalBytes   = "bytes"
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

	op, err := c.buildOp(ctx, params, req.From, req.To, req.Nonce)
	if err != nil {
		return nil, ffcapi.ErrorReasonInvalidInputs, err
	}

	return &ffcapi.TransactionPrepareResponse{
		Gas:             req.Gas,
		TransactionData: hex.EncodeToString(op.Bytes()),
	}, "", nil
}

func (c *tezosConnector) buildOp(ctx context.Context, params micheline.Parameters, fromString, toString string, nonce *fftypes.FFBigInt) (*codec.Op, error) {
	op := codec.NewOp()

	toAddress, err := tezos.ParseAddress(toString)
	if err != nil {
		return nil, i18n.NewError(ctx, msgs.MsgInvalidToAddress, toString, err)
	}

	txArgs := contract.TxArgs{}
	txArgs.WithParameters(params)
	txArgs.WithDestination(toAddress)
	op.WithContents(txArgs.Encode())

	c.completeOp(ctx, op, fromString, nonce)

	return op, nil
}

func (c *tezosConnector) completeOp(ctx context.Context, op *codec.Op, fromString string, nonce *fftypes.FFBigInt) error {
	fromAddress, err := tezos.ParseAddress(fromString)
	if err != nil {
		return i18n.NewError(ctx, msgs.MsgInvalidFromAddress, fromString, err)
	}
	op.WithSource(fromAddress)

	hash, _ := c.client.GetBlockHash(ctx, rpc.Head)
	op.WithBranch(hash)

	op.WithParams(getNetworkParamsByName(c.networkName))

	// assign nonce
	nextCounter := nonce.Int64()
	for _, op := range op.Contents {
		// skip non-manager ops
		if op.GetCounter() < 0 {
			continue
		}
		op.WithCounter(nextCounter)
		nextCounter++
	}

	return nil
}

func getNetworkParamsByName(name string) *tezos.Params {
	switch strings.ToLower(name) {
	case "ghostnet":
		return tezos.GhostnetParams
	case "mumbainet":
		return tezos.MumbainetParams
	case "nairobinet":
		return tezos.NairobinetParams
	default:
		return tezos.DefaultParams
	}
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
	if rootType.(string) != _jsonArray {
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
	propType := argDef["type"].(string)
	details := argDef["details"].(map[string]interface{})
	name := argDef["name"]
	if name == nil {
		return resp, fmt.Errorf("property definitions of the \"prefixItems\" in the payload schema must have a \"name\"")
	}

	entry := argsMap[name.(string)]
	entryStrValue, ok := entry.(string)
	if !ok {
		return resp, errors.New("invalid object passed")
	}

	err := json.Unmarshal([]byte(entryStrValue), &entry)
	if err != nil {
		return resp, err
	}

	if propType == _jsonArray {
		resp = micheline.NewSeq()
		for _, item := range entry.([]interface{}) {
			prop, err := processMichelson(item, details)
			if err != nil {
				return resp, err
			}

			resp.Args = append(resp.Args, prop)
		}
	} else {
		resp, err = processMichelson(entry, details)
		if err != nil {
			return resp, err
		}
	}

	return resp, nil
}

func processMichelson(entry interface{}, details map[string]interface{}) (micheline.Prim, error) {
	resp := micheline.Prim{}
	var err error

	if details["type"] == "schema" {
		internalSchema := details["internalSchema"].(map[string]interface{})
		resp, err = processSchemaEntry(entry, internalSchema)
	} else {
		internalType := details["internalType"].(string)
		resp, err = processPrimitive(entry, internalType)

		propKind := details["kind"].(string)
		resp = applyKind(resp, propKind)
	}

	return resp, err
}

func processSchemaEntry(entry interface{}, schema map[string]interface{}) (micheline.Prim, error) {
	resp := micheline.Prim{}
	var err error
	entryType := schema["type"].(string)
	switch schema["type"].(string) {
	case _internalStruct:
		schemaArgs := schema["args"].([]interface{})

		var rightPairElem *micheline.Prim
		for i := len(schemaArgs) - 1; i >= 0; i-- {
			arg := schemaArgs[i].(map[string]interface{})

			argName := arg["name"].(string)
			elem := entry.(map[string]interface{})
			if _, ok := elem[argName]; !ok {
				return resp, errors.New("Schema field '" + argName + "' wasn't found")
			}

			processedEntry, err := processSchemaEntry(elem[argName], arg)
			if err != nil {
				return resp, err
			}
			newPair := forgePair(processedEntry, rightPairElem)
			rightPairElem = &newPair

			resp = newPair
		}
	case _internalList:
		schemaArgs := schema["args"].([]interface{})

		for i := len(schemaArgs) - 1; i >= 0; i-- {
			arg := schemaArgs[i].(map[string]interface{})

			listResp := micheline.NewSeq()
			for _, listElem := range entry.([]interface{}) {
				processedEntry, err := processSchemaEntry(listElem, arg)
				if err != nil {
					return resp, err
				}
				listResp.Args = append(listResp.Args, processedEntry)
			}
			resp = listResp
		}
	case _internalVariant:
		schemaArgs := schema["args"].([]interface{})
		arg := schemaArgs[0].(map[string]interface{})
		elem := entry.(map[string]interface{})

		variants := schema["variants"].([]interface{})
		for i, variant := range variants {
			if el, ok := elem[variant.(string)]; ok {
				processedEntry, err := processSchemaEntry(el, arg)
				if err != nil {
					return resp, err
				}
				if len(variants) == 1 || len(variants) > 4 {
					return resp, errors.New("wrong number of variants")
				}
				resp = wrapWithVariant(processedEntry, i+1, len(variants))
				break
			}
		}
	default:
		resp, err = processPrimitive(entry, entryType)
	}

	return resp, err
}

// TODO: think about an algorithm to support any number of variants.
// at the moment, support for up to 4 variants covers most cases
func wrapWithVariant(elem micheline.Prim, variantPos int, totalVariantsCount int) micheline.Prim {
	res := micheline.Prim{}
	if totalVariantsCount == 2 {
		branch := micheline.D_LEFT
		if variantPos == 2 {
			branch = micheline.D_RIGHT
		}
		res = micheline.NewCode(
			branch,
			elem,
		)
	} else if totalVariantsCount == 3 {
		switch variantPos {
		case 1:
			res = micheline.NewCode(
				micheline.D_LEFT,
				elem,
			)
		case 2:
			res = micheline.NewCode(
				micheline.D_RIGHT,
				micheline.NewCode(
					micheline.D_LEFT,
					elem,
				),
			)
		case 3:
			res = micheline.NewCode(
				micheline.D_RIGHT,
				micheline.NewCode(
					micheline.D_RIGHT,
					elem,
				),
			)
		}
	} else if totalVariantsCount == 4 {
		switch variantPos {
		case 1:
			res = micheline.NewCode(
				micheline.D_LEFT,
				micheline.NewCode(
					micheline.D_LEFT,
					elem,
				),
			)
		case 2:
			res = micheline.NewCode(
				micheline.D_LEFT,
				micheline.NewCode(
					micheline.D_RIGHT,
					elem,
				),
			)
		case 3:
			res = micheline.NewCode(
				micheline.D_RIGHT,
				micheline.NewCode(
					micheline.D_LEFT,
					elem,
				),
			)
		case 4:
			res = micheline.NewCode(
				micheline.D_RIGHT,
				micheline.NewCode(
					micheline.D_RIGHT,
					elem,
				),
			)
		}
	}

	return res
}

func forgePair(leftElem micheline.Prim, rightElem *micheline.Prim) micheline.Prim {
	if rightElem == nil {
		return leftElem
	}
	return micheline.NewPair(leftElem, *rightElem)
}

func processPrimitive(entry interface{}, propType string) (micheline.Prim, error) {
	resp := micheline.Prim{}
	switch propType {
	case _internalInteger, _internalNat:
		entryValue, ok := entry.(float64)
		if !ok {
			return resp, errors.New("invalid object passed")
		}

		resp = micheline.NewInt64(int64(entryValue))
	case _internalString:
		arg, ok := entry.(string)
		if !ok {
			return resp, errors.New("invalid object passed")
		}

		resp = micheline.NewString(arg)
	case _internalBytes:
		entryValue, ok := entry.(string)
		if !ok {
			return resp, errors.New("invalid object passed")
		}

		resp = micheline.NewBytes([]byte(entryValue))
	case _internalBoolean:
		entryValue, ok := entry.(bool)
		if !ok {
			return resp, errors.New("invalid object passed")
		}

		opCode := micheline.D_FALSE
		if entryValue {
			opCode = micheline.D_TRUE
		}

		resp = micheline.NewPrim(opCode)
	case _internalAddress:
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
	case _internalOption:
		return micheline.NewOption(param)
	}
	return param
}
