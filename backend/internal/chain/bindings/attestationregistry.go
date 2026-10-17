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
	CertType  [32]byte
	DocHash   [32]byte
	ValidFrom uint64
	ValidTo   uint64
	Revoked   bool
}

// AttestationRegistryMetaData contains all meta data concerning the AttestationRegistry contract.
var AttestationRegistryMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"contractOrgRegistry\",\"name\":\"orgs_\",\"type\":\"address\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"attestationId\",\"type\":\"bytes32\"}],\"name\":\"AlreadyRevoked\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"attestationId\",\"type\":\"bytes32\"}],\"name\":\"AttestationExists\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"attestationId\",\"type\":\"bytes32\"}],\"name\":\"AttestationUnknown\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint64\",\"name\":\"validFrom\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"validTo\",\"type\":\"uint64\"}],\"name\":\"BadValidity\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"orgId\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"certType\",\"type\":\"bytes32\"}],\"name\":\"NotAccredited\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotAdmin\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotOperator\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotRelayer\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"ZeroValue\",\"type\":\"error\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"attestationId\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"issuerOrg\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"subject\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"certType\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"docHash\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"validFrom\",\"type\":\"uint64\"},{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"validTo\",\"type\":\"uint64\"}],\"name\":\"AttestationIssued\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"attestationId\",\"type\":\"bytes32\"}],\"name\":\"Revoked\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"attestationId\",\"type\":\"bytes32\"}],\"name\":\"attestation\",\"outputs\":[{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"issuerOrg\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"subject\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"certType\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"docHash\",\"type\":\"bytes32\"},{\"internalType\":\"uint64\",\"name\":\"validFrom\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"validTo\",\"type\":\"uint64\"},{\"internalType\":\"bool\",\"name\":\"revoked\",\"type\":\"bool\"}],\"internalType\":\"structAttestationRegistry.Attestation\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"attestationId\",\"type\":\"bytes32\"}],\"name\":\"isValid\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"attestationId\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"issuerOrg\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"subject\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"certType\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"docHash\",\"type\":\"bytes32\"},{\"internalType\":\"uint64\",\"name\":\"validFrom\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"validTo\",\"type\":\"uint64\"}],\"name\":\"issue\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"orgs\",\"outputs\":[{\"internalType\":\"contractOrgRegistry\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"attestationId\",\"type\":\"bytes32\"}],\"name\":\"revoke\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
	Bin: "0x60a060405234801561000f575f5ffd5b50604051610a0e380380610a0e83398101604081905261002e9161003f565b6001600160a01b031660805261006c565b5f6020828403121561004f575f5ffd5b81516001600160a01b0381168114610065575f5ffd5b9392505050565b6080516109676100a75f395f818161012e0152818161022801528181610257015281816103e20152818161068e01526106bd01526109675ff3fe608060405234801561000f575f5ffd5b5060043610610055575f3560e01c80631a7fca861461005957806349e1904c146100de5780636a938567146100f3578063b75c7dc614610116578063cab19cdc14610129575b5f5ffd5b61006c610067366004610870565b610168565b6040516100d591905f60e082019050825182526020830151602083015260408301516040830152606083015160608301526001600160401b0360808401511660808301526001600160401b0360a08401511660a083015260c0830151151560c083015292915050565b60405180910390f35b6100f16100ec36600461089d565b610226565b005b610106610101366004610870565b610623565b60405190151581526020016100d5565b6100f1610124366004610870565b61068c565b6101507f000000000000000000000000000000000000000000000000000000000000000081565b6040516001600160a01b0390911681526020016100d5565b6101b56040518060e001604052805f81526020015f81526020015f81526020015f81526020015f6001600160401b031681526020015f6001600160401b031681526020015f151581525090565b6101be8261083c565b6040805160e0810182528254815260018301546020820152600283015491810191909152600382015460608201526004909101546001600160401b038082166080840152600160401b82041660a0830152600160801b900460ff16151560c082015292915050565b7f00000000000000000000000000000000000000000000000000000000000000006001600160a01b03166391d148547f00000000000000000000000000000000000000000000000000000000000000006001600160a01b031663926d7d7f6040518163ffffffff1660e01b8152600401602060405180830381865afa1580156102b1573d5f5f3e3d5ffd5b505050506040513d601f19601f820116820180604052508101906102d591906108fb565b6040516001600160e01b031960e084901b1681526004810191909152336024820152604401602060405180830381865afa158015610315573d5f5f3e3d5ffd5b505050506040513d601f19601f820116820180604052508101906103399190610912565b6103565760405163c64891a560e01b815260040160405180910390fd5b861580610361575084155b8061036a575083155b80610373575082155b1561039157604051637c946ed760e01b815260040160405180910390fd5b5f87815260208190526040902054156103c557604051631722b9b560e31b8152600481018890526024015b60405180910390fd5b60405163056806ef60e51b815260048101879052602481018590527f00000000000000000000000000000000000000000000000000000000000000006001600160a01b03169063ad00dde090604401602060405180830381865afa15801561042f573d5f5f3e3d5ffd5b505050506040513d601f19601f820116820180604052508101906104539190610912565b61047a576040516383cddc1960e01b815260048101879052602481018590526044016103bc565b816001600160401b0316816001600160401b0316116104bf576040516355da959560e11b81526001600160401b038084166004830152821660248201526044016103bc565b6040518060e00160405280878152602001868152602001858152602001848152602001836001600160401b03168152602001826001600160401b031681526020015f15158152505f5f8981526020019081526020015f205f820151815f01556020820151816001015560408201518160020155606082015181600301556080820151816004015f6101000a8154816001600160401b0302191690836001600160401b0316021790555060a08201518160040160086101000a8154816001600160401b0302191690836001600160401b0316021790555060c08201518160040160106101000a81548160ff0219169083151502179055509050508486887ff25c7c8778fe6025840ac69375908fd6e7cd7ce7b90cd70535f09bf734372bc087878787604051610612949392919093845260208401929092526001600160401b03908116604084015216606082015260800190565b60405180910390a450505050505050565b5f81815260208190526040812080541580159061064c57506004810154600160801b900460ff16155b8015610665575060048101546001600160401b03164210155b801561068557506004810154600160401b90046001600160401b03164211155b9392505050565b7f00000000000000000000000000000000000000000000000000000000000000006001600160a01b03166391d148547f00000000000000000000000000000000000000000000000000000000000000006001600160a01b031663926d7d7f6040518163ffffffff1660e01b8152600401602060405180830381865afa158015610717573d5f5f3e3d5ffd5b505050506040513d601f19601f8201168201806040525081019061073b91906108fb565b6040516001600160e01b031960e084901b1681526004810191909152336024820152604401602060405180830381865afa15801561077b573d5f5f3e3d5ffd5b505050506040513d601f19601f8201168201806040525081019061079f9190610912565b6107bc5760405163c64891a560e01b815260040160405180910390fd5b5f6107c68261083c565b6004810154909150600160801b900460ff16156107f9576040516390315de160e01b8152600481018390526024016103bc565b60048101805460ff60801b1916600160801b17905560405182907fe5af7daed5ab2a2dc5f98d53619f05089c0c14d11a6621f6b906a2366c9a7ab3905f90a25050565b5f818152602081905260409020805461086b5760405163b515dbdd60e01b8152600481018390526024016103bc565b919050565b5f60208284031215610880575f5ffd5b5035919050565b80356001600160401b038116811461086b575f5ffd5b5f5f5f5f5f5f5f60e0888a0312156108b3575f5ffd5b87359650602088013595506040880135945060608801359350608088013592506108df60a08901610887565b91506108ed60c08901610887565b905092959891949750929550565b5f6020828403121561090b575f5ffd5b5051919050565b5f60208284031215610922575f5ffd5b81518015158114610685575f5ffdfea26469706673582212207c20a298b66e1212e14eaa971fd192ce9ed032edbb1813c53124364a26c5c7d664736f6c63430008220033",
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
// Solidity: function attestation(bytes32 attestationId) view returns((bytes32,bytes32,bytes32,bytes32,uint64,uint64,bool))
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
// Solidity: function attestation(bytes32 attestationId) view returns((bytes32,bytes32,bytes32,bytes32,uint64,uint64,bool))
func (_AttestationRegistry *AttestationRegistrySession) Attestation(attestationId [32]byte) (AttestationRegistryAttestation, error) {
	return _AttestationRegistry.Contract.Attestation(&_AttestationRegistry.CallOpts, attestationId)
}

