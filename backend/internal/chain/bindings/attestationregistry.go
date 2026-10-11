// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package bindings

import (
	"context"
	"errors"
	"math/big"
	"strings"
	"time"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = errors.New
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
	_ = abi.ConvertType
	_ = time.Tick
	_ = context.Background
)

// AttestationRegistryAttestation is an auto generated low-level Go binding around an user-defined struct.
type AttestationRegistryAttestation struct {
	IssuerOrg [32]byte
	Subject   [32]byte
	DocHash   [32]byte
	ValidFrom uint64
	ValidTo   uint64
	Revoked   bool
}

// AttestationRegistryMetaData contains all meta data concerning the AttestationRegistry contract.
var AttestationRegistryMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"contractOrgRegistry\",\"name\":\"orgs_\",\"type\":\"address\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"attestationId\",\"type\":\"bytes32\"}],\"name\":\"AlreadyRevoked\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"attestationId\",\"type\":\"bytes32\"}],\"name\":\"AttestationExists\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"attestationId\",\"type\":\"bytes32\"}],\"name\":\"AttestationUnknown\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint64\",\"name\":\"validFrom\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"validTo\",\"type\":\"uint64\"}],\"name\":\"BadValidity\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"orgId\",\"type\":\"bytes32\"}],\"name\":\"NotAccredited\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotAdmin\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotRelayer\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"ZeroValue\",\"type\":\"error\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"attestationId\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"issuerOrg\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"subject\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"docHash\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"validFrom\",\"type\":\"uint64\"},{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"validTo\",\"type\":\"uint64\"}],\"name\":\"AttestationIssued\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"attestationId\",\"type\":\"bytes32\"}],\"name\":\"Revoked\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"attestationId\",\"type\":\"bytes32\"}],\"name\":\"attestation\",\"outputs\":[{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"issuerOrg\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"subject\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"docHash\",\"type\":\"bytes32\"},{\"internalType\":\"uint64\",\"name\":\"validFrom\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"validTo\",\"type\":\"uint64\"},{\"internalType\":\"bool\",\"name\":\"revoked\",\"type\":\"bool\"}],\"internalType\":\"structAttestationRegistry.Attestation\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"attestationId\",\"type\":\"bytes32\"}],\"name\":\"isValid\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"attestationId\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"issuerOrg\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"subject\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"docHash\",\"type\":\"bytes32\"},{\"internalType\":\"uint64\",\"name\":\"validFrom\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"validTo\",\"type\":\"uint64\"}],\"name\":\"issue\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"orgs\",\"outputs\":[{\"internalType\":\"contractOrgRegistry\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"attestationId\",\"type\":\"bytes32\"}],\"name\":\"revoke\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
	Bin: "0x60a060405234801561000f575f5ffd5b5060405161096d38038061096d83398101604081905261002e9161003f565b6001600160a01b031660805261006c565b5f6020828403121561004f575f5ffd5b81516001600160a01b0381168114610065575f5ffd5b9392505050565b6080516108c66100a75f395f8181610126015281816101fb0152818161022a015281816103a5015281816105f6015261062501526108c65ff3fe608060405234801561000f575f5ffd5b5060043610610055575f3560e01c80631a7fca861461005957806353573282146100d65780636a938567146100eb578063b75c7dc61461010e578063cab19cdc14610121575b5f5ffd5b61006c6100673660046107d8565b610160565b6040516100cd91905f60c08201905082518252602083015160208301526040830151604083015267ffffffffffffffff606084015116606083015267ffffffffffffffff608084015116608083015260a0830151151560a083015292915050565b60405180910390f35b6100e96100e4366004610806565b6101f9565b005b6100fe6100f93660046107d8565b610589565b60405190151581526020016100cd565b6100e961011c3660046107d8565b6105f4565b6101487f000000000000000000000000000000000000000000000000000000000000000081565b6040516001600160a01b0390911681526020016100cd565b6040805160c0810182525f80825260208201819052918101829052606081018290526080810182905260a081019190915261019a826107a4565b6040805160c081018252825481526001830154602082015260028301549181019190915260039091015467ffffffffffffffff8082166060840152600160401b8204166080830152600160801b900460ff16151560a082015292915050565b7f00000000000000000000000000000000000000000000000000000000000000006001600160a01b03166391d148547f00000000000000000000000000000000000000000000000000000000000000006001600160a01b031663926d7d7f6040518163ffffffff1660e01b8152600401602060405180830381865afa158015610284573d5f5f3e3d5ffd5b505050506040513d601f19601f820116820180604052508101906102a8919061085a565b6040516001600160e01b031960e084901b1681526004810191909152336024820152604401602060405180830381865afa1580156102e8573d5f5f3e3d5ffd5b505050506040513d601f19601f8201168201806040525081019061030c9190610871565b6103295760405163c64891a560e01b815260040160405180910390fd5b851580610334575083155b8061033d575082155b1561035b57604051637c946ed760e01b815260040160405180910390fd5b5f868152602081905260409020541561038f57604051631722b9b560e31b8152600481018790526024015b60405180910390fd5b6040516383c0ce7b60e01b8152600481018690527f00000000000000000000000000000000000000000000000000000000000000006001600160a01b0316906383c0ce7b90602401602060405180830381865afa1580156103f2573d5f5f3e3d5ffd5b505050506040513d601f19601f820116820180604052508101906104169190610871565b61043657604051631c25f13360e21b815260048101869052602401610386565b8167ffffffffffffffff168167ffffffffffffffff161161047e576040516355da959560e11b815267ffffffffffffffff808416600483015282166024820152604401610386565b6040805160c081018252868152602080820187815282840187815267ffffffffffffffff80881660608601908152878216608087019081525f60a088018181528f825296819052889020965187559351600187015591516002860155905160039094018054925193511515600160801b0260ff60801b19948316600160401b026fffffffffffffffffffffffffffffffff1990941695909216949094179190911791909116179055518490869088907f3daefcddcb2091e7c1ba7d3ac7a9dc5e876d8ffaea179670674a35e16a315523906105799088908890889092835267ffffffffffffffff918216602084015216604082015260600190565b60405180910390a4505050505050565b5f8181526020819052604081208054158015906105b257506003810154600160801b900460ff16155b80156105cc5750600381015467ffffffffffffffff164210155b80156105ed57506003810154600160401b900467ffffffffffffffff164211155b9392505050565b7f00000000000000000000000000000000000000000000000000000000000000006001600160a01b03166391d148547f00000000000000000000000000000000000000000000000000000000000000006001600160a01b031663926d7d7f6040518163ffffffff1660e01b8152600401602060405180830381865afa15801561067f573d5f5f3e3d5ffd5b505050506040513d601f19601f820116820180604052508101906106a3919061085a565b6040516001600160e01b031960e084901b1681526004810191909152336024820152604401602060405180830381865afa1580156106e3573d5f5f3e3d5ffd5b505050506040513d601f19601f820116820180604052508101906107079190610871565b6107245760405163c64891a560e01b815260040160405180910390fd5b5f61072e826107a4565b6003810154909150600160801b900460ff1615610761576040516390315de160e01b815260048101839052602401610386565b60038101805460ff60801b1916600160801b17905560405182907fe5af7daed5ab2a2dc5f98d53619f05089c0c14d11a6621f6b906a2366c9a7ab3905f90a25050565b5f81815260208190526040902080546107d35760405163b515dbdd60e01b815260048101839052602401610386565b919050565b5f602082840312156107e8575f5ffd5b5035919050565b803567ffffffffffffffff811681146107d3575f5ffd5b5f5f5f5f5f5f60c0878903121561081b575f5ffd5b86359550602087013594506040870135935060608701359250610840608088016107ef565b915061084e60a088016107ef565b90509295509295509295565b5f6020828403121561086a575f5ffd5b5051919050565b5f60208284031215610881575f5ffd5b815180151581146105ed575f5ffdfea26469706673582212205160efd516f28c3e0cf9a96f8ad6a479ed30279be9425e4ce7d4a4138600411464736f6c63430008220033",
}

