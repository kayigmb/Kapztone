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

// FacilityRegistryFacility is an auto generated low-level Go binding around an user-defined struct.
type FacilityRegistryFacility struct {
	OrgId             [32]byte
	DataHash          [32]byte
	CapacityPerPeriod *big.Int
	Suspended         bool
}

// FacilityRegistryMetaData contains all meta data concerning the FacilityRegistry contract.
var FacilityRegistryMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"contractOrgRegistry\",\"name\":\"orgs_\",\"type\":\"address\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"facilityId\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"period\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"requested\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"available\",\"type\":\"uint256\"}],\"name\":\"CapacityExceeded\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"facilityId\",\"type\":\"bytes32\"}],\"name\":\"FacilityExists\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"facilityId\",\"type\":\"bytes32\"}],\"name\":\"FacilityIsSuspended\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"facilityId\",\"type\":\"bytes32\"}],\"name\":\"FacilityUnknown\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotAdmin\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotOperator\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotOriginator\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotRelayer\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"orgId\",\"type\":\"bytes32\"}],\"name\":\"OrgNotVerified\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"OriginatorAlreadySet\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"ZeroValue\",\"type\":\"error\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"facilityId\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"orgId\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"dataHash\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"capacityPerPeriod\",\"type\":\"uint256\"}],\"name\":\"FacilityRegistered\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"facilityId\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"bool\",\"name\":\"suspended\",\"type\":\"bool\"}],\"name\":\"FacilitySuspended\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"originator\",\"type\":\"address\"}],\"name\":\"OriginatorSet\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"facilityId\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"period\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"quantity\",\"type\":\"uint256\"}],\"name\":\"consumeCapacity\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"orgId\",\"type\":\"bytes32\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"facilityId\",\"type\":\"bytes32\"}],\"name\":\"facility\",\"outputs\":[{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"orgId\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"dataHash\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"capacityPerPeriod\",\"type\":\"uint256\"},{\"internalType\":\"bool\",\"name\":\"suspended\",\"type\":\"bool\"}],\"internalType\":\"structFacilityRegistry.Facility\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"orgs\",\"outputs\":[{\"internalType\":\"contractOrgRegistry\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"facilityId\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"period\",\"type\":\"uint256\"}],\"name\":\"originatedInPeriod\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"originator\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"facilityId\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"orgId\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"dataHash\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"capacityPerPeriod\",\"type\":\"uint256\"}],\"name\":\"register\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"originator_\",\"type\":\"address\"}],\"name\":\"setOriginator\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"facilityId\",\"type\":\"bytes32\"},{\"internalType\":\"bool\",\"name\":\"suspended\",\"type\":\"bool\"}],\"name\":\"setSuspended\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
	Bin: "0x60a060405234801561000f575f5ffd5b50604051610c7f380380610c7f83398101604081905261002e9161003f565b6001600160a01b031660805261006c565b5f6020828403121561004f575f5ffd5b81516001600160a01b0381168114610065575f5ffd5b9392505050565b608051610bc36100bc5f395f8181610179015281816101b0015281816101df0152818161033701528181610366015281816104e1015281816106820152818161081a01526108490152610bc35ff3fe608060405234801561000f575f5ffd5b5060043610610085575f3560e01c80636e6b15b1116100585780636e6b15b114610102578063c9b4cc6c1461014a578063cab19cdc14610174578063e7715b051461019b575f5ffd5b806329f72845146100895780633c815d9e1461009e57806342ee3d0a146100ce57806363cba5ca146100e1575b5f5ffd5b61009c610097366004610a31565b6101ae565b005b6002546100b1906001600160a01b031681565b6040516001600160a01b0390911681526020015b60405180910390f35b61009c6100dc366004610a5f565b610335565b6100f46100ef366004610a8e565b610606565b6040519081526020016100c5565b610115610110366004610ab7565b6107ac565b6040516100c5919081518152602080830151908201526040808301519082015260609182015115159181019190915260800190565b6100f4610158366004610ace565b600160209081525f928352604080842090915290825290205481565b6100b17f000000000000000000000000000000000000000000000000000000000000000081565b61009c6101a9366004610aee565b610818565b7f00000000000000000000000000000000000000000000000000000000000000006001600160a01b03166391d148547f00000000000000000000000000000000000000000000000000000000000000006001600160a01b031663f5b541a66040518163ffffffff1660e01b8152600401602060405180830381865afa158015610239573d5f5f3e3d5ffd5b505050506040513d601f19601f8201168201806040525081019061025d9190610b1b565b6040516001600160e01b031960e084901b1681526004810191909152336024820152604401602060405180830381865afa15801561029d573d5f5f3e3d5ffd5b505050506040513d601f19601f820116820180604052508101906102c19190610b32565b6102de57604051631f0853c160e21b815260040160405180910390fd5b806102e8836109ed565b600301805460ff1916911515919091179055604051811515815282907f7a67dd3e17e1b246cbb7ea1c7a3e96d62195b88cbbea41490e2b06bd59d034699060200160405180910390a25050565b7f00000000000000000000000000000000000000000000000000000000000000006001600160a01b03166391d148547f00000000000000000000000000000000000000000000000000000000000000006001600160a01b031663926d7d7f6040518163ffffffff1660e01b8152600401602060405180830381865afa1580156103c0573d5f5f3e3d5ffd5b505050506040513d601f19601f820116820180604052508101906103e49190610b1b565b6040516001600160e01b031960e084901b1681526004810191909152336024820152604401602060405180830381865afa158015610424573d5f5f3e3d5ffd5b505050506040513d601f19601f820116820180604052508101906104489190610b32565b6104655760405163c64891a560e01b815260040160405180910390fd5b831580610470575081155b80610479575080155b1561049757604051637c946ed760e01b815260040160405180910390fd5b5f84815260208190526040902054156104cb5760405163d161f13b60e01b8152600481018590526024015b60405180910390fd5b60405163c181b27360e01b8152600481018490527f00000000000000000000000000000000000000000000000000000000000000006001600160a01b03169063c181b27390602401602060405180830381865afa15801561052e573d5f5f3e3d5ffd5b505050506040513d601f19601f820116820180604052508101906105529190610b32565b61057257604051636bed6a6760e01b8152600481018490526024016104c2565b6040805160808101825284815260208082018581528284018581525f606085018181528a8252818552908690209451855591516001850155516002840155516003909201805460ff1916921515929092179091558151848152908101839052849186917fa7e7e5534486d104bc4e154993522de6fedc9e38050288ce452c5ca1235443a3910160405180910390a350505050565b6002545f906001600160a01b0316331461063357604051638977534f60e01b815260040160405180910390fd5b5f61063d856109ed565b600381015490915060ff161561066957604051630c3f1fbb60e41b8152600481018690526024016104c2565b805460405163c181b27360e01b815260048101919091527f00000000000000000000000000000000000000000000000000000000000000006001600160a01b03169063c181b27390602401602060405180830381865afa1580156106cf573d5f5f3e3d5ffd5b505050506040513d601f19601f820116820180604052508101906106f39190610b32565b610716578054604051636bed6a6760e01b815260048101919091526024016104c2565b5f8581526001602090815260408083208784529091528120546002830154909190610742908390610b61565b90508085111561077d576040516310c9fba760e11b8152600481018890526024810187905260448101869052606481018290526084016104c2565b6107878583610b7a565b5f8881526001602090815260408083208a845290915290205550505490509392505050565b6107d560405180608001604052805f81526020015f81526020015f81526020015f151581525090565b6107de826109ed565b60408051608081018252825481526001830154602082015260028301549181019190915260039091015460ff161515606082015292915050565b7f00000000000000000000000000000000000000000000000000000000000000006001600160a01b03166391d148547f00000000000000000000000000000000000000000000000000000000000000006001600160a01b031663a217fddf6040518163ffffffff1660e01b8152600401602060405180830381865afa1580156108a3573d5f5f3e3d5ffd5b505050506040513d601f19601f820116820180604052508101906108c79190610b1b565b6040516001600160e01b031960e084901b1681526004810191909152336024820152604401602060405180830381865afa158015610907573d5f5f3e3d5ffd5b505050506040513d601f19601f8201168201806040525081019061092b9190610b32565b61094857604051637bfa4b9f60e01b815260040160405180910390fd5b6002546001600160a01b03161561097257604051630b1300c360e11b815260040160405180910390fd5b6001600160a01b03811661099957604051637c946ed760e01b815260040160405180910390fd5b600280546001600160a01b0319166001600160a01b0383169081179091556040519081527f6999c66b444fbfcba8ecb9e4c505c90a037636fc69b689682064d60e6b574d269060200160405180910390a150565b5f8181526020819052604090208054610a1c57604051631dd9db2d60e31b8152600481018390526024016104c2565b919050565b8015158114610a2e575f5ffd5b50565b5f5f60408385031215610a42575f5ffd5b823591506020830135610a5481610a21565b809150509250929050565b5f5f5f5f60808587031215610a72575f5ffd5b5050823594602084013594506040840135936060013592509050565b5f5f5f60608486031215610aa0575f5ffd5b505081359360208301359350604090920135919050565b5f60208284031215610ac7575f5ffd5b5035919050565b5f5f60408385031215610adf575f5ffd5b50508035926020909101359150565b5f60208284031215610afe575f5ffd5b81356001600160a01b0381168114610b14575f5ffd5b9392505050565b5f60208284031215610b2b575f5ffd5b5051919050565b5f60208284031215610b42575f5ffd5b8151610b1481610a21565b634e487b7160e01b5f52601160045260245ffd5b81810381811115610b7457610b74610b4d565b92915050565b80820180821115610b7457610b74610b4d56fea264697066735822122021b2ee597beb279a767f154cab303d6262eb87e176b9db40f4c9a0abeef42e9a64736f6c63430008220033",
}

