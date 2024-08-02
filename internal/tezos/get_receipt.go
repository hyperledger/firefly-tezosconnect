package tezos

import (
	"context"
	"encoding/json"

	"github.com/hyperledger/firefly-common/pkg/fftypes"
	"github.com/hyperledger/firefly-common/pkg/log"
	"github.com/hyperledger/firefly-transaction-manager/pkg/ffcapi"
	"github.com/trilitech/tzgo/micheline"
	"github.com/trilitech/tzgo/rpc"
	"github.com/trilitech/tzgo/tezos"
)

const _address = "address"

type receiptExtraInfo struct {
	ContractAddress     *tezos.Address    `json:"contractAddress"`
	ConsumedGas         *fftypes.FFBigInt `json:"consumedGas"`
	GasLimit            *fftypes.FFBigInt `json:"gasLimit"`
	PaidStorageSizeDiff *fftypes.FFBigInt `json:"paidStorageSizeDiff"`
	StorageSize         *fftypes.FFBigInt `json:"storageSize"`
	StorageLimit        *fftypes.FFBigInt `json:"storageLimit"`
	From                *tezos.Address    `json:"from"`
	To                  *tezos.Address    `json:"to"`
	Counter             *fftypes.FFBigInt `json:"counter"`
	Fee                 *fftypes.FFBigInt `json:"fee"`
	Status              *string           `json:"status"`
	ErrorMessage        *string           `json:"errorMessage"`
	Storage             *fftypes.JSONAny  `json:"storage"`
}