// Attestation is a free data retrieval call binding the contract method 0x1a7fca86.
//
// Solidity: function attestation(bytes32 attestationId) view returns((bytes32,bytes32,bytes32,bytes32,uint64,uint64,bool))
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

// Issue is a paid mutator transaction binding the contract method 0x49e1904c.
//
// Solidity: function issue(bytes32 attestationId, bytes32 issuerOrg, bytes32 subject, bytes32 certType, bytes32 docHash, uint64 validFrom, uint64 validTo) returns()
func (_AttestationRegistry *AttestationRegistryTransactor) Issue(opts *bind.TransactOpts, attestationId [32]byte, issuerOrg [32]byte, subject [32]byte, certType [32]byte, docHash [32]byte, validFrom uint64, validTo uint64) (*types.Transaction, error) {
	return _AttestationRegistry.contract.Transact(opts, "issue", attestationId, issuerOrg, subject, certType, docHash, validFrom, validTo)
}

// Issue is a paid mutator transaction binding the contract method 0x49e1904c.
//
// Solidity: function issue(bytes32 attestationId, bytes32 issuerOrg, bytes32 subject, bytes32 certType, bytes32 docHash, uint64 validFrom, uint64 validTo) returns()
func (_AttestationRegistry *AttestationRegistrySession) Issue(attestationId [32]byte, issuerOrg [32]byte, subject [32]byte, certType [32]byte, docHash [32]byte, validFrom uint64, validTo uint64) (*types.Transaction, error) {
	return _AttestationRegistry.Contract.Issue(&_AttestationRegistry.TransactOpts, attestationId, issuerOrg, subject, certType, docHash, validFrom, validTo)
}