// FacilityRegistryABI is the input ABI used to generate the binding from.
// Deprecated: Use FacilityRegistryMetaData.ABI instead.
var FacilityRegistryABI = FacilityRegistryMetaData.ABI

// FacilityRegistryBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use FacilityRegistryMetaData.Bin instead.
var FacilityRegistryBin = FacilityRegistryMetaData.Bin

// DeployFacilityRegistry deploys a new Ethereum contract, binding an instance of FacilityRegistry to it.
func DeployFacilityRegistry(auth *bind.TransactOpts, backend bind.ContractBackend, orgs_ common.Address) (common.Address, *types.Transaction, *FacilityRegistry, error) {
	parsed, err := FacilityRegistryMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(FacilityRegistryBin), backend, orgs_)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &FacilityRegistry{FacilityRegistryCaller: FacilityRegistryCaller{contract: contract}, FacilityRegistryTransactor: FacilityRegistryTransactor{contract: contract}, FacilityRegistryFilterer: FacilityRegistryFilterer{contract: contract}}, nil
}

// FacilityRegistry is an auto generated Go binding around an Ethereum contract.
type FacilityRegistry struct {
	FacilityRegistryCaller     // Read-only binding to the contract
	FacilityRegistryTransactor // Write-only binding to the contract
	FacilityRegistryFilterer   // Log filterer for contract events
}

