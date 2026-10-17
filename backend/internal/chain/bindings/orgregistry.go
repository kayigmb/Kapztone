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

// OrgRegistryMetaData contains all meta data concerning the OrgRegistry contract.
var OrgRegistryMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"address\",\"name\":\"admin\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"relayer\",\"type\":\"address\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"inputs\":[],\"name\":\"AccessControlBadConfirmation\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"neededRole\",\"type\":\"bytes32\"}],\"name\":\"AccessControlUnauthorizedAccount\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint64\",\"name\":\"validFrom\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"validTo\",\"type\":\"uint64\"}],\"name\":\"BadValidity\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"orgId\",\"type\":\"bytes32\"}],\"name\":\"OrgExists\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"orgId\",\"type\":\"bytes32\"}],\"name\":\"OrgUnknown\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"signer\",\"type\":\"address\"}],\"name\":\"SignerInUse\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"ZeroValue\",\"type\":\"error\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"orgId\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"certType\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"validFrom\",\"type\":\"uint64\"},{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"validTo\",\"type\":\"uint64\"}],\"name\":\"IssuerAccredited\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"orgId\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"signer\",\"type\":\"address\"}],\"name\":\"OrgRegistered\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"orgId\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"signer\",\"type\":\"address\"}],\"name\":\"OrgSignerChanged\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"orgId\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"bool\",\"name\":\"verified\",\"type\":\"bool\"}],\"name\":\"OrgVerified\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"previousAdminRole\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"newAdminRole\",\"type\":\"bytes32\"}],\"name\":\"RoleAdminChanged\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"}],\"name\":\"RoleGranted\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"}],\"name\":\"RoleRevoked\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"DEFAULT_ADMIN_ROLE\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"OPERATOR_ROLE\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"RELAYER_ROLE\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"}],\"name\":\"getRoleAdmin\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"grantRole\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"hasRole\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"orgId\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"certType\",\"type\":\"bytes32\"}],\"name\":\"isAccredited\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"orgId\",\"type\":\"bytes32\"}],\"name\":\"isVerified\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"signer\",\"type\":\"address\"}],\"name\":\"orgOfSigner\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"orgId\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"orgId\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"signer\",\"type\":\"address\"}],\"name\":\"register\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"callerConfirmation\",\"type\":\"address\"}],\"name\":\"renounceRole\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"revokeRole\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"orgId\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"certType\",\"type\":\"bytes32\"},{\"internalType\":\"uint64\",\"name\":\"validFrom\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"validTo\",\"type\":\"uint64\"}],\"name\":\"setAccreditation\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"orgId\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"signer\",\"type\":\"address\"}],\"name\":\"setSigner\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"orgId\",\"type\":\"bytes32\"},{\"internalType\":\"bool\",\"name\":\"verified\",\"type\":\"bool\"}],\"name\":\"setVerified\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"orgId\",\"type\":\"bytes32\"}],\"name\":\"signerOf\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes4\",\"name\":\"interfaceId\",\"type\":\"bytes4\"}],\"name\":\"supportsInterface\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]",
	Bin: "0x608060405234801561000f575f5ffd5b50604051610ce6380380610ce683398101604081905261002e91610195565b6001600160a01b038216158061004b57506001600160a01b038116155b1561006957604051637c946ed760e01b815260040160405180910390fd5b6100735f836100d1565b5061009e7fe2b7fb3b832174769106daebcfd6d1970523240dda11281102db9363b83b0dc4826100d1565b506100c97f97667070c54ef182b0f5858b034beac1b6f3089aa2d3188bb1e8929f4fa9b929826100d1565b5050506101c6565b5f828152602081815260408083206001600160a01b038516845290915281205460ff16610171575f838152602081815260408083206001600160a01b03861684529091529020805460ff191660011790556101293390565b6001600160a01b0316826001600160a01b0316847f2f8788117e7eff1d82e926ec794901d17c78024a50270940304540a733656f0d60405160405180910390a4506001610174565b505f5b92915050565b80516001600160a01b0381168114610190575f5ffd5b919050565b5f5f604083850312156101a6575f5ffd5b6101af8361017a565b91506101bd6020840161017a565b90509250929050565b610b13806101d35f395ff3fe608060405234801561000f575f5ffd5b5060043610610106575f3560e01c8063926d7d7f1161009e578063c181b2731161006e578063c181b27314610263578063ca839c5b1461028c578063d22057a91461029f578063d547741f146102b2578063f5b541a6146102c5575f5ffd5b8063926d7d7f146101e25780639823e68614610209578063a217fddf14610249578063ad00dde014610250575f5ffd5b806336568abe116100d957806336568abe1461018a5780633a17eb591461019d5780634354647e146101bc57806391d14854146101cf575f5ffd5b806301ffc9a71461010a5780631122e86a14610132578063248a9ca3146101475780632f2ff15d14610177575b5f5ffd5b61011d610118366004610973565b6102d9565b60405190151581526020015b60405180910390f35b6101456101403660046109a1565b61030f565b005b6101696101553660046109d3565b5f9081526020819052604090206001015490565b604051908152602001610129565b610145610185366004610a00565b610386565b610145610198366004610a00565b6103b0565b6101696101ab366004610a2a565b60036020525f908152604090205481565b6101456101ca366004610a5a565b6103e8565b61011d6101dd366004610a00565b61055a565b6101697fe2b7fb3b832174769106daebcfd6d1970523240dda11281102db9363b83b0dc481565b6102316102173660046109d3565b5f908152600160205260409020546001600160a01b031690565b6040516001600160a01b039091168152602001610129565b6101695f81565b61011d61025e366004610a9d565b610582565b61011d6102713660046109d3565b5f90815260016020526040902054600160a01b900460ff1690565b61014561029a366004610a00565b6105ec565b6101456102ad366004610a00565b61066c565b6101456102c0366004610a00565b610730565b6101695f516020610abe5f395f51905f5281565b5f6001600160e01b03198216637965db0b60e01b148061030957506301ffc9a760e01b6001600160e01b03198316145b92915050565b5f516020610abe5f395f51905f5261032681610754565b8161033084610761565b8054911515600160a01b0260ff60a01b19909216919091179055604051821515815283907f2d7172d230d58c7ea2dc75b0d91921bba67143a4dd2db824c9db79cc5dedf9479060200160405180910390a2505050565b5f828152602081905260409020600101546103a081610754565b6103aa838361079e565b50505050565b6001600160a01b03811633146103d95760405163334bd91960e11b815260040160405180910390fd5b6103e3828261082d565b505050565b5f516020610abe5f395f51905f526103ff81610754565b61040885610761565b508361042757604051637c946ed760e01b815260040160405180910390fd5b8167ffffffffffffffff165f03610463575f858152600260209081526040808320878452909152902080546001600160801b031916905561050b565b8267ffffffffffffffff168267ffffffffffffffff16116104b0576040516355da959560e11b815267ffffffffffffffff8085166004830152831660248201526044015b60405180910390fd5b60408051808201825267ffffffffffffffff808616825284811660208084019182525f8a8152600282528581208a8252909152939093209151825493518216600160401b026001600160801b03199094169116179190911790555b6040805167ffffffffffffffff808616825284166020820152859187917fc134584a084803f245fd1d6e6f39dd2294a150d4c42922cd77c9e6b4141d98ce910160405180910390a35050505050565b5f918252602082815260408084206001600160a01b0393909316845291905290205460ff1690565b5f82815260026020908152604080832084845282528083208584526001909252822054600160a01b900460ff1680156105c65750805467ffffffffffffffff164210155b80156105e457508054600160401b900467ffffffffffffffff164211155b949350505050565b5f516020610abe5f395f51905f5261060381610754565b5f61060d84610761565b546001600160a01b03165f8181526003602052604081205590506106318484610896565b6040516001600160a01b0384169085907f26b16e183a2e89c88cddba8dcbbb48e189211b539ca87798573a4c02c7a3452a905f90a350505050565b7fe2b7fb3b832174769106daebcfd6d1970523240dda11281102db9363b83b0dc461069681610754565b826106b457604051637c946ed760e01b815260040160405180910390fd5b5f838152600160205260409020546001600160a01b0316156106ec57604051638201971760e01b8152600481018490526024016104a7565b6106f68383610896565b6040516001600160a01b0383169084907f7d53b93d471795182329253ee77657987f08c389449dab4ff1cf4b22b16c087a905f90a3505050565b5f8281526020819052604090206001015461074a81610754565b6103aa838361082d565b61075e8133610936565b50565b5f81815260016020526040902080546001600160a01b031661079957604051635d2ea2a960e11b8152600481018390526024016104a7565b919050565b5f6107a9838361055a565b610826575f838152602081815260408083206001600160a01b03861684529091529020805460ff191660011790556107de3390565b6001600160a01b0316826001600160a01b0316847f2f8788117e7eff1d82e926ec794901d17c78024a50270940304540a733656f0d60405160405180910390a4506001610309565b505f610309565b5f610838838361055a565b15610826575f838152602081815260408083206001600160a01b0386168085529252808320805460ff1916905551339286917ff6391f5c32d9c69d2a47ea670b442974b53935d1edc7fd64eb21e047a839171b9190a4506001610309565b6001600160a01b0381166108bd57604051637c946ed760e01b815260040160405180910390fd5b6001600160a01b0381165f90815260036020526040902054156108fe57604051633dd2866360e21b81526001600160a01b03821660048201526024016104a7565b5f82815260016020908152604080832080546001600160a01b039095166001600160a01b031990951685179055928252600390522055565b610940828261055a565b61096f5760405163e2517d3f60e01b81526001600160a01b0382166004820152602481018390526044016104a7565b5050565b5f60208284031215610983575f5ffd5b81356001600160e01b03198116811461099a575f5ffd5b9392505050565b5f5f604083850312156109b2575f5ffd5b82359150602083013580151581146109c8575f5ffd5b809150509250929050565b5f602082840312156109e3575f5ffd5b5035919050565b80356001600160a01b0381168114610799575f5ffd5b5f5f60408385031215610a11575f5ffd5b82359150610a21602084016109ea565b90509250929050565b5f60208284031215610a3a575f5ffd5b61099a826109ea565b803567ffffffffffffffff81168114610799575f5ffd5b5f5f5f5f60808587031215610a6d575f5ffd5b8435935060208501359250610a8460408601610a43565b9150610a9260608601610a43565b905092959194509250565b5f5f60408385031215610aae575f5ffd5b5050803592602090910135915056fe97667070c54ef182b0f5858b034beac1b6f3089aa2d3188bb1e8929f4fa9b929a2646970667358221220f531f571a84f67fb029cd17c5d9beaee3fc6adc8afaad8a27bfaf83c50f9f94c64736f6c63430008220033",
}