// AttestationRegistryABI is the input ABI used to generate the binding from.
// Deprecated: Use AttestationRegistryMetaData.ABI instead.
var AttestationRegistryABI = AttestationRegistryMetaData.ABI

// AttestationRegistryBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use AttestationRegistryMetaData.Bin instead.
var AttestationRegistryBin = AttestationRegistryMetaData.Bin

// DeployAttestationRegistry deploys a new Ethereum contract, binding an instance of AttestationRegistry to it.
func DeployAttestationRegistry(auth *bind.TransactOpts, backend bind.ContractBackend, orgs_ common.Address) (common.Address, *types.Transaction, *AttestationRegistry, error) {
	parsed, err := AttestationRegistryMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(AttestationRegistryBin), backend, orgs_)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &AttestationRegistry{AttestationRegistryCaller: AttestationRegistryCaller{contract: contract}, AttestationRegistryTransactor: AttestationRegistryTransactor{contract: contract}, AttestationRegistryFilterer: AttestationRegistryFilterer{contract: contract}}, nil
}

// AttestationRegistry is an auto generated Go binding around an Ethereum contract.
type AttestationRegistry struct {
	AttestationRegistryCaller     // Read-only binding to the contract
	AttestationRegistryTransactor // Write-only binding to the contract
	AttestationRegistryFilterer   // Log filterer for contract events
}