// Issue is a paid mutator transaction binding the contract method 0x49e1904c.
//
// Solidity: function issue(bytes32 attestationId, bytes32 issuerOrg, bytes32 subject, bytes32 certType, bytes32 docHash, uint64 validFrom, uint64 validTo) returns()
func (_AttestationRegistry *AttestationRegistryTransactorSession) Issue(attestationId [32]byte, issuerOrg [32]byte, subject [32]byte, certType [32]byte, docHash [32]byte, validFrom uint64, validTo uint64) (*types.Transaction, error) {
	return _AttestationRegistry.Contract.Issue(&_AttestationRegistry.TransactOpts, attestationId, issuerOrg, subject, certType, docHash, validFrom, validTo)
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
	CertType      [32]byte
	DocHash       [32]byte
	ValidFrom     uint64
	ValidTo       uint64
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterAttestationIssued is a free log retrieval operation binding the contract event 0xf25c7c8778fe6025840ac69375908fd6e7cd7ce7b90cd70535f09bf734372bc0.
//
// Solidity: event AttestationIssued(bytes32 indexed attestationId, bytes32 indexed issuerOrg, bytes32 indexed subject, bytes32 certType, bytes32 docHash, uint64 validFrom, uint64 validTo)
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

// WatchAttestationIssued is a free log subscription operation binding the contract event 0xf25c7c8778fe6025840ac69375908fd6e7cd7ce7b90cd70535f09bf734372bc0.
//
// Solidity: event AttestationIssued(bytes32 indexed attestationId, bytes32 indexed issuerOrg, bytes32 indexed subject, bytes32 certType, bytes32 docHash, uint64 validFrom, uint64 validTo)
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

// ParseAttestationIssued is a log parse operation binding the contract event 0xf25c7c8778fe6025840ac69375908fd6e7cd7ce7b90cd70535f09bf734372bc0.
//
// Solidity: event AttestationIssued(bytes32 indexed attestationId, bytes32 indexed issuerOrg, bytes32 indexed subject, bytes32 certType, bytes32 docHash, uint64 validFrom, uint64 validTo)
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