// OrgRegistryABI is the input ABI used to generate the binding from.
// Deprecated: Use OrgRegistryMetaData.ABI instead.
var OrgRegistryABI = OrgRegistryMetaData.ABI

// OrgRegistryBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use OrgRegistryMetaData.Bin instead.
var OrgRegistryBin = OrgRegistryMetaData.Bin

// DeployOrgRegistry deploys a new Ethereum contract, binding an instance of OrgRegistry to it.
func DeployOrgRegistry(auth *bind.TransactOpts, backend bind.ContractBackend, admin common.Address, relayer common.Address) (common.Address, *types.Transaction, *OrgRegistry, error) {
	parsed, err := OrgRegistryMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(OrgRegistryBin), backend, admin, relayer)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &OrgRegistry{OrgRegistryCaller: OrgRegistryCaller{contract: contract}, OrgRegistryTransactor: OrgRegistryTransactor{contract: contract}, OrgRegistryFilterer: OrgRegistryFilterer{contract: contract}}, nil
}

// OrgRegistry is an auto generated Go binding around an Ethereum contract.
type OrgRegistry struct {
	OrgRegistryCaller     // Read-only binding to the contract
	OrgRegistryTransactor // Write-only binding to the contract
	OrgRegistryFilterer   // Log filterer for contract events
}

// OrgRegistryCaller is an auto generated read-only Go binding around an Ethereum contract.
type OrgRegistryCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// OrgRegistryTransactor is an auto generated write-only Go binding around an Ethereum contract.
type OrgRegistryTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// OrgRegistryFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type OrgRegistryFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// OrgRegistrySession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type OrgRegistrySession struct {
	Contract     *OrgRegistry      // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// OrgRegistryCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type OrgRegistryCallerSession struct {
	Contract *OrgRegistryCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts      // Call options to use throughout this session
}

// OrgRegistryTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type OrgRegistryTransactorSession struct {
	Contract     *OrgRegistryTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts      // Transaction auth options to use throughout this session
}

// OrgRegistryRaw is an auto generated low-level Go binding around an Ethereum contract.
type OrgRegistryRaw struct {
	Contract *OrgRegistry // Generic contract binding to access the raw methods on
}

// OrgRegistryCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type OrgRegistryCallerRaw struct {
	Contract *OrgRegistryCaller // Generic read-only contract binding to access the raw methods on
}

// OrgRegistryTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type OrgRegistryTransactorRaw struct {
	Contract *OrgRegistryTransactor // Generic write-only contract binding to access the raw methods on
}