// AttestationRegistryCaller is an auto generated read-only Go binding around an Ethereum contract.
type AttestationRegistryCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AttestationRegistryTransactor is an auto generated write-only Go binding around an Ethereum contract.
type AttestationRegistryTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AttestationRegistryFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type AttestationRegistryFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AttestationRegistrySession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type AttestationRegistrySession struct {
	Contract     *AttestationRegistry // Generic contract binding to set the session for
	CallOpts     bind.CallOpts        // Call options to use throughout this session
	TransactOpts bind.TransactOpts    // Transaction auth options to use throughout this session
}

// AttestationRegistryCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type AttestationRegistryCallerSession struct {
	Contract *AttestationRegistryCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts              // Call options to use throughout this session
}

// AttestationRegistryTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type AttestationRegistryTransactorSession struct {
	Contract     *AttestationRegistryTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts              // Transaction auth options to use throughout this session
}

// AttestationRegistryRaw is an auto generated low-level Go binding around an Ethereum contract.
type AttestationRegistryRaw struct {
	Contract *AttestationRegistry // Generic contract binding to access the raw methods on
}

// AttestationRegistryCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type AttestationRegistryCallerRaw struct {
	Contract *AttestationRegistryCaller // Generic read-only contract binding to access the raw methods on
}

// AttestationRegistryTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type AttestationRegistryTransactorRaw struct {
	Contract *AttestationRegistryTransactor // Generic write-only contract binding to access the raw methods on
}