// TransactionReceipt queries to see if a receipt is available for a given transaction hash
func (c *tezosConnector) TransactionReceipt(ctx context.Context, req *ffcapi.TransactionReceiptRequest) (*ffcapi.TransactionReceiptResponse, ffcapi.ErrorReason, error) {
	// ensure block observer is running
	rpcClient := c.client.(*rpc.Client)
	rpcClient.Listen()

	// wait for confirmations
	res := rpc.NewResult(tezos.MustParseOpHash(req.TransactionHash)) // .WithTTL(op.TTL).WithConfirmations(opts.Confirmations)
	res.Listen(rpcClient.BlockObserver)
	res.WaitContext(ctx)
	if err := res.Err(); err != nil {
		return nil, "", err
	}

	// return receipt
	receipt, err := res.GetReceipt(ctx)
	if err != nil {
		return nil, "", err
	}

	blockNumber := receipt.Block.Int64()
	block, _, err := c.BlockInfoByHash(ctx, &ffcapi.BlockInfoByHashRequest{
		BlockHash: receipt.Block.String(),
	})
	if err != nil {
		log.L(ctx).Error("error getting block: ", err)
	} else {
		blockNumber = block.BlockNumber.Int64()
	}

	receiptResponse := &ffcapi.TransactionReceiptResponse{
		TransactionReceiptResponseBase: ffcapi.TransactionReceiptResponseBase{
			BlockNumber:      fftypes.NewFFBigInt(blockNumber),
			TransactionIndex: fftypes.NewFFBigInt(int64(receipt.Pos)),
			BlockHash:        receipt.Block.String(),
			Success:          receipt.IsSuccess(),
			ProtocolID:       receipt.Op.Protocol.String(),
		},
	}

	if receipt.Op != nil {
		var fullReceipt []byte
		operationReceipts := make([]receiptExtraInfo, 0, len(receipt.Op.Contents))

		for _, o := range receipt.Op.Contents {
			if o.Kind() == tezos.OpTypeTransaction {
				tx := o.(*rpc.Transaction)

				txStatus := tx.Result().Status.String()
				extraInfo := receiptExtraInfo{
					ConsumedGas:         fftypes.NewFFBigInt(tx.Metadata.Result.ConsumedMilliGas / 1000),
					GasLimit:            fftypes.NewFFBigInt(tx.GasLimit),
					PaidStorageSizeDiff: fftypes.NewFFBigInt(tx.Metadata.Result.PaidStorageSizeDiff),
					StorageSize:         fftypes.NewFFBigInt(tx.Metadata.Result.StorageSize),
					StorageLimit:        fftypes.NewFFBigInt(tx.StorageLimit),
					From:                &tx.Source,
					To:                  &tx.Destination,
					Counter:             fftypes.NewFFBigInt(tx.Counter),
					Fee:                 fftypes.NewFFBigInt(tx.Fee),
					Status:              &txStatus,
				}

				var script *micheline.Script
				if tx.Destination.IsContract() {
					location, _ := json.Marshal(map[string]string{
						_address: tx.Destination.String(),
					})
					receiptResponse.ContractLocation = fftypes.JSONAnyPtrBytes(location)
					extraInfo.ContractAddress = &tx.Destination

					script, err = c.client.GetContractScript(ctx, tx.Destination)
					if err != nil {
						log.L(ctx).Error("error getting contract script: ", err)
					}
				}
				if len(tx.Result().Errors) > 0 {
					errorMessage := ""
					for _, err := range tx.Result().Errors {
						errorMessage += err.Error()
					}
					extraInfo.ErrorMessage = &errorMessage
				}
				if prim := tx.Metadata.Result.Storage; prim != nil {
					val := micheline.NewValue(script.StorageType(), *prim)
					m, err := val.Map()
					if err != nil {
						log.L(ctx).Error("error parsing contract storage: ", err)
					}
					storageBytes, _ := json.Marshal(m)
					extraInfo.Storage = fftypes.JSONAnyPtrBytes(storageBytes)
				}

				operationReceipts = append(operationReceipts, extraInfo)
				fullReceipt, _ = json.Marshal(operationReceipts)
			} else if o.Kind() == tezos.OpTypeOrigination {
				result := o.(*rpc.Origination).Result()
				originatedContracts := result.OriginatedContracts
				if len(originatedContracts) > 0 {
					location, _ := json.Marshal(map[string]string{
						_address: originatedContracts[0].ContractAddress(),
					})
					receiptResponse.ContractLocation = fftypes.JSONAnyPtrBytes(location)
				}
				fullReceipt = c.extraInfoForDeployTransactionReceipt(ctx, result, operationReceipts)
			}
		}

		receiptResponse.ExtraInfo = fftypes.JSONAnyPtrBytes(fullReceipt)
	}

	return receiptResponse, "", nil
}

func (c *tezosConnector) extraInfoForDeployTransactionReceipt(ctx context.Context, res rpc.OperationResult, operationReceipts []receiptExtraInfo) []byte {
	status := res.Status.String()
	extraInfo := receiptExtraInfo{
		ConsumedGas:         fftypes.NewFFBigInt(res.ConsumedMilliGas / 1000),
		StorageSize:         fftypes.NewFFBigInt(res.StorageSize),
		PaidStorageSizeDiff: fftypes.NewFFBigInt(res.PaidStorageSizeDiff),
		Status:              &status,
	}

	if len(res.Errors) > 0 {
		errorMessage := ""
		for _, err := range res.Errors {
			errorMessage += err.Error()
		}
		extraInfo.ErrorMessage = &errorMessage
	}

	if prim := res.Storage; prim != nil {
		val := micheline.NewValue(res.Storage.BuildType(), *prim)
		m, err := val.Map()
		if err != nil {
			log.L(ctx).Error("error parsing contract storage: ", err)
		}
		storageBytes, _ := json.Marshal(m)
		extraInfo.Storage = fftypes.JSONAnyPtrBytes(storageBytes)
	}

	if len(res.OriginatedContracts) > 0 {
		extraInfo.ContractAddress = &res.OriginatedContracts[0]
	}

	operationReceipts = append(operationReceipts, extraInfo)
	fullReceipt, _ := json.Marshal(operationReceipts)

	return fullReceipt
}