// NewOrgRegistry creates a new instance of OrgRegistry, bound to a specific deployed contract.
func NewOrgRegistry(address common.Address, backend bind.ContractBackend) (*OrgRegistry, error) {
	contract, err := bindOrgRegistry(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &OrgRegistry{OrgRegistryCaller: OrgRegistryCaller{contract: contract}, OrgRegistryTransactor: OrgRegistryTransactor{contract: contract}, OrgRegistryFilterer: OrgRegistryFilterer{contract: contract}}, nil
}

// NewOrgRegistryCaller creates a new read-only instance of OrgRegistry, bound to a specific deployed contract.
func NewOrgRegistryCaller(address common.Address, caller bind.ContractCaller) (*OrgRegistryCaller, error) {
	contract, err := bindOrgRegistry(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &OrgRegistryCaller{contract: contract}, nil
}

// NewOrgRegistryTransactor creates a new write-only instance of OrgRegistry, bound to a specific deployed contract.
func NewOrgRegistryTransactor(address common.Address, transactor bind.ContractTransactor) (*OrgRegistryTransactor, error) {
	contract, err := bindOrgRegistry(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &OrgRegistryTransactor{contract: contract}, nil
}

// NewOrgRegistryFilterer creates a new log filterer instance of OrgRegistry, bound to a specific deployed contract.
func NewOrgRegistryFilterer(address common.Address, filterer bind.ContractFilterer) (*OrgRegistryFilterer, error) {
	contract, err := bindOrgRegistry(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &OrgRegistryFilterer{contract: contract}, nil
}

// bindOrgRegistry binds a generic wrapper to an already deployed contract.
func bindOrgRegistry(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := OrgRegistryMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_OrgRegistry *OrgRegistryRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _OrgRegistry.Contract.OrgRegistryCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_OrgRegistry *OrgRegistryRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _OrgRegistry.Contract.OrgRegistryTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_OrgRegistry *OrgRegistryRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _OrgRegistry.Contract.OrgRegistryTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_OrgRegistry *OrgRegistryCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _OrgRegistry.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_OrgRegistry *OrgRegistryTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _OrgRegistry.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_OrgRegistry *OrgRegistryTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _OrgRegistry.Contract.contract.Transact(opts, method, params...)
}

// DEFAULTADMINROLE is a free data retrieval call binding the contract method 0xa217fddf.
//
// Solidity: function DEFAULT_ADMIN_ROLE() view returns(bytes32)
func (_OrgRegistry *OrgRegistryCaller) DEFAULTADMINROLE(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _OrgRegistry.contract.Call(opts, &out, "DEFAULT_ADMIN_ROLE")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// DEFAULTADMINROLE is a free data retrieval call binding the contract method 0xa217fddf.
//
// Solidity: function DEFAULT_ADMIN_ROLE() view returns(bytes32)
func (_OrgRegistry *OrgRegistrySession) DEFAULTADMINROLE() ([32]byte, error) {
	return _OrgRegistry.Contract.DEFAULTADMINROLE(&_OrgRegistry.CallOpts)
}

// DEFAULTADMINROLE is a free data retrieval call binding the contract method 0xa217fddf.
//
// Solidity: function DEFAULT_ADMIN_ROLE() view returns(bytes32)
func (_OrgRegistry *OrgRegistryCallerSession) DEFAULTADMINROLE() ([32]byte, error) {
	return _OrgRegistry.Contract.DEFAULTADMINROLE(&_OrgRegistry.CallOpts)
}

// OPERATORROLE is a free data retrieval call binding the contract method 0xf5b541a6.
//
// Solidity: function OPERATOR_ROLE() view returns(bytes32)
func (_OrgRegistry *OrgRegistryCaller) OPERATORROLE(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _OrgRegistry.contract.Call(opts, &out, "OPERATOR_ROLE")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// OPERATORROLE is a free data retrieval call binding the contract method 0xf5b541a6.
//
// Solidity: function OPERATOR_ROLE() view returns(bytes32)
func (_OrgRegistry *OrgRegistrySession) OPERATORROLE() ([32]byte, error) {
	return _OrgRegistry.Contract.OPERATORROLE(&_OrgRegistry.CallOpts)
}

// OPERATORROLE is a free data retrieval call binding the contract method 0xf5b541a6.
//
// Solidity: function OPERATOR_ROLE() view returns(bytes32)
func (_OrgRegistry *OrgRegistryCallerSession) OPERATORROLE() ([32]byte, error) {
	return _OrgRegistry.Contract.OPERATORROLE(&_OrgRegistry.CallOpts)
}

// RELAYERROLE is a free data retrieval call binding the contract method 0x926d7d7f.
//
// Solidity: function RELAYER_ROLE() view returns(bytes32)
func (_OrgRegistry *OrgRegistryCaller) RELAYERROLE(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _OrgRegistry.contract.Call(opts, &out, "RELAYER_ROLE")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// RELAYERROLE is a free data retrieval call binding the contract method 0x926d7d7f.
//
// Solidity: function RELAYER_ROLE() view returns(bytes32)
func (_OrgRegistry *OrgRegistrySession) RELAYERROLE() ([32]byte, error) {
	return _OrgRegistry.Contract.RELAYERROLE(&_OrgRegistry.CallOpts)
}

// RELAYERROLE is a free data retrieval call binding the contract method 0x926d7d7f.
//
// Solidity: function RELAYER_ROLE() view returns(bytes32)
func (_OrgRegistry *OrgRegistryCallerSession) RELAYERROLE() ([32]byte, error) {
	return _OrgRegistry.Contract.RELAYERROLE(&_OrgRegistry.CallOpts)
}

// GetRoleAdmin is a free data retrieval call binding the contract method 0x248a9ca3.
//
// Solidity: function getRoleAdmin(bytes32 role) view returns(bytes32)
func (_OrgRegistry *OrgRegistryCaller) GetRoleAdmin(opts *bind.CallOpts, role [32]byte) ([32]byte, error) {
	var out []interface{}
	err := _OrgRegistry.contract.Call(opts, &out, "getRoleAdmin", role)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// GetRoleAdmin is a free data retrieval call binding the contract method 0x248a9ca3.
//
// Solidity: function getRoleAdmin(bytes32 role) view returns(bytes32)
func (_OrgRegistry *OrgRegistrySession) GetRoleAdmin(role [32]byte) ([32]byte, error) {
	return _OrgRegistry.Contract.GetRoleAdmin(&_OrgRegistry.CallOpts, role)
}

// GetRoleAdmin is a free data retrieval call binding the contract method 0x248a9ca3.
//
// Solidity: function getRoleAdmin(bytes32 role) view returns(bytes32)
func (_OrgRegistry *OrgRegistryCallerSession) GetRoleAdmin(role [32]byte) ([32]byte, error) {
	return _OrgRegistry.Contract.GetRoleAdmin(&_OrgRegistry.CallOpts, role)
}

// HasRole is a free data retrieval call binding the contract method 0x91d14854.
//
// Solidity: function hasRole(bytes32 role, address account) view returns(bool)
func (_OrgRegistry *OrgRegistryCaller) HasRole(opts *bind.CallOpts, role [32]byte, account common.Address) (bool, error) {
	var out []interface{}
	err := _OrgRegistry.contract.Call(opts, &out, "hasRole", role, account)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// HasRole is a free data retrieval call binding the contract method 0x91d14854.
//
// Solidity: function hasRole(bytes32 role, address account) view returns(bool)
func (_OrgRegistry *OrgRegistrySession) HasRole(role [32]byte, account common.Address) (bool, error) {
	return _OrgRegistry.Contract.HasRole(&_OrgRegistry.CallOpts, role, account)
}

// HasRole is a free data retrieval call binding the contract method 0x91d14854.
//
// Solidity: function hasRole(bytes32 role, address account) view returns(bool)
func (_OrgRegistry *OrgRegistryCallerSession) HasRole(role [32]byte, account common.Address) (bool, error) {
	return _OrgRegistry.Contract.HasRole(&_OrgRegistry.CallOpts, role, account)
}

// IsAccredited is a free data retrieval call binding the contract method 0xad00dde0.
//
// Solidity: function isAccredited(bytes32 orgId, bytes32 certType) view returns(bool)
func (_OrgRegistry *OrgRegistryCaller) IsAccredited(opts *bind.CallOpts, orgId [32]byte, certType [32]byte) (bool, error) {
	var out []interface{}
	err := _OrgRegistry.contract.Call(opts, &out, "isAccredited", orgId, certType)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsAccredited is a free data retrieval call binding the contract method 0xad00dde0.
//
// Solidity: function isAccredited(bytes32 orgId, bytes32 certType) view returns(bool)
func (_OrgRegistry *OrgRegistrySession) IsAccredited(orgId [32]byte, certType [32]byte) (bool, error) {
	return _OrgRegistry.Contract.IsAccredited(&_OrgRegistry.CallOpts, orgId, certType)
}

// IsAccredited is a free data retrieval call binding the contract method 0xad00dde0.
//
// Solidity: function isAccredited(bytes32 orgId, bytes32 certType) view returns(bool)
func (_OrgRegistry *OrgRegistryCallerSession) IsAccredited(orgId [32]byte, certType [32]byte) (bool, error) {
	return _OrgRegistry.Contract.IsAccredited(&_OrgRegistry.CallOpts, orgId, certType)
}

// IsVerified is a free data retrieval call binding the contract method 0xc181b273.
//
// Solidity: function isVerified(bytes32 orgId) view returns(bool)
func (_OrgRegistry *OrgRegistryCaller) IsVerified(opts *bind.CallOpts, orgId [32]byte) (bool, error) {
	var out []interface{}
	err := _OrgRegistry.contract.Call(opts, &out, "isVerified", orgId)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsVerified is a free data retrieval call binding the contract method 0xc181b273.
//
// Solidity: function isVerified(bytes32 orgId) view returns(bool)
func (_OrgRegistry *OrgRegistrySession) IsVerified(orgId [32]byte) (bool, error) {
	return _OrgRegistry.Contract.IsVerified(&_OrgRegistry.CallOpts, orgId)
}

// IsVerified is a free data retrieval call binding the contract method 0xc181b273.
//
// Solidity: function isVerified(bytes32 orgId) view returns(bool)
func (_OrgRegistry *OrgRegistryCallerSession) IsVerified(orgId [32]byte) (bool, error) {
	return _OrgRegistry.Contract.IsVerified(&_OrgRegistry.CallOpts, orgId)
}

// OrgOfSigner is a free data retrieval call binding the contract method 0x3a17eb59.
//
// Solidity: function orgOfSigner(address signer) view returns(bytes32 orgId)
func (_OrgRegistry *OrgRegistryCaller) OrgOfSigner(opts *bind.CallOpts, signer common.Address) ([32]byte, error) {
	var out []interface{}
	err := _OrgRegistry.contract.Call(opts, &out, "orgOfSigner", signer)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// OrgOfSigner is a free data retrieval call binding the contract method 0x3a17eb59.
//
// Solidity: function orgOfSigner(address signer) view returns(bytes32 orgId)
func (_OrgRegistry *OrgRegistrySession) OrgOfSigner(signer common.Address) ([32]byte, error) {
	return _OrgRegistry.Contract.OrgOfSigner(&_OrgRegistry.CallOpts, signer)
}

// OrgOfSigner is a free data retrieval call binding the contract method 0x3a17eb59.
//
// Solidity: function orgOfSigner(address signer) view returns(bytes32 orgId)
func (_OrgRegistry *OrgRegistryCallerSession) OrgOfSigner(signer common.Address) ([32]byte, error) {
	return _OrgRegistry.Contract.OrgOfSigner(&_OrgRegistry.CallOpts, signer)
}

// SignerOf is a free data retrieval call binding the contract method 0x9823e686.
//
// Solidity: function signerOf(bytes32 orgId) view returns(address)
func (_OrgRegistry *OrgRegistryCaller) SignerOf(opts *bind.CallOpts, orgId [32]byte) (common.Address, error) {
	var out []interface{}
	err := _OrgRegistry.contract.Call(opts, &out, "signerOf", orgId)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// SignerOf is a free data retrieval call binding the contract method 0x9823e686.
//
// Solidity: function signerOf(bytes32 orgId) view returns(address)
func (_OrgRegistry *OrgRegistrySession) SignerOf(orgId [32]byte) (common.Address, error) {
	return _OrgRegistry.Contract.SignerOf(&_OrgRegistry.CallOpts, orgId)
}

// SignerOf is a free data retrieval call binding the contract method 0x9823e686.
//
// Solidity: function signerOf(bytes32 orgId) view returns(address)
func (_OrgRegistry *OrgRegistryCallerSession) SignerOf(orgId [32]byte) (common.Address, error) {
	return _OrgRegistry.Contract.SignerOf(&_OrgRegistry.CallOpts, orgId)
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_OrgRegistry *OrgRegistryCaller) SupportsInterface(opts *bind.CallOpts, interfaceId [4]byte) (bool, error) {
	var out []interface{}
	err := _OrgRegistry.contract.Call(opts, &out, "supportsInterface", interfaceId)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_OrgRegistry *OrgRegistrySession) SupportsInterface(interfaceId [4]byte) (bool, error) {
	return _OrgRegistry.Contract.SupportsInterface(&_OrgRegistry.CallOpts, interfaceId)
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_OrgRegistry *OrgRegistryCallerSession) SupportsInterface(interfaceId [4]byte) (bool, error) {
	return _OrgRegistry.Contract.SupportsInterface(&_OrgRegistry.CallOpts, interfaceId)
}

// GrantRole is a paid mutator transaction binding the contract method 0x2f2ff15d.
//
// Solidity: function grantRole(bytes32 role, address account) returns()
func (_OrgRegistry *OrgRegistryTransactor) GrantRole(opts *bind.TransactOpts, role [32]byte, account common.Address) (*types.Transaction, error) {
	return _OrgRegistry.contract.Transact(opts, "grantRole", role, account)
}

// GrantRole is a paid mutator transaction binding the contract method 0x2f2ff15d.
//
// Solidity: function grantRole(bytes32 role, address account) returns()
func (_OrgRegistry *OrgRegistrySession) GrantRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _OrgRegistry.Contract.GrantRole(&_OrgRegistry.TransactOpts, role, account)
}

// GrantRole is a paid mutator transaction binding the contract method 0x2f2ff15d.
//
// Solidity: function grantRole(bytes32 role, address account) returns()
func (_OrgRegistry *OrgRegistryTransactorSession) GrantRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _OrgRegistry.Contract.GrantRole(&_OrgRegistry.TransactOpts, role, account)
}

// Register is a paid mutator transaction binding the contract method 0xd22057a9.
//
// Solidity: function register(bytes32 orgId, address signer) returns()
func (_OrgRegistry *OrgRegistryTransactor) Register(opts *bind.TransactOpts, orgId [32]byte, signer common.Address) (*types.Transaction, error) {
	return _OrgRegistry.contract.Transact(opts, "register", orgId, signer)
}

// Register is a paid mutator transaction binding the contract method 0xd22057a9.
//
// Solidity: function register(bytes32 orgId, address signer) returns()
func (_OrgRegistry *OrgRegistrySession) Register(orgId [32]byte, signer common.Address) (*types.Transaction, error) {
	return _OrgRegistry.Contract.Register(&_OrgRegistry.TransactOpts, orgId, signer)
}

// Register is a paid mutator transaction binding the contract method 0xd22057a9.
//
// Solidity: function register(bytes32 orgId, address signer) returns()
func (_OrgRegistry *OrgRegistryTransactorSession) Register(orgId [32]byte, signer common.Address) (*types.Transaction, error) {
	return _OrgRegistry.Contract.Register(&_OrgRegistry.TransactOpts, orgId, signer)
}

// RenounceRole is a paid mutator transaction binding the contract method 0x36568abe.
//
// Solidity: function renounceRole(bytes32 role, address callerConfirmation) returns()
func (_OrgRegistry *OrgRegistryTransactor) RenounceRole(opts *bind.TransactOpts, role [32]byte, callerConfirmation common.Address) (*types.Transaction, error) {
	return _OrgRegistry.contract.Transact(opts, "renounceRole", role, callerConfirmation)
}

// RenounceRole is a paid mutator transaction binding the contract method 0x36568abe.
//
// Solidity: function renounceRole(bytes32 role, address callerConfirmation) returns()
func (_OrgRegistry *OrgRegistrySession) RenounceRole(role [32]byte, callerConfirmation common.Address) (*types.Transaction, error) {
	return _OrgRegistry.Contract.RenounceRole(&_OrgRegistry.TransactOpts, role, callerConfirmation)
}

// RenounceRole is a paid mutator transaction binding the contract method 0x36568abe.
//
// Solidity: function renounceRole(bytes32 role, address callerConfirmation) returns()
func (_OrgRegistry *OrgRegistryTransactorSession) RenounceRole(role [32]byte, callerConfirmation common.Address) (*types.Transaction, error) {
	return _OrgRegistry.Contract.RenounceRole(&_OrgRegistry.TransactOpts, role, callerConfirmation)
}

// RevokeRole is a paid mutator transaction binding the contract method 0xd547741f.
//
// Solidity: function revokeRole(bytes32 role, address account) returns()
func (_OrgRegistry *OrgRegistryTransactor) RevokeRole(opts *bind.TransactOpts, role [32]byte, account common.Address) (*types.Transaction, error) {
	return _OrgRegistry.contract.Transact(opts, "revokeRole", role, account)
}

// RevokeRole is a paid mutator transaction binding the contract method 0xd547741f.
//
// Solidity: function revokeRole(bytes32 role, address account) returns()
func (_OrgRegistry *OrgRegistrySession) RevokeRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _OrgRegistry.Contract.RevokeRole(&_OrgRegistry.TransactOpts, role, account)
}

// RevokeRole is a paid mutator transaction binding the contract method 0xd547741f.
//
// Solidity: function revokeRole(bytes32 role, address account) returns()
func (_OrgRegistry *OrgRegistryTransactorSession) RevokeRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _OrgRegistry.Contract.RevokeRole(&_OrgRegistry.TransactOpts, role, account)
}

// SetAccreditation is a paid mutator transaction binding the contract method 0x4354647e.
//
// Solidity: function setAccreditation(bytes32 orgId, bytes32 certType, uint64 validFrom, uint64 validTo) returns()
func (_OrgRegistry *OrgRegistryTransactor) SetAccreditation(opts *bind.TransactOpts, orgId [32]byte, certType [32]byte, validFrom uint64, validTo uint64) (*types.Transaction, error) {
	return _OrgRegistry.contract.Transact(opts, "setAccreditation", orgId, certType, validFrom, validTo)
}

// SetAccreditation is a paid mutator transaction binding the contract method 0x4354647e.
//
// Solidity: function setAccreditation(bytes32 orgId, bytes32 certType, uint64 validFrom, uint64 validTo) returns()
func (_OrgRegistry *OrgRegistrySession) SetAccreditation(orgId [32]byte, certType [32]byte, validFrom uint64, validTo uint64) (*types.Transaction, error) {
	return _OrgRegistry.Contract.SetAccreditation(&_OrgRegistry.TransactOpts, orgId, certType, validFrom, validTo)
}

// SetAccreditation is a paid mutator transaction binding the contract method 0x4354647e.
//
// Solidity: function setAccreditation(bytes32 orgId, bytes32 certType, uint64 validFrom, uint64 validTo) returns()
func (_OrgRegistry *OrgRegistryTransactorSession) SetAccreditation(orgId [32]byte, certType [32]byte, validFrom uint64, validTo uint64) (*types.Transaction, error) {
	return _OrgRegistry.Contract.SetAccreditation(&_OrgRegistry.TransactOpts, orgId, certType, validFrom, validTo)
}

// SetSigner is a paid mutator transaction binding the contract method 0xca839c5b.
//
// Solidity: function setSigner(bytes32 orgId, address signer) returns()
func (_OrgRegistry *OrgRegistryTransactor) SetSigner(opts *bind.TransactOpts, orgId [32]byte, signer common.Address) (*types.Transaction, error) {
	return _OrgRegistry.contract.Transact(opts, "setSigner", orgId, signer)
}

// SetSigner is a paid mutator transaction binding the contract method 0xca839c5b.
//
// Solidity: function setSigner(bytes32 orgId, address signer) returns()
func (_OrgRegistry *OrgRegistrySession) SetSigner(orgId [32]byte, signer common.Address) (*types.Transaction, error) {
	return _OrgRegistry.Contract.SetSigner(&_OrgRegistry.TransactOpts, orgId, signer)
}

// SetSigner is a paid mutator transaction binding the contract method 0xca839c5b.
//
// Solidity: function setSigner(bytes32 orgId, address signer) returns()
func (_OrgRegistry *OrgRegistryTransactorSession) SetSigner(orgId [32]byte, signer common.Address) (*types.Transaction, error) {
	return _OrgRegistry.Contract.SetSigner(&_OrgRegistry.TransactOpts, orgId, signer)
}

// SetVerified is a paid mutator transaction binding the contract method 0x1122e86a.
//
// Solidity: function setVerified(bytes32 orgId, bool verified) returns()
func (_OrgRegistry *OrgRegistryTransactor) SetVerified(opts *bind.TransactOpts, orgId [32]byte, verified bool) (*types.Transaction, error) {
	return _OrgRegistry.contract.Transact(opts, "setVerified", orgId, verified)
}

// SetVerified is a paid mutator transaction binding the contract method 0x1122e86a.
//
// Solidity: function setVerified(bytes32 orgId, bool verified) returns()
func (_OrgRegistry *OrgRegistrySession) SetVerified(orgId [32]byte, verified bool) (*types.Transaction, error) {
	return _OrgRegistry.Contract.SetVerified(&_OrgRegistry.TransactOpts, orgId, verified)
}

// SetVerified is a paid mutator transaction binding the contract method 0x1122e86a.
//
// Solidity: function setVerified(bytes32 orgId, bool verified) returns()
func (_OrgRegistry *OrgRegistryTransactorSession) SetVerified(orgId [32]byte, verified bool) (*types.Transaction, error) {
	return _OrgRegistry.Contract.SetVerified(&_OrgRegistry.TransactOpts, orgId, verified)
}

// OrgRegistryIssuerAccreditedIterator is returned from FilterIssuerAccredited and is used to iterate over the raw logs and unpacked data for IssuerAccredited events raised by the OrgRegistry contract.
type OrgRegistryIssuerAccreditedIterator struct {
	Event *OrgRegistryIssuerAccredited // Event containing the contract specifics and raw log

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
func (it *OrgRegistryIssuerAccreditedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(OrgRegistryIssuerAccredited)
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
		it.Event = new(OrgRegistryIssuerAccredited)
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
func (it *OrgRegistryIssuerAccreditedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *OrgRegistryIssuerAccreditedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// OrgRegistryIssuerAccredited represents a IssuerAccredited event raised by the OrgRegistry contract.
type OrgRegistryIssuerAccredited struct {
	OrgId     [32]byte
	CertType  [32]byte
	ValidFrom uint64
	ValidTo   uint64
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterIssuerAccredited is a free log retrieval operation binding the contract event 0xc134584a084803f245fd1d6e6f39dd2294a150d4c42922cd77c9e6b4141d98ce.
//
// Solidity: event IssuerAccredited(bytes32 indexed orgId, bytes32 indexed certType, uint64 validFrom, uint64 validTo)
func (_OrgRegistry *OrgRegistryFilterer) FilterIssuerAccredited(opts *bind.FilterOpts, orgId [][32]byte, certType [][32]byte) (*OrgRegistryIssuerAccreditedIterator, error) {

	var orgIdRule []interface{}
	for _, orgIdItem := range orgId {
		orgIdRule = append(orgIdRule, orgIdItem)
	}
	var certTypeRule []interface{}
	for _, certTypeItem := range certType {
		certTypeRule = append(certTypeRule, certTypeItem)
	}

	logs, sub, err := _OrgRegistry.contract.FilterLogs(opts, "IssuerAccredited", orgIdRule, certTypeRule)
	if err != nil {
		return nil, err
	}
	return &OrgRegistryIssuerAccreditedIterator{contract: _OrgRegistry.contract, event: "IssuerAccredited", logs: logs, sub: sub}, nil
}

// WatchIssuerAccredited is a free log subscription operation binding the contract event 0xc134584a084803f245fd1d6e6f39dd2294a150d4c42922cd77c9e6b4141d98ce.
//
// Solidity: event IssuerAccredited(bytes32 indexed orgId, bytes32 indexed certType, uint64 validFrom, uint64 validTo)
func (_OrgRegistry *OrgRegistryFilterer) WatchIssuerAccredited(opts *bind.WatchOpts, sink chan<- *OrgRegistryIssuerAccredited, orgId [][32]byte, certType [][32]byte) (event.Subscription, error) {

	var orgIdRule []interface{}
	for _, orgIdItem := range orgId {
		orgIdRule = append(orgIdRule, orgIdItem)
	}
	var certTypeRule []interface{}
	for _, certTypeItem := range certType {
		certTypeRule = append(certTypeRule, certTypeItem)
	}

	logs, sub, err := _OrgRegistry.contract.WatchLogs(opts, "IssuerAccredited", orgIdRule, certTypeRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(OrgRegistryIssuerAccredited)
				if err := _OrgRegistry.contract.UnpackLog(event, "IssuerAccredited", log); err != nil {
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

// ParseIssuerAccredited is a log parse operation binding the contract event 0xc134584a084803f245fd1d6e6f39dd2294a150d4c42922cd77c9e6b4141d98ce.
//
// Solidity: event IssuerAccredited(bytes32 indexed orgId, bytes32 indexed certType, uint64 validFrom, uint64 validTo)
func (_OrgRegistry *OrgRegistryFilterer) ParseIssuerAccredited(log types.Log) (*OrgRegistryIssuerAccredited, error) {
	event := new(OrgRegistryIssuerAccredited)
	if err := _OrgRegistry.contract.UnpackLog(event, "IssuerAccredited", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// OrgRegistryOrgRegisteredIterator is returned from FilterOrgRegistered and is used to iterate over the raw logs and unpacked data for OrgRegistered events raised by the OrgRegistry contract.
type OrgRegistryOrgRegisteredIterator struct {
	Event *OrgRegistryOrgRegistered // Event containing the contract specifics and raw log

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
func (it *OrgRegistryOrgRegisteredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(OrgRegistryOrgRegistered)
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
		it.Event = new(OrgRegistryOrgRegistered)
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
func (it *OrgRegistryOrgRegisteredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *OrgRegistryOrgRegisteredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// OrgRegistryOrgRegistered represents a OrgRegistered event raised by the OrgRegistry contract.
type OrgRegistryOrgRegistered struct {
	OrgId  [32]byte
	Signer common.Address
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterOrgRegistered is a free log retrieval operation binding the contract event 0x7d53b93d471795182329253ee77657987f08c389449dab4ff1cf4b22b16c087a.
//
// Solidity: event OrgRegistered(bytes32 indexed orgId, address indexed signer)
func (_OrgRegistry *OrgRegistryFilterer) FilterOrgRegistered(opts *bind.FilterOpts, orgId [][32]byte, signer []common.Address) (*OrgRegistryOrgRegisteredIterator, error) {

	var orgIdRule []interface{}
	for _, orgIdItem := range orgId {
		orgIdRule = append(orgIdRule, orgIdItem)
	}
	var signerRule []interface{}
	for _, signerItem := range signer {
		signerRule = append(signerRule, signerItem)
	}

	logs, sub, err := _OrgRegistry.contract.FilterLogs(opts, "OrgRegistered", orgIdRule, signerRule)
	if err != nil {
		return nil, err
	}
	return &OrgRegistryOrgRegisteredIterator{contract: _OrgRegistry.contract, event: "OrgRegistered", logs: logs, sub: sub}, nil
}

// WatchOrgRegistered is a free log subscription operation binding the contract event 0x7d53b93d471795182329253ee77657987f08c389449dab4ff1cf4b22b16c087a.
//
// Solidity: event OrgRegistered(bytes32 indexed orgId, address indexed signer)
func (_OrgRegistry *OrgRegistryFilterer) WatchOrgRegistered(opts *bind.WatchOpts, sink chan<- *OrgRegistryOrgRegistered, orgId [][32]byte, signer []common.Address) (event.Subscription, error) {

	var orgIdRule []interface{}
	for _, orgIdItem := range orgId {
		orgIdRule = append(orgIdRule, orgIdItem)
	}
	var signerRule []interface{}
	for _, signerItem := range signer {
		signerRule = append(signerRule, signerItem)
	}

	logs, sub, err := _OrgRegistry.contract.WatchLogs(opts, "OrgRegistered", orgIdRule, signerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(OrgRegistryOrgRegistered)
				if err := _OrgRegistry.contract.UnpackLog(event, "OrgRegistered", log); err != nil {
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

// ParseOrgRegistered is a log parse operation binding the contract event 0x7d53b93d471795182329253ee77657987f08c389449dab4ff1cf4b22b16c087a.
//
// Solidity: event OrgRegistered(bytes32 indexed orgId, address indexed signer)
func (_OrgRegistry *OrgRegistryFilterer) ParseOrgRegistered(log types.Log) (*OrgRegistryOrgRegistered, error) {
	event := new(OrgRegistryOrgRegistered)
	if err := _OrgRegistry.contract.UnpackLog(event, "OrgRegistered", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// OrgRegistryOrgSignerChangedIterator is returned from FilterOrgSignerChanged and is used to iterate over the raw logs and unpacked data for OrgSignerChanged events raised by the OrgRegistry contract.
type OrgRegistryOrgSignerChangedIterator struct {
	Event *OrgRegistryOrgSignerChanged // Event containing the contract specifics and raw log

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
func (it *OrgRegistryOrgSignerChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(OrgRegistryOrgSignerChanged)
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
		it.Event = new(OrgRegistryOrgSignerChanged)
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
func (it *OrgRegistryOrgSignerChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *OrgRegistryOrgSignerChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// OrgRegistryOrgSignerChanged represents a OrgSignerChanged event raised by the OrgRegistry contract.
type OrgRegistryOrgSignerChanged struct {
	OrgId  [32]byte
	Signer common.Address
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterOrgSignerChanged is a free log retrieval operation binding the contract event 0x26b16e183a2e89c88cddba8dcbbb48e189211b539ca87798573a4c02c7a3452a.
//
// Solidity: event OrgSignerChanged(bytes32 indexed orgId, address indexed signer)
func (_OrgRegistry *OrgRegistryFilterer) FilterOrgSignerChanged(opts *bind.FilterOpts, orgId [][32]byte, signer []common.Address) (*OrgRegistryOrgSignerChangedIterator, error) {

	var orgIdRule []interface{}
	for _, orgIdItem := range orgId {
		orgIdRule = append(orgIdRule, orgIdItem)
	}
	var signerRule []interface{}
	for _, signerItem := range signer {
		signerRule = append(signerRule, signerItem)
	}

	logs, sub, err := _OrgRegistry.contract.FilterLogs(opts, "OrgSignerChanged", orgIdRule, signerRule)
	if err != nil {
		return nil, err
	}
	return &OrgRegistryOrgSignerChangedIterator{contract: _OrgRegistry.contract, event: "OrgSignerChanged", logs: logs, sub: sub}, nil
}

// WatchOrgSignerChanged is a free log subscription operation binding the contract event 0x26b16e183a2e89c88cddba8dcbbb48e189211b539ca87798573a4c02c7a3452a.
//
// Solidity: event OrgSignerChanged(bytes32 indexed orgId, address indexed signer)
func (_OrgRegistry *OrgRegistryFilterer) WatchOrgSignerChanged(opts *bind.WatchOpts, sink chan<- *OrgRegistryOrgSignerChanged, orgId [][32]byte, signer []common.Address) (event.Subscription, error) {

	var orgIdRule []interface{}
	for _, orgIdItem := range orgId {
		orgIdRule = append(orgIdRule, orgIdItem)
	}
	var signerRule []interface{}
	for _, signerItem := range signer {
		signerRule = append(signerRule, signerItem)
	}

	logs, sub, err := _OrgRegistry.contract.WatchLogs(opts, "OrgSignerChanged", orgIdRule, signerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(OrgRegistryOrgSignerChanged)
				if err := _OrgRegistry.contract.UnpackLog(event, "OrgSignerChanged", log); err != nil {
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

// ParseOrgSignerChanged is a log parse operation binding the contract event 0x26b16e183a2e89c88cddba8dcbbb48e189211b539ca87798573a4c02c7a3452a.
//
// Solidity: event OrgSignerChanged(bytes32 indexed orgId, address indexed signer)
func (_OrgRegistry *OrgRegistryFilterer) ParseOrgSignerChanged(log types.Log) (*OrgRegistryOrgSignerChanged, error) {
	event := new(OrgRegistryOrgSignerChanged)
	if err := _OrgRegistry.contract.UnpackLog(event, "OrgSignerChanged", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// OrgRegistryOrgVerifiedIterator is returned from FilterOrgVerified and is used to iterate over the raw logs and unpacked data for OrgVerified events raised by the OrgRegistry contract.
type OrgRegistryOrgVerifiedIterator struct {
	Event *OrgRegistryOrgVerified // Event containing the contract specifics and raw log

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
func (it *OrgRegistryOrgVerifiedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(OrgRegistryOrgVerified)
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
		it.Event = new(OrgRegistryOrgVerified)
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
func (it *OrgRegistryOrgVerifiedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *OrgRegistryOrgVerifiedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// OrgRegistryOrgVerified represents a OrgVerified event raised by the OrgRegistry contract.
type OrgRegistryOrgVerified struct {
	OrgId    [32]byte
	Verified bool
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterOrgVerified is a free log retrieval operation binding the contract event 0x2d7172d230d58c7ea2dc75b0d91921bba67143a4dd2db824c9db79cc5dedf947.
//
// Solidity: event OrgVerified(bytes32 indexed orgId, bool verified)
func (_OrgRegistry *OrgRegistryFilterer) FilterOrgVerified(opts *bind.FilterOpts, orgId [][32]byte) (*OrgRegistryOrgVerifiedIterator, error) {

	var orgIdRule []interface{}
	for _, orgIdItem := range orgId {
		orgIdRule = append(orgIdRule, orgIdItem)
	}

	logs, sub, err := _OrgRegistry.contract.FilterLogs(opts, "OrgVerified", orgIdRule)
	if err != nil {
		return nil, err
	}
	return &OrgRegistryOrgVerifiedIterator{contract: _OrgRegistry.contract, event: "OrgVerified", logs: logs, sub: sub}, nil
}

// WatchOrgVerified is a free log subscription operation binding the contract event 0x2d7172d230d58c7ea2dc75b0d91921bba67143a4dd2db824c9db79cc5dedf947.
//
// Solidity: event OrgVerified(bytes32 indexed orgId, bool verified)
func (_OrgRegistry *OrgRegistryFilterer) WatchOrgVerified(opts *bind.WatchOpts, sink chan<- *OrgRegistryOrgVerified, orgId [][32]byte) (event.Subscription, error) {

	var orgIdRule []interface{}
	for _, orgIdItem := range orgId {
		orgIdRule = append(orgIdRule, orgIdItem)
	}

	logs, sub, err := _OrgRegistry.contract.WatchLogs(opts, "OrgVerified", orgIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(OrgRegistryOrgVerified)
				if err := _OrgRegistry.contract.UnpackLog(event, "OrgVerified", log); err != nil {
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

// ParseOrgVerified is a log parse operation binding the contract event 0x2d7172d230d58c7ea2dc75b0d91921bba67143a4dd2db824c9db79cc5dedf947.
//
// Solidity: event OrgVerified(bytes32 indexed orgId, bool verified)
func (_OrgRegistry *OrgRegistryFilterer) ParseOrgVerified(log types.Log) (*OrgRegistryOrgVerified, error) {
	event := new(OrgRegistryOrgVerified)
	if err := _OrgRegistry.contract.UnpackLog(event, "OrgVerified", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// OrgRegistryRoleAdminChangedIterator is returned from FilterRoleAdminChanged and is used to iterate over the raw logs and unpacked data for RoleAdminChanged events raised by the OrgRegistry contract.
type OrgRegistryRoleAdminChangedIterator struct {
	Event *OrgRegistryRoleAdminChanged // Event containing the contract specifics and raw log

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
func (it *OrgRegistryRoleAdminChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(OrgRegistryRoleAdminChanged)
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
		it.Event = new(OrgRegistryRoleAdminChanged)
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
func (it *OrgRegistryRoleAdminChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *OrgRegistryRoleAdminChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// OrgRegistryRoleAdminChanged represents a RoleAdminChanged event raised by the OrgRegistry contract.
type OrgRegistryRoleAdminChanged struct {
	Role              [32]byte
	PreviousAdminRole [32]byte
	NewAdminRole      [32]byte
	Raw               types.Log // Blockchain specific contextual infos
}

// FilterRoleAdminChanged is a free log retrieval operation binding the contract event 0xbd79b86ffe0ab8e8776151514217cd7cacd52c909f66475c3af44e129f0b00ff.
//
// Solidity: event RoleAdminChanged(bytes32 indexed role, bytes32 indexed previousAdminRole, bytes32 indexed newAdminRole)
func (_OrgRegistry *OrgRegistryFilterer) FilterRoleAdminChanged(opts *bind.FilterOpts, role [][32]byte, previousAdminRole [][32]byte, newAdminRole [][32]byte) (*OrgRegistryRoleAdminChangedIterator, error) {

	var roleRule []interface{}
	for _, roleItem := range role {
		roleRule = append(roleRule, roleItem)
	}
	var previousAdminRoleRule []interface{}
	for _, previousAdminRoleItem := range previousAdminRole {
		previousAdminRoleRule = append(previousAdminRoleRule, previousAdminRoleItem)
	}
	var newAdminRoleRule []interface{}
	for _, newAdminRoleItem := range newAdminRole {
		newAdminRoleRule = append(newAdminRoleRule, newAdminRoleItem)
	}

	logs, sub, err := _OrgRegistry.contract.FilterLogs(opts, "RoleAdminChanged", roleRule, previousAdminRoleRule, newAdminRoleRule)
	if err != nil {
		return nil, err
	}
	return &OrgRegistryRoleAdminChangedIterator{contract: _OrgRegistry.contract, event: "RoleAdminChanged", logs: logs, sub: sub}, nil
}

// WatchRoleAdminChanged is a free log subscription operation binding the contract event 0xbd79b86ffe0ab8e8776151514217cd7cacd52c909f66475c3af44e129f0b00ff.
//
// Solidity: event RoleAdminChanged(bytes32 indexed role, bytes32 indexed previousAdminRole, bytes32 indexed newAdminRole)
func (_OrgRegistry *OrgRegistryFilterer) WatchRoleAdminChanged(opts *bind.WatchOpts, sink chan<- *OrgRegistryRoleAdminChanged, role [][32]byte, previousAdminRole [][32]byte, newAdminRole [][32]byte) (event.Subscription, error) {

	var roleRule []interface{}
	for _, roleItem := range role {
		roleRule = append(roleRule, roleItem)
	}
	var previousAdminRoleRule []interface{}
	for _, previousAdminRoleItem := range previousAdminRole {
		previousAdminRoleRule = append(previousAdminRoleRule, previousAdminRoleItem)
	}
	var newAdminRoleRule []interface{}
	for _, newAdminRoleItem := range newAdminRole {
		newAdminRoleRule = append(newAdminRoleRule, newAdminRoleItem)
	}

	logs, sub, err := _OrgRegistry.contract.WatchLogs(opts, "RoleAdminChanged", roleRule, previousAdminRoleRule, newAdminRoleRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(OrgRegistryRoleAdminChanged)
				if err := _OrgRegistry.contract.UnpackLog(event, "RoleAdminChanged", log); err != nil {
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

// ParseRoleAdminChanged is a log parse operation binding the contract event 0xbd79b86ffe0ab8e8776151514217cd7cacd52c909f66475c3af44e129f0b00ff.
//
// Solidity: event RoleAdminChanged(bytes32 indexed role, bytes32 indexed previousAdminRole, bytes32 indexed newAdminRole)
func (_OrgRegistry *OrgRegistryFilterer) ParseRoleAdminChanged(log types.Log) (*OrgRegistryRoleAdminChanged, error) {
	event := new(OrgRegistryRoleAdminChanged)
	if err := _OrgRegistry.contract.UnpackLog(event, "RoleAdminChanged", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// OrgRegistryRoleGrantedIterator is returned from FilterRoleGranted and is used to iterate over the raw logs and unpacked data for RoleGranted events raised by the OrgRegistry contract.
type OrgRegistryRoleGrantedIterator struct {
	Event *OrgRegistryRoleGranted // Event containing the contract specifics and raw log

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
func (it *OrgRegistryRoleGrantedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(OrgRegistryRoleGranted)
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
		it.Event = new(OrgRegistryRoleGranted)
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
func (it *OrgRegistryRoleGrantedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *OrgRegistryRoleGrantedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// OrgRegistryRoleGranted represents a RoleGranted event raised by the OrgRegistry contract.
type OrgRegistryRoleGranted struct {
	Role    [32]byte
	Account common.Address
	Sender  common.Address
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterRoleGranted is a free log retrieval operation binding the contract event 0x2f8788117e7eff1d82e926ec794901d17c78024a50270940304540a733656f0d.
//
// Solidity: event RoleGranted(bytes32 indexed role, address indexed account, address indexed sender)
func (_OrgRegistry *OrgRegistryFilterer) FilterRoleGranted(opts *bind.FilterOpts, role [][32]byte, account []common.Address, sender []common.Address) (*OrgRegistryRoleGrantedIterator, error) {

	var roleRule []interface{}
	for _, roleItem := range role {
		roleRule = append(roleRule, roleItem)
	}
	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}
	var senderRule []interface{}
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}

	logs, sub, err := _OrgRegistry.contract.FilterLogs(opts, "RoleGranted", roleRule, accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return &OrgRegistryRoleGrantedIterator{contract: _OrgRegistry.contract, event: "RoleGranted", logs: logs, sub: sub}, nil
}

// WatchRoleGranted is a free log subscription operation binding the contract event 0x2f8788117e7eff1d82e926ec794901d17c78024a50270940304540a733656f0d.
//
// Solidity: event RoleGranted(bytes32 indexed role, address indexed account, address indexed sender)
func (_OrgRegistry *OrgRegistryFilterer) WatchRoleGranted(opts *bind.WatchOpts, sink chan<- *OrgRegistryRoleGranted, role [][32]byte, account []common.Address, sender []common.Address) (event.Subscription, error) {

	var roleRule []interface{}
	for _, roleItem := range role {
		roleRule = append(roleRule, roleItem)
	}
	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}
	var senderRule []interface{}
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}

	logs, sub, err := _OrgRegistry.contract.WatchLogs(opts, "RoleGranted", roleRule, accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(OrgRegistryRoleGranted)
				if err := _OrgRegistry.contract.UnpackLog(event, "RoleGranted", log); err != nil {
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

// ParseRoleGranted is a log parse operation binding the contract event 0x2f8788117e7eff1d82e926ec794901d17c78024a50270940304540a733656f0d.
//
// Solidity: event RoleGranted(bytes32 indexed role, address indexed account, address indexed sender)
func (_OrgRegistry *OrgRegistryFilterer) ParseRoleGranted(log types.Log) (*OrgRegistryRoleGranted, error) {
	event := new(OrgRegistryRoleGranted)
	if err := _OrgRegistry.contract.UnpackLog(event, "RoleGranted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// OrgRegistryRoleRevokedIterator is returned from FilterRoleRevoked and is used to iterate over the raw logs and unpacked data for RoleRevoked events raised by the OrgRegistry contract.
type OrgRegistryRoleRevokedIterator struct {
	Event *OrgRegistryRoleRevoked // Event containing the contract specifics and raw log

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
func (it *OrgRegistryRoleRevokedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(OrgRegistryRoleRevoked)
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
		it.Event = new(OrgRegistryRoleRevoked)
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
func (it *OrgRegistryRoleRevokedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *OrgRegistryRoleRevokedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// OrgRegistryRoleRevoked represents a RoleRevoked event raised by the OrgRegistry contract.
type OrgRegistryRoleRevoked struct {
	Role    [32]byte
	Account common.Address
	Sender  common.Address
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterRoleRevoked is a free log retrieval operation binding the contract event 0xf6391f5c32d9c69d2a47ea670b442974b53935d1edc7fd64eb21e047a839171b.
//
// Solidity: event RoleRevoked(bytes32 indexed role, address indexed account, address indexed sender)
func (_OrgRegistry *OrgRegistryFilterer) FilterRoleRevoked(opts *bind.FilterOpts, role [][32]byte, account []common.Address, sender []common.Address) (*OrgRegistryRoleRevokedIterator, error) {

	var roleRule []interface{}
	for _, roleItem := range role {
		roleRule = append(roleRule, roleItem)
	}
	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}
	var senderRule []interface{}
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}

	logs, sub, err := _OrgRegistry.contract.FilterLogs(opts, "RoleRevoked", roleRule, accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return &OrgRegistryRoleRevokedIterator{contract: _OrgRegistry.contract, event: "RoleRevoked", logs: logs, sub: sub}, nil
}

// WatchRoleRevoked is a free log subscription operation binding the contract event 0xf6391f5c32d9c69d2a47ea670b442974b53935d1edc7fd64eb21e047a839171b.
//
// Solidity: event RoleRevoked(bytes32 indexed role, address indexed account, address indexed sender)
func (_OrgRegistry *OrgRegistryFilterer) WatchRoleRevoked(opts *bind.WatchOpts, sink chan<- *OrgRegistryRoleRevoked, role [][32]byte, account []common.Address, sender []common.Address) (event.Subscription, error) {

	var roleRule []interface{}
	for _, roleItem := range role {
		roleRule = append(roleRule, roleItem)
	}
	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}
	var senderRule []interface{}
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}

	logs, sub, err := _OrgRegistry.contract.WatchLogs(opts, "RoleRevoked", roleRule, accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(OrgRegistryRoleRevoked)
				if err := _OrgRegistry.contract.UnpackLog(event, "RoleRevoked", log); err != nil {
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

// ParseRoleRevoked is a log parse operation binding the contract event 0xf6391f5c32d9c69d2a47ea670b442974b53935d1edc7fd64eb21e047a839171b.
//
// Solidity: event RoleRevoked(bytes32 indexed role, address indexed account, address indexed sender)
func (_OrgRegistry *OrgRegistryFilterer) ParseRoleRevoked(log types.Log) (*OrgRegistryRoleRevoked, error) {
	event := new(OrgRegistryRoleRevoked)
	if err := _OrgRegistry.contract.UnpackLog(event, "RoleRevoked", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