// FacilityRegistryCaller is an auto generated read-only Go binding around an Ethereum contract.
type FacilityRegistryCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// FacilityRegistryTransactor is an auto generated write-only Go binding around an Ethereum contract.
type FacilityRegistryTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// FacilityRegistryFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type FacilityRegistryFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// FacilityRegistrySession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type FacilityRegistrySession struct {
	Contract     *FacilityRegistry // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// FacilityRegistryCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type FacilityRegistryCallerSession struct {
	Contract *FacilityRegistryCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts           // Call options to use throughout this session
}

// FacilityRegistryTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type FacilityRegistryTransactorSession struct {
	Contract     *FacilityRegistryTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts           // Transaction auth options to use throughout this session
}

// FacilityRegistryRaw is an auto generated low-level Go binding around an Ethereum contract.
type FacilityRegistryRaw struct {
	Contract *FacilityRegistry // Generic contract binding to access the raw methods on
}

// FacilityRegistryCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type FacilityRegistryCallerRaw struct {
	Contract *FacilityRegistryCaller // Generic read-only contract binding to access the raw methods on
}

// FacilityRegistryTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type FacilityRegistryTransactorRaw struct {
	Contract *FacilityRegistryTransactor // Generic write-only contract binding to access the raw methods on
}

// NewFacilityRegistry creates a new instance of FacilityRegistry, bound to a specific deployed contract.
func NewFacilityRegistry(address common.Address, backend bind.ContractBackend) (*FacilityRegistry, error) {
	contract, err := bindFacilityRegistry(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &FacilityRegistry{FacilityRegistryCaller: FacilityRegistryCaller{contract: contract}, FacilityRegistryTransactor: FacilityRegistryTransactor{contract: contract}, FacilityRegistryFilterer: FacilityRegistryFilterer{contract: contract}}, nil
}

// NewFacilityRegistryCaller creates a new read-only instance of FacilityRegistry, bound to a specific deployed contract.
func NewFacilityRegistryCaller(address common.Address, caller bind.ContractCaller) (*FacilityRegistryCaller, error) {
	contract, err := bindFacilityRegistry(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &FacilityRegistryCaller{contract: contract}, nil
}

// NewFacilityRegistryTransactor creates a new write-only instance of FacilityRegistry, bound to a specific deployed contract.
func NewFacilityRegistryTransactor(address common.Address, transactor bind.ContractTransactor) (*FacilityRegistryTransactor, error) {
	contract, err := bindFacilityRegistry(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &FacilityRegistryTransactor{contract: contract}, nil
}

// NewFacilityRegistryFilterer creates a new log filterer instance of FacilityRegistry, bound to a specific deployed contract.
func NewFacilityRegistryFilterer(address common.Address, filterer bind.ContractFilterer) (*FacilityRegistryFilterer, error) {
	contract, err := bindFacilityRegistry(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &FacilityRegistryFilterer{contract: contract}, nil
}

// bindFacilityRegistry binds a generic wrapper to an already deployed contract.
func bindFacilityRegistry(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := FacilityRegistryMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_FacilityRegistry *FacilityRegistryRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _FacilityRegistry.Contract.FacilityRegistryCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_FacilityRegistry *FacilityRegistryRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _FacilityRegistry.Contract.FacilityRegistryTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_FacilityRegistry *FacilityRegistryRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _FacilityRegistry.Contract.FacilityRegistryTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_FacilityRegistry *FacilityRegistryCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _FacilityRegistry.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_FacilityRegistry *FacilityRegistryTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _FacilityRegistry.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_FacilityRegistry *FacilityRegistryTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _FacilityRegistry.Contract.contract.Transact(opts, method, params...)
}

// Facility is a free data retrieval call binding the contract method 0x6e6b15b1.
//
// Solidity: function facility(bytes32 facilityId) view returns((bytes32,bytes32,uint256,bool))
func (_FacilityRegistry *FacilityRegistryCaller) Facility(opts *bind.CallOpts, facilityId [32]byte) (FacilityRegistryFacility, error) {
	var out []interface{}
	err := _FacilityRegistry.contract.Call(opts, &out, "facility", facilityId)

	if err != nil {
		return *new(FacilityRegistryFacility), err
	}

	out0 := *abi.ConvertType(out[0], new(FacilityRegistryFacility)).(*FacilityRegistryFacility)

	return out0, err

}

// Facility is a free data retrieval call binding the contract method 0x6e6b15b1.
//
// Solidity: function facility(bytes32 facilityId) view returns((bytes32,bytes32,uint256,bool))
func (_FacilityRegistry *FacilityRegistrySession) Facility(facilityId [32]byte) (FacilityRegistryFacility, error) {
	return _FacilityRegistry.Contract.Facility(&_FacilityRegistry.CallOpts, facilityId)
}

// Facility is a free data retrieval call binding the contract method 0x6e6b15b1.
//
// Solidity: function facility(bytes32 facilityId) view returns((bytes32,bytes32,uint256,bool))
func (_FacilityRegistry *FacilityRegistryCallerSession) Facility(facilityId [32]byte) (FacilityRegistryFacility, error) {
	return _FacilityRegistry.Contract.Facility(&_FacilityRegistry.CallOpts, facilityId)
}

// Orgs is a free data retrieval call binding the contract method 0xcab19cdc.
//
// Solidity: function orgs() view returns(address)
func (_FacilityRegistry *FacilityRegistryCaller) Orgs(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _FacilityRegistry.contract.Call(opts, &out, "orgs")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Orgs is a free data retrieval call binding the contract method 0xcab19cdc.
//
// Solidity: function orgs() view returns(address)
func (_FacilityRegistry *FacilityRegistrySession) Orgs() (common.Address, error) {
	return _FacilityRegistry.Contract.Orgs(&_FacilityRegistry.CallOpts)
}

// Orgs is a free data retrieval call binding the contract method 0xcab19cdc.
//
// Solidity: function orgs() view returns(address)
func (_FacilityRegistry *FacilityRegistryCallerSession) Orgs() (common.Address, error) {
	return _FacilityRegistry.Contract.Orgs(&_FacilityRegistry.CallOpts)
}

// OriginatedInPeriod is a free data retrieval call binding the contract method 0xc9b4cc6c.
//
// Solidity: function originatedInPeriod(bytes32 facilityId, uint256 period) view returns(uint256)
func (_FacilityRegistry *FacilityRegistryCaller) OriginatedInPeriod(opts *bind.CallOpts, facilityId [32]byte, period *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _FacilityRegistry.contract.Call(opts, &out, "originatedInPeriod", facilityId, period)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// OriginatedInPeriod is a free data retrieval call binding the contract method 0xc9b4cc6c.
//
// Solidity: function originatedInPeriod(bytes32 facilityId, uint256 period) view returns(uint256)
func (_FacilityRegistry *FacilityRegistrySession) OriginatedInPeriod(facilityId [32]byte, period *big.Int) (*big.Int, error) {
	return _FacilityRegistry.Contract.OriginatedInPeriod(&_FacilityRegistry.CallOpts, facilityId, period)
}

// OriginatedInPeriod is a free data retrieval call binding the contract method 0xc9b4cc6c.
//
// Solidity: function originatedInPeriod(bytes32 facilityId, uint256 period) view returns(uint256)
func (_FacilityRegistry *FacilityRegistryCallerSession) OriginatedInPeriod(facilityId [32]byte, period *big.Int) (*big.Int, error) {
	return _FacilityRegistry.Contract.OriginatedInPeriod(&_FacilityRegistry.CallOpts, facilityId, period)
}

// Originator is a free data retrieval call binding the contract method 0x3c815d9e.
//
// Solidity: function originator() view returns(address)
func (_FacilityRegistry *FacilityRegistryCaller) Originator(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _FacilityRegistry.contract.Call(opts, &out, "originator")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Originator is a free data retrieval call binding the contract method 0x3c815d9e.
//
// Solidity: function originator() view returns(address)
func (_FacilityRegistry *FacilityRegistrySession) Originator() (common.Address, error) {
	return _FacilityRegistry.Contract.Originator(&_FacilityRegistry.CallOpts)
}

// Originator is a free data retrieval call binding the contract method 0x3c815d9e.
//
// Solidity: function originator() view returns(address)
func (_FacilityRegistry *FacilityRegistryCallerSession) Originator() (common.Address, error) {
	return _FacilityRegistry.Contract.Originator(&_FacilityRegistry.CallOpts)
}

// ConsumeCapacity is a paid mutator transaction binding the contract method 0x63cba5ca.
//
// Solidity: function consumeCapacity(bytes32 facilityId, uint256 period, uint256 quantity) returns(bytes32 orgId)
func (_FacilityRegistry *FacilityRegistryTransactor) ConsumeCapacity(opts *bind.TransactOpts, facilityId [32]byte, period *big.Int, quantity *big.Int) (*types.Transaction, error) {
	return _FacilityRegistry.contract.Transact(opts, "consumeCapacity", facilityId, period, quantity)
}

// ConsumeCapacity is a paid mutator transaction binding the contract method 0x63cba5ca.
//
// Solidity: function consumeCapacity(bytes32 facilityId, uint256 period, uint256 quantity) returns(bytes32 orgId)
func (_FacilityRegistry *FacilityRegistrySession) ConsumeCapacity(facilityId [32]byte, period *big.Int, quantity *big.Int) (*types.Transaction, error) {
	return _FacilityRegistry.Contract.ConsumeCapacity(&_FacilityRegistry.TransactOpts, facilityId, period, quantity)
}

// ConsumeCapacity is a paid mutator transaction binding the contract method 0x63cba5ca.
//
// Solidity: function consumeCapacity(bytes32 facilityId, uint256 period, uint256 quantity) returns(bytes32 orgId)
func (_FacilityRegistry *FacilityRegistryTransactorSession) ConsumeCapacity(facilityId [32]byte, period *big.Int, quantity *big.Int) (*types.Transaction, error) {
	return _FacilityRegistry.Contract.ConsumeCapacity(&_FacilityRegistry.TransactOpts, facilityId, period, quantity)
}

// Register is a paid mutator transaction binding the contract method 0x42ee3d0a.
//
// Solidity: function register(bytes32 facilityId, bytes32 orgId, bytes32 dataHash, uint256 capacityPerPeriod) returns()
func (_FacilityRegistry *FacilityRegistryTransactor) Register(opts *bind.TransactOpts, facilityId [32]byte, orgId [32]byte, dataHash [32]byte, capacityPerPeriod *big.Int) (*types.Transaction, error) {
	return _FacilityRegistry.contract.Transact(opts, "register", facilityId, orgId, dataHash, capacityPerPeriod)
}

// Register is a paid mutator transaction binding the contract method 0x42ee3d0a.
//
// Solidity: function register(bytes32 facilityId, bytes32 orgId, bytes32 dataHash, uint256 capacityPerPeriod) returns()
func (_FacilityRegistry *FacilityRegistrySession) Register(facilityId [32]byte, orgId [32]byte, dataHash [32]byte, capacityPerPeriod *big.Int) (*types.Transaction, error) {
	return _FacilityRegistry.Contract.Register(&_FacilityRegistry.TransactOpts, facilityId, orgId, dataHash, capacityPerPeriod)
}

// Register is a paid mutator transaction binding the contract method 0x42ee3d0a.
//
// Solidity: function register(bytes32 facilityId, bytes32 orgId, bytes32 dataHash, uint256 capacityPerPeriod) returns()
func (_FacilityRegistry *FacilityRegistryTransactorSession) Register(facilityId [32]byte, orgId [32]byte, dataHash [32]byte, capacityPerPeriod *big.Int) (*types.Transaction, error) {
	return _FacilityRegistry.Contract.Register(&_FacilityRegistry.TransactOpts, facilityId, orgId, dataHash, capacityPerPeriod)
}

// SetOriginator is a paid mutator transaction binding the contract method 0xe7715b05.
//
// Solidity: function setOriginator(address originator_) returns()
func (_FacilityRegistry *FacilityRegistryTransactor) SetOriginator(opts *bind.TransactOpts, originator_ common.Address) (*types.Transaction, error) {
	return _FacilityRegistry.contract.Transact(opts, "setOriginator", originator_)
}

// SetOriginator is a paid mutator transaction binding the contract method 0xe7715b05.
//
// Solidity: function setOriginator(address originator_) returns()
func (_FacilityRegistry *FacilityRegistrySession) SetOriginator(originator_ common.Address) (*types.Transaction, error) {
	return _FacilityRegistry.Contract.SetOriginator(&_FacilityRegistry.TransactOpts, originator_)
}

// SetOriginator is a paid mutator transaction binding the contract method 0xe7715b05.
//
// Solidity: function setOriginator(address originator_) returns()
func (_FacilityRegistry *FacilityRegistryTransactorSession) SetOriginator(originator_ common.Address) (*types.Transaction, error) {
	return _FacilityRegistry.Contract.SetOriginator(&_FacilityRegistry.TransactOpts, originator_)
}

// SetSuspended is a paid mutator transaction binding the contract method 0x29f72845.
//
// Solidity: function setSuspended(bytes32 facilityId, bool suspended) returns()
func (_FacilityRegistry *FacilityRegistryTransactor) SetSuspended(opts *bind.TransactOpts, facilityId [32]byte, suspended bool) (*types.Transaction, error) {
	return _FacilityRegistry.contract.Transact(opts, "setSuspended", facilityId, suspended)
}

// SetSuspended is a paid mutator transaction binding the contract method 0x29f72845.
//
// Solidity: function setSuspended(bytes32 facilityId, bool suspended) returns()
func (_FacilityRegistry *FacilityRegistrySession) SetSuspended(facilityId [32]byte, suspended bool) (*types.Transaction, error) {
	return _FacilityRegistry.Contract.SetSuspended(&_FacilityRegistry.TransactOpts, facilityId, suspended)
}

// SetSuspended is a paid mutator transaction binding the contract method 0x29f72845.
//
// Solidity: function setSuspended(bytes32 facilityId, bool suspended) returns()
func (_FacilityRegistry *FacilityRegistryTransactorSession) SetSuspended(facilityId [32]byte, suspended bool) (*types.Transaction, error) {
	return _FacilityRegistry.Contract.SetSuspended(&_FacilityRegistry.TransactOpts, facilityId, suspended)
}

// FacilityRegistryFacilityRegisteredIterator is returned from FilterFacilityRegistered and is used to iterate over the raw logs and unpacked data for FacilityRegistered events raised by the FacilityRegistry contract.
type FacilityRegistryFacilityRegisteredIterator struct {
	Event *FacilityRegistryFacilityRegistered // Event containing the contract specifics and raw log

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
func (it *FacilityRegistryFacilityRegisteredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(FacilityRegistryFacilityRegistered)
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
		it.Event = new(FacilityRegistryFacilityRegistered)
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
func (it *FacilityRegistryFacilityRegisteredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *FacilityRegistryFacilityRegisteredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// FacilityRegistryFacilityRegistered represents a FacilityRegistered event raised by the FacilityRegistry contract.
type FacilityRegistryFacilityRegistered struct {
	FacilityId        [32]byte
	OrgId             [32]byte
	DataHash          [32]byte
	CapacityPerPeriod *big.Int
	Raw               types.Log // Blockchain specific contextual infos
}

// FilterFacilityRegistered is a free log retrieval operation binding the contract event 0xa7e7e5534486d104bc4e154993522de6fedc9e38050288ce452c5ca1235443a3.
//
// Solidity: event FacilityRegistered(bytes32 indexed facilityId, bytes32 indexed orgId, bytes32 dataHash, uint256 capacityPerPeriod)
func (_FacilityRegistry *FacilityRegistryFilterer) FilterFacilityRegistered(opts *bind.FilterOpts, facilityId [][32]byte, orgId [][32]byte) (*FacilityRegistryFacilityRegisteredIterator, error) {

	var facilityIdRule []interface{}
	for _, facilityIdItem := range facilityId {
		facilityIdRule = append(facilityIdRule, facilityIdItem)
	}
	var orgIdRule []interface{}
	for _, orgIdItem := range orgId {
		orgIdRule = append(orgIdRule, orgIdItem)
	}

	logs, sub, err := _FacilityRegistry.contract.FilterLogs(opts, "FacilityRegistered", facilityIdRule, orgIdRule)
	if err != nil {
		return nil, err
	}
	return &FacilityRegistryFacilityRegisteredIterator{contract: _FacilityRegistry.contract, event: "FacilityRegistered", logs: logs, sub: sub}, nil
}

// WatchFacilityRegistered is a free log subscription operation binding the contract event 0xa7e7e5534486d104bc4e154993522de6fedc9e38050288ce452c5ca1235443a3.
//
// Solidity: event FacilityRegistered(bytes32 indexed facilityId, bytes32 indexed orgId, bytes32 dataHash, uint256 capacityPerPeriod)
func (_FacilityRegistry *FacilityRegistryFilterer) WatchFacilityRegistered(opts *bind.WatchOpts, sink chan<- *FacilityRegistryFacilityRegistered, facilityId [][32]byte, orgId [][32]byte) (event.Subscription, error) {

	var facilityIdRule []interface{}
	for _, facilityIdItem := range facilityId {
		facilityIdRule = append(facilityIdRule, facilityIdItem)
	}
	var orgIdRule []interface{}
	for _, orgIdItem := range orgId {
		orgIdRule = append(orgIdRule, orgIdItem)
	}

	logs, sub, err := _FacilityRegistry.contract.WatchLogs(opts, "FacilityRegistered", facilityIdRule, orgIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(FacilityRegistryFacilityRegistered)
				if err := _FacilityRegistry.contract.UnpackLog(event, "FacilityRegistered", log); err != nil {
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

// ParseFacilityRegistered is a log parse operation binding the contract event 0xa7e7e5534486d104bc4e154993522de6fedc9e38050288ce452c5ca1235443a3.
//
// Solidity: event FacilityRegistered(bytes32 indexed facilityId, bytes32 indexed orgId, bytes32 dataHash, uint256 capacityPerPeriod)
func (_FacilityRegistry *FacilityRegistryFilterer) ParseFacilityRegistered(log types.Log) (*FacilityRegistryFacilityRegistered, error) {
	event := new(FacilityRegistryFacilityRegistered)
	if err := _FacilityRegistry.contract.UnpackLog(event, "FacilityRegistered", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// FacilityRegistryFacilitySuspendedIterator is returned from FilterFacilitySuspended and is used to iterate over the raw logs and unpacked data for FacilitySuspended events raised by the FacilityRegistry contract.
type FacilityRegistryFacilitySuspendedIterator struct {
	Event *FacilityRegistryFacilitySuspended // Event containing the contract specifics and raw log

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
func (it *FacilityRegistryFacilitySuspendedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(FacilityRegistryFacilitySuspended)
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
		it.Event = new(FacilityRegistryFacilitySuspended)
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
func (it *FacilityRegistryFacilitySuspendedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *FacilityRegistryFacilitySuspendedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// FacilityRegistryFacilitySuspended represents a FacilitySuspended event raised by the FacilityRegistry contract.
type FacilityRegistryFacilitySuspended struct {
	FacilityId [32]byte
	Suspended  bool
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterFacilitySuspended is a free log retrieval operation binding the contract event 0x7a67dd3e17e1b246cbb7ea1c7a3e96d62195b88cbbea41490e2b06bd59d03469.
//
// Solidity: event FacilitySuspended(bytes32 indexed facilityId, bool suspended)
func (_FacilityRegistry *FacilityRegistryFilterer) FilterFacilitySuspended(opts *bind.FilterOpts, facilityId [][32]byte) (*FacilityRegistryFacilitySuspendedIterator, error) {

	var facilityIdRule []interface{}
	for _, facilityIdItem := range facilityId {
		facilityIdRule = append(facilityIdRule, facilityIdItem)
	}

	logs, sub, err := _FacilityRegistry.contract.FilterLogs(opts, "FacilitySuspended", facilityIdRule)
	if err != nil {
		return nil, err
	}
	return &FacilityRegistryFacilitySuspendedIterator{contract: _FacilityRegistry.contract, event: "FacilitySuspended", logs: logs, sub: sub}, nil
}

// WatchFacilitySuspended is a free log subscription operation binding the contract event 0x7a67dd3e17e1b246cbb7ea1c7a3e96d62195b88cbbea41490e2b06bd59d03469.
//
// Solidity: event FacilitySuspended(bytes32 indexed facilityId, bool suspended)
func (_FacilityRegistry *FacilityRegistryFilterer) WatchFacilitySuspended(opts *bind.WatchOpts, sink chan<- *FacilityRegistryFacilitySuspended, facilityId [][32]byte) (event.Subscription, error) {

	var facilityIdRule []interface{}
	for _, facilityIdItem := range facilityId {
		facilityIdRule = append(facilityIdRule, facilityIdItem)
	}

	logs, sub, err := _FacilityRegistry.contract.WatchLogs(opts, "FacilitySuspended", facilityIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(FacilityRegistryFacilitySuspended)
				if err := _FacilityRegistry.contract.UnpackLog(event, "FacilitySuspended", log); err != nil {
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

// ParseFacilitySuspended is a log parse operation binding the contract event 0x7a67dd3e17e1b246cbb7ea1c7a3e96d62195b88cbbea41490e2b06bd59d03469.
//
// Solidity: event FacilitySuspended(bytes32 indexed facilityId, bool suspended)
func (_FacilityRegistry *FacilityRegistryFilterer) ParseFacilitySuspended(log types.Log) (*FacilityRegistryFacilitySuspended, error) {
	event := new(FacilityRegistryFacilitySuspended)
	if err := _FacilityRegistry.contract.UnpackLog(event, "FacilitySuspended", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// FacilityRegistryOriginatorSetIterator is returned from FilterOriginatorSet and is used to iterate over the raw logs and unpacked data for OriginatorSet events raised by the FacilityRegistry contract.
type FacilityRegistryOriginatorSetIterator struct {
	Event *FacilityRegistryOriginatorSet // Event containing the contract specifics and raw log

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
func (it *FacilityRegistryOriginatorSetIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(FacilityRegistryOriginatorSet)
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
		it.Event = new(FacilityRegistryOriginatorSet)
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
func (it *FacilityRegistryOriginatorSetIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *FacilityRegistryOriginatorSetIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// FacilityRegistryOriginatorSet represents a OriginatorSet event raised by the FacilityRegistry contract.
type FacilityRegistryOriginatorSet struct {
	Originator common.Address
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterOriginatorSet is a free log retrieval operation binding the contract event 0x6999c66b444fbfcba8ecb9e4c505c90a037636fc69b689682064d60e6b574d26.
//
// Solidity: event OriginatorSet(address originator)
func (_FacilityRegistry *FacilityRegistryFilterer) FilterOriginatorSet(opts *bind.FilterOpts) (*FacilityRegistryOriginatorSetIterator, error) {

	logs, sub, err := _FacilityRegistry.contract.FilterLogs(opts, "OriginatorSet")
	if err != nil {
		return nil, err
	}
	return &FacilityRegistryOriginatorSetIterator{contract: _FacilityRegistry.contract, event: "OriginatorSet", logs: logs, sub: sub}, nil
}

// WatchOriginatorSet is a free log subscription operation binding the contract event 0x6999c66b444fbfcba8ecb9e4c505c90a037636fc69b689682064d60e6b574d26.
//
// Solidity: event OriginatorSet(address originator)
func (_FacilityRegistry *FacilityRegistryFilterer) WatchOriginatorSet(opts *bind.WatchOpts, sink chan<- *FacilityRegistryOriginatorSet) (event.Subscription, error) {

	logs, sub, err := _FacilityRegistry.contract.WatchLogs(opts, "OriginatorSet")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(FacilityRegistryOriginatorSet)
				if err := _FacilityRegistry.contract.UnpackLog(event, "OriginatorSet", log); err != nil {
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

// ParseOriginatorSet is a log parse operation binding the contract event 0x6999c66b444fbfcba8ecb9e4c505c90a037636fc69b689682064d60e6b574d26.
//
// Solidity: event OriginatorSet(address originator)
func (_FacilityRegistry *FacilityRegistryFilterer) ParseOriginatorSet(log types.Log) (*FacilityRegistryOriginatorSet, error) {
	event := new(FacilityRegistryOriginatorSet)
	if err := _FacilityRegistry.contract.UnpackLog(event, "OriginatorSet", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
