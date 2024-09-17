package voucher_v2

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"

	"github.com/layer-3/clearsync/pkg/abi/ivoucher_v2"
)

var (
	addressTy = must(abi.NewType("address", "", nil))
	uint32Ty  = must(abi.NewType("uint32", "", nil))
	uint64Ty  = must(abi.NewType("uint64", "", nil))
	uint128Ty = must(abi.NewType("uint128", "", nil))
	bytesTy   = must(abi.NewType("bytes", "", nil))

	voucherArgs = abi.Arguments{
		{Name: "chainId", Type: uint32Ty},
		{Name: "router", Type: addressTy},
		{Name: "executor", Type: addressTy},
		{Name: "beneficiary", Type: addressTy},
		{Name: "expireAt", Type: uint64Ty},
		{Name: "nonce", Type: uint128Ty},
		{Name: "data", Type: bytesTy},
		{Name: "signature", Type: bytesTy},
	}
)

func must[T any](x T, err error) T {
	if err != nil {
		panic(err)
	}
	return x
}

// Encode multiple vouchers into a byte slice according to Ethereum ABI.
// Useful for encoding "use" method argument and generation signature.
func EncodeMultiple(vouchers []ivoucher_v2.IVoucherVoucher) ([]byte, error) {
	voucherABI, err := ivoucher_v2.IVoucherMetaData.GetAbi()
	if err != nil {
		return nil, err
	}

	packed, err := voucherABI.Methods["use"].Inputs[0:1].Pack(vouchers)
	if err != nil {
		return nil, fmt.Errorf("failed to encode: %w", err)
	}

	return packed, nil
}

// Encode encodes the Voucher into a byte slice according to Ethereum ABI.
// TODO: not sure if this is the right way to encode the voucher.
// NOTE: to receive data to sign first encode the voucher without signature,
// then sign the data and encode it again with the signature.
func Encode(voucher ivoucher_v2.IVoucherVoucher) ([]byte, error) {
	packed, err := voucherArgs.Pack(
		voucher.ChainId,
		voucher.Router,
		voucher.Executor,
		voucher.Beneficiary,
		voucher.ExpireAt,
		voucher.Nonce,
		voucher.Data,
		voucher.Signature,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to encode: %w", err)
	}

	return packed, nil
}

// Decode decodes a byte slice into a Voucher struct according to Ethereum ABI.
func Decode(voucher []byte) (ivoucher_v2.IVoucherVoucher, error) {
	data := make(map[string]any)
	if err := voucherArgs.UnpackIntoMap(data, voucher); err != nil {
		return ivoucher_v2.IVoucherVoucher{}, fmt.Errorf("failed to decode: %w", err)
	}

	var result ivoucher_v2.IVoucherVoucher
	fields := map[string]reflect.Type{
		"chainId":     reflect.TypeOf(result.ChainId),
		"router":      reflect.TypeOf(result.Router),
		"executor":    reflect.TypeOf(result.Executor),
		"beneficiary": reflect.TypeOf(result.Beneficiary),
		"expireAt":    reflect.TypeOf(result.ExpireAt),
		"nonce":       reflect.TypeOf(result.Nonce),
		"data":        reflect.TypeOf(result.Data),
		"signature":   reflect.TypeOf(result.Signature),
	}

	valResult := reflect.ValueOf(&result).Elem()
	for key, expectedType := range fields {
		rawVal, ok := data[key]
		if !ok {
			return ivoucher_v2.IVoucherVoucher{}, fmt.Errorf("missing %s", key)
		}
		val := reflect.ValueOf(rawVal)
		if !val.Type().AssignableTo(expectedType) {
			return ivoucher_v2.IVoucherVoucher{}, fmt.Errorf("%s field has wrong type", key)
		}

		fieldVal := valResult.FieldByName(strings.Title(key))
		if !fieldVal.IsValid() {
			return ivoucher_v2.IVoucherVoucher{}, fmt.Errorf("no such field: %s in result", key)
		}
		if !fieldVal.CanSet() {
			return ivoucher_v2.IVoucherVoucher{}, fmt.Errorf("cannot set field: %s", key)
		}
		fieldVal.Set(val)
	}

	return result, nil
}