// NewAttestationRegistry creates a new instance of AttestationRegistry, bound to a specific deployed contract.
func NewAttestationRegistry(address common.Address, backend bind.ContractBackend) (*AttestationRegistry, error) {
	contract, err := bindAttestationRegistry(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &AttestationRegistry{AttestationRegistryCaller: AttestationRegistryCaller{contract: contract}, AttestationRegistryTransactor: AttestationRegistryTransactor{contract: contract}, AttestationRegistryFilterer: AttestationRegistryFilterer{contract: contract}}, nil
}

// NewAttestationRegistryCaller creates a new read-only instance of AttestationRegistry, bound to a specific deployed contract.
func NewAttestationRegistryCaller(address common.Address, caller bind.ContractCaller) (*AttestationRegistryCaller, error) {
	contract, err := bindAttestationRegistry(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &AttestationRegistryCaller{contract: contract}, nil
}

// NewAttestationRegistryTransactor creates a new write-only instance of AttestationRegistry, bound to a specific deployed contract.
func NewAttestationRegistryTransactor(address common.Address, transactor bind.ContractTransactor) (*AttestationRegistryTransactor, error) {
	contract, err := bindAttestationRegistry(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &AttestationRegistryTransactor{contract: contract}, nil
}

// NewAttestationRegistryFilterer creates a new log filterer instance of AttestationRegistry, bound to a specific deployed contract.
func NewAttestationRegistryFilterer(address common.Address, filterer bind.ContractFilterer) (*AttestationRegistryFilterer, error) {
	contract, err := bindAttestationRegistry(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &AttestationRegistryFilterer{contract: contract}, nil
}

// bindAttestationRegistry binds a generic wrapper to an already deployed contract.
func bindAttestationRegistry(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := AttestationRegistryMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_AttestationRegistry *AttestationRegistryRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _AttestationRegistry.Contract.AttestationRegistryCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_AttestationRegistry *AttestationRegistryRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _AttestationRegistry.Contract.AttestationRegistryTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_AttestationRegistry *AttestationRegistryRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _AttestationRegistry.Contract.AttestationRegistryTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_AttestationRegistry *AttestationRegistryCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _AttestationRegistry.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_AttestationRegistry *AttestationRegistryTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _AttestationRegistry.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_AttestationRegistry *AttestationRegistryTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _AttestationRegistry.Contract.contract.Transact(opts, method, params...)
}

// Attestation is a free data retrieval call binding the contract method 0x1a7fca86.
//
// Solidity: function attestation(bytes32 attestationId) view returns((bytes32,bytes32,bytes32,uint64,uint64,bool))
func (_AttestationRegistry *AttestationRegistryCaller) Attestation(opts *bind.CallOpts, attestationId [32]byte) (AttestationRegistryAttestation, error) {
	var out []interface{}
	err := _AttestationRegistry.contract.Call(opts, &out, "attestation", attestationId)

	if err != nil {
		return *new(AttestationRegistryAttestation), err
	}

	out0 := *abi.ConvertType(out[0], new(AttestationRegistryAttestation)).(*AttestationRegistryAttestation)

	return out0, err

}

// Attestation is a free data retrieval call binding the contract method 0x1a7fca86.
//
// Solidity: function attestation(bytes32 attestationId) view returns((bytes32,bytes32,bytes32,uint64,uint64,bool))
func (_AttestationRegistry *AttestationRegistrySession) Attestation(attestationId [32]byte) (AttestationRegistryAttestation, error) {
	return _AttestationRegistry.Contract.Attestation(&_AttestationRegistry.CallOpts, attestationId)
}

// Attestation is a free data retrieval call binding the contract method 0x1a7fca86.
//
// Solidity: function attestation(bytes32 attestationId) view returns((bytes32,bytes32,bytes32,uint64,uint64,bool))
func (_AttestationRegistry *AttestationRegistryCallerSession) Attestation(attestationId [32]byte) (AttestationRegistryAttestation, error) {
	return _AttestationRegistry.Contract.Attestation(&_AttestationRegistry.CallOpts, attestationId)
}

// IsValid is a free data retrieval call binding the contract method 0x6a938567.
//
// Solidity: function isValid(bytes32 attestationId) view returns(bool)
func (_AttestationRegistry *AttestationRegistryCaller) IsValid(opts *bind.CallOpts, attestationId [32]byte) (bool, error) {
	var out []interface{}
	err := _AttestationRegistry.contract.Call(opts, &out, "isValid", attestationId)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsValid is a free data retrieval call binding the contract method 0x6a938567.
//
// Solidity: function isValid(bytes32 attestationId) view returns(bool)
func (_AttestationRegistry *AttestationRegistrySession) IsValid(attestationId [32]byte) (bool, error) {
	return _AttestationRegistry.Contract.IsValid(&_AttestationRegistry.CallOpts, attestationId)
}

// IsValid is a free data retrieval call binding the contract method 0x6a938567.
//
// Solidity: function isValid(bytes32 attestationId) view returns(bool)
func (_AttestationRegistry *AttestationRegistryCallerSession) IsValid(attestationId [32]byte) (bool, error) {
	return _AttestationRegistry.Contract.IsValid(&_AttestationRegistry.CallOpts, attestationId)
}

// Orgs is a free data retrieval call binding the contract method 0xcab19cdc.
//
// Solidity: function orgs() view returns(address)
func (_AttestationRegistry *AttestationRegistryCaller) Orgs(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _AttestationRegistry.contract.Call(opts, &out, "orgs")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Orgs is a free data retrieval call binding the contract method 0xcab19cdc.
//
// Solidity: function orgs() view returns(address)
func (_AttestationRegistry *AttestationRegistrySession) Orgs() (common.Address, error) {
	return _AttestationRegistry.Contract.Orgs(&_AttestationRegistry.CallOpts)
}

// Orgs is a free data retrieval call binding the contract method 0xcab19cdc.
//
// Solidity: function orgs() view returns(address)
func (_AttestationRegistry *AttestationRegistryCallerSession) Orgs() (common.Address, error) {
	return _AttestationRegistry.Contract.Orgs(&_AttestationRegistry.CallOpts)
}

// Issue is a paid mutator transaction binding the contract method 0x53573282.
//
// Solidity: function issue(bytes32 attestationId, bytes32 issuerOrg, bytes32 subject, bytes32 docHash, uint64 validFrom, uint64 validTo) returns()
func (_AttestationRegistry *AttestationRegistryTransactor) Issue(opts *bind.TransactOpts, attestationId [32]byte, issuerOrg [32]byte, subject [32]byte, docHash [32]byte, validFrom uint64, validTo uint64) (*types.Transaction, error) {
	return _AttestationRegistry.contract.Transact(opts, "issue", attestationId, issuerOrg, subject, docHash, validFrom, validTo)
}

// Issue is a paid mutator transaction binding the contract method 0x53573282.
//
// Solidity: function issue(bytes32 attestationId, bytes32 issuerOrg, bytes32 subject, bytes32 docHash, uint64 validFrom, uint64 validTo) returns()
func (_AttestationRegistry *AttestationRegistrySession) Issue(attestationId [32]byte, issuerOrg [32]byte, subject [32]byte, docHash [32]byte, validFrom uint64, validTo uint64) (*types.Transaction, error) {
	return _AttestationRegistry.Contract.Issue(&_AttestationRegistry.TransactOpts, attestationId, issuerOrg, subject, docHash, validFrom, validTo)
}

// Issue is a paid mutator transaction binding the contract method 0x53573282.
//
// Solidity: function issue(bytes32 attestationId, bytes32 issuerOrg, bytes32 subject, bytes32 docHash, uint64 validFrom, uint64 validTo) returns()
func (_AttestationRegistry *AttestationRegistryTransactorSession) Issue(attestationId [32]byte, issuerOrg [32]byte, subject [32]byte, docHash [32]byte, validFrom uint64, validTo uint64) (*types.Transaction, error) {
	return _AttestationRegistry.Contract.Issue(&_AttestationRegistry.TransactOpts, attestationId, issuerOrg, subject, docHash, validFrom, validTo)
}

// Revoke is a paid mutator transaction binding the contract method 0xb75c7dc6.
//
// Solidity: function revoke(bytes32 attestationId) returns()
func (_AttestationRegistry *AttestationRegistryTransactor) Revoke(opts *bind.TransactOpts, attestationId [32]byte) (*types.Transaction, error) {
	return _AttestationRegistry.contract.Transact(opts, "revoke", attestationId)
}

// Revoke is a paid mutator transaction binding the contract method 0xb75c7dc6.
//
// Solidity: function revoke(bytes32 attestationId) returns()
func (_AttestationRegistry *AttestationRegistrySession) Revoke(attestationId [32]byte) (*types.Transaction, error) {
	return _AttestationRegistry.Contract.Revoke(&_AttestationRegistry.TransactOpts, attestationId)
}

// Revoke is a paid mutator transaction binding the contract method 0xb75c7dc6.
//
// Solidity: function revoke(bytes32 attestationId) returns()
func (_AttestationRegistry *AttestationRegistryTransactorSession) Revoke(attestationId [32]byte) (*types.Transaction, error) {
	return _AttestationRegistry.Contract.Revoke(&_AttestationRegistry.TransactOpts, attestationId)
}

// AttestationRegistryAttestationIssuedIterator is returned from FilterAttestationIssued and is used to iterate over the raw logs and unpacked data for AttestationIssued events raised by the AttestationRegistry contract.
type AttestationRegistryAttestationIssuedIterator struct {
	Event *AttestationRegistryAttestationIssued // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *AttestationRegistryAttestationIssuedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AttestationRegistryAttestationIssued)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(AttestationRegistryAttestationIssued)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *AttestationRegistryAttestationIssuedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AttestationRegistryAttestationIssuedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AttestationRegistryAttestationIssued represents a AttestationIssued event raised by the AttestationRegistry contract.
type AttestationRegistryAttestationIssued struct {
	AttestationId [32]byte
	IssuerOrg     [32]byte
	Subject       [32]byte
	DocHash       [32]byte
	ValidFrom     uint64
	ValidTo       uint64
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterAttestationIssued is a free log retrieval operation binding the contract event 0x3daefcddcb2091e7c1ba7d3ac7a9dc5e876d8ffaea179670674a35e16a315523.
//
// Solidity: event AttestationIssued(bytes32 indexed attestationId, bytes32 indexed issuerOrg, bytes32 indexed subject, bytes32 docHash, uint64 validFrom, uint64 validTo)
func (_AttestationRegistry *AttestationRegistryFilterer) FilterAttestationIssued(opts *bind.FilterOpts, attestationId [][32]byte, issuerOrg [][32]byte, subject [][32]byte) (*AttestationRegistryAttestationIssuedIterator, error) {

	var attestationIdRule []interface{}
	for _, attestationIdItem := range attestationId {
		attestationIdRule = append(attestationIdRule, attestationIdItem)
	}
	var issuerOrgRule []interface{}
	for _, issuerOrgItem := range issuerOrg {
		issuerOrgRule = append(issuerOrgRule, issuerOrgItem)
	}
	var subjectRule []interface{}
	for _, subjectItem := range subject {
		subjectRule = append(subjectRule, subjectItem)
	}

	logs, sub, err := _AttestationRegistry.contract.FilterLogs(opts, "AttestationIssued", attestationIdRule, issuerOrgRule, subjectRule)
	if err != nil {
		return nil, err
	}
	return &AttestationRegistryAttestationIssuedIterator{contract: _AttestationRegistry.contract, event: "AttestationIssued", logs: logs, sub: sub}, nil
}

// WatchAttestationIssued is a free log subscription operation binding the contract event 0x3daefcddcb2091e7c1ba7d3ac7a9dc5e876d8ffaea179670674a35e16a315523.
//
// Solidity: event AttestationIssued(bytes32 indexed attestationId, bytes32 indexed issuerOrg, bytes32 indexed subject, bytes32 docHash, uint64 validFrom, uint64 validTo)
func (_AttestationRegistry *AttestationRegistryFilterer) WatchAttestationIssued(opts *bind.WatchOpts, sink chan<- *AttestationRegistryAttestationIssued, attestationId [][32]byte, issuerOrg [][32]byte, subject [][32]byte) (event.Subscription, error) {

	var attestationIdRule []interface{}
	for _, attestationIdItem := range attestationId {
		attestationIdRule = append(attestationIdRule, attestationIdItem)
	}
	var issuerOrgRule []interface{}
	for _, issuerOrgItem := range issuerOrg {
		issuerOrgRule = append(issuerOrgRule, issuerOrgItem)
	}
	var subjectRule []interface{}
	for _, subjectItem := range subject {
		subjectRule = append(subjectRule, subjectItem)
	}

	logs, sub, err := _AttestationRegistry.contract.WatchLogs(opts, "AttestationIssued", attestationIdRule, issuerOrgRule, subjectRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AttestationRegistryAttestationIssued)
				if err := _AttestationRegistry.contract.UnpackLog(event, "AttestationIssued", log); err != nil {
					// If the signature doesn't match, skip this log.
					if errors.Is(err, bind.ErrEventSignatureMismatch) {
						continue
					}
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseAttestationIssued is a log parse operation binding the contract event 0x3daefcddcb2091e7c1ba7d3ac7a9dc5e876d8ffaea179670674a35e16a315523.
//
// Solidity: event AttestationIssued(bytes32 indexed attestationId, bytes32 indexed issuerOrg, bytes32 indexed subject, bytes32 docHash, uint64 validFrom, uint64 validTo)
func (_AttestationRegistry *AttestationRegistryFilterer) ParseAttestationIssued(log types.Log) (*AttestationRegistryAttestationIssued, error) {
	event := new(AttestationRegistryAttestationIssued)
	if err := _AttestationRegistry.contract.UnpackLog(event, "AttestationIssued", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AttestationRegistryRevokedIterator is returned from FilterRevoked and is used to iterate over the raw logs and unpacked data for Revoked events raised by the AttestationRegistry contract.
type AttestationRegistryRevokedIterator struct {
	Event *AttestationRegistryRevoked // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *AttestationRegistryRevokedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AttestationRegistryRevoked)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(AttestationRegistryRevoked)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *AttestationRegistryRevokedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AttestationRegistryRevokedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AttestationRegistryRevoked represents a Revoked event raised by the AttestationRegistry contract.
type AttestationRegistryRevoked struct {
	AttestationId [32]byte
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterRevoked is a free log retrieval operation binding the contract event 0xe5af7daed5ab2a2dc5f98d53619f05089c0c14d11a6621f6b906a2366c9a7ab3.
//
// Solidity: event Revoked(bytes32 indexed attestationId)
func (_AttestationRegistry *AttestationRegistryFilterer) FilterRevoked(opts *bind.FilterOpts, attestationId [][32]byte) (*AttestationRegistryRevokedIterator, error) {

	var attestationIdRule []interface{}
	for _, attestationIdItem := range attestationId {
		attestationIdRule = append(attestationIdRule, attestationIdItem)
	}

	logs, sub, err := _AttestationRegistry.contract.FilterLogs(opts, "Revoked", attestationIdRule)
	if err != nil {
		return nil, err
	}
	return &AttestationRegistryRevokedIterator{contract: _AttestationRegistry.contract, event: "Revoked", logs: logs, sub: sub}, nil
}

// WatchRevoked is a free log subscription operation binding the contract event 0xe5af7daed5ab2a2dc5f98d53619f05089c0c14d11a6621f6b906a2366c9a7ab3.
//
// Solidity: event Revoked(bytes32 indexed attestationId)
func (_AttestationRegistry *AttestationRegistryFilterer) WatchRevoked(opts *bind.WatchOpts, sink chan<- *AttestationRegistryRevoked, attestationId [][32]byte) (event.Subscription, error) {

	var attestationIdRule []interface{}
	for _, attestationIdItem := range attestationId {
		attestationIdRule = append(attestationIdRule, attestationIdItem)
	}

	logs, sub, err := _AttestationRegistry.contract.WatchLogs(opts, "Revoked", attestationIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AttestationRegistryRevoked)
				if err := _AttestationRegistry.contract.UnpackLog(event, "Revoked", log); err != nil {
					// If the signature doesn't match, skip this log.
					if errors.Is(err, bind.ErrEventSignatureMismatch) {
						continue
					}
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseRevoked is a log parse operation binding the contract event 0xe5af7daed5ab2a2dc5f98d53619f05089c0c14d11a6621f6b906a2366c9a7ab3.
//
// Solidity: event Revoked(bytes32 indexed attestationId)
func (_AttestationRegistry *AttestationRegistryFilterer) ParseRevoked(log types.Log) (*AttestationRegistryRevoked, error) {
	event := new(AttestationRegistryRevoked)
	if err := _AttestationRegistry.contract.UnpackLog(event, "Revoked", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
