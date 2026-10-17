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

// CustodyManagerHandover is an auto generated low-level Go binding around an user-defined struct.
type CustodyManagerHandover struct {
	PassportId [32]byte
	FromOrg    [32]byte
	ToOrg      [32]byte
	Quantity   *big.Int
	DataHash   [32]byte
	Nonce      *big.Int
	Deadline   *big.Int
}

// CustodyManagerMetaData contains all meta data concerning the CustodyManager contract.
var CustodyManagerMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"contractOrgRegistry\",\"name\":\"orgs_\",\"type\":\"address\"},{\"internalType\":\"contractPassportRegistry\",\"name\":\"passports_\",\"type\":\"address\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"given\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"expected\",\"type\":\"uint256\"}],\"name\":\"BadNonce\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"orgId\",\"type\":\"bytes32\"}],\"name\":\"BadSignature\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"deadline\",\"type\":\"uint256\"}],\"name\":\"Expired\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"InvalidShortString\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotAdmin\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"passportId\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"fromOrg\",\"type\":\"bytes32\"}],\"name\":\"NotHolder\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotOperator\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotRelayer\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"orgId\",\"type\":\"bytes32\"}],\"name\":\"OrgNotVerified\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"passportId\",\"type\":\"bytes32\"}],\"name\":\"PassportNotActive\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"given\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"held\",\"type\":\"uint256\"}],\"name\":\"QuantityMismatch\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"SameOrg\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"str\",\"type\":\"string\"}],\"name\":\"StringTooLong\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"ZeroValue\",\"type\":\"error\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"passportId\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"fromOrg\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"toOrg\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"quantity\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"dataHash\",\"type\":\"bytes32\"}],\"name\":\"CustodyTransferred\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[],\"name\":\"EIP712DomainChanged\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"passportId\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"holder\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"kind\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"dataHash\",\"type\":\"bytes32\"}],\"name\":\"TraceEvent\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"HANDOVER_TYPEHASH\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"eip712Domain\",\"outputs\":[{\"internalType\":\"bytes1\",\"name\":\"fields\",\"type\":\"bytes1\"},{\"internalType\":\"string\",\"name\":\"name\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"version\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"chainId\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"verifyingContract\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"salt\",\"type\":\"bytes32\"},{\"internalType\":\"uint256[]\",\"name\":\"extensions\",\"type\":\"uint256[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"passportId\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"fromOrg\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"toOrg\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"quantity\",\"type\":\"uint256\"},{\"internalType\":\"bytes32\",\"name\":\"dataHash\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"deadline\",\"type\":\"uint256\"}],\"internalType\":\"structCustodyManager.Handover\",\"name\":\"h\",\"type\":\"tuple\"}],\"name\":\"hashHandover\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"passportId\",\"type\":\"bytes32\"}],\"name\":\"nonces\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"orgs\",\"outputs\":[{\"internalType\":\"contractOrgRegistry\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"passports\",\"outputs\":[{\"internalType\":\"contractPassportRegistry\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"passportId\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"kind\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"dataHash\",\"type\":\"bytes32\"}],\"name\":\"recordEvent\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"passportId\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"fromOrg\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"toOrg\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"quantity\",\"type\":\"uint256\"},{\"internalType\":\"bytes32\",\"name\":\"dataHash\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"deadline\",\"type\":\"uint256\"}],\"internalType\":\"structCustodyManager.Handover\",\"name\":\"h\",\"type\":\"tuple\"},{\"internalType\":\"bytes\",\"name\":\"sigFrom\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"sigTo\",\"type\":\"bytes\"}],\"name\":\"recordHandover\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
	Bin: "0x6101a0604052348015610010575f5ffd5b5060405161176438038061176483398101604081905261002f916101c8565b604080518082018252600a815269436861696e547261636b60b01b602080830191909152825180840190935260018352603160f81b908301526001600160a01b03841660805290610080825f610139565b6101405261008f816001610139565b61016052815160208084019190912061010052815190820120610120524660c05261011d6101005161012051604080517f8b73c3c69bb8fe3d512ecc4cf759cc79239f7b179b0ffacaa9a75d522b39400f60208201529081019290925260608201524660808201523060a08201525f9060c00160405160208183030381529060405280519060200120905090565b60a05250503060e0526001600160a01b031661018052506103b9565b5f6020835110156101545761014d8361016b565b9050610165565b8161015f84826102a3565b5060ff90505b92915050565b5f5f829050601f8151111561019e578260405163305a27a960e01b81526004016101959190610361565b60405180910390fd5b80516101a982610396565b179392505050565b6001600160a01b03811681146101c5575f5ffd5b50565b5f5f604083850312156101d9575f5ffd5b82516101e4816101b1565b60208401519092506101f5816101b1565b809150509250929050565b634e487b7160e01b5f52604160045260245ffd5b600181811c9082168061022857607f821691505b60208210810361024657634e487b7160e01b5f52602260045260245ffd5b50919050565b601f82111561029e578282111561029e57805f5260205f20601f840160051c602085101561027757505f5b90810190601f840160051c035f5b8181101561029a575f83820155600101610285565b5050505b505050565b81516001600160401b038111156102bc576102bc610200565b6102d0816102ca8454610214565b8461024c565b6020601f821160018114610302575f83156102eb5750848201515b5f19600385901b1c1916600184901b17845561035a565b5f84815260208120601f198516915b828110156103315787850151825560209485019460019092019101610311565b508482101561034e57868401515f19600387901b60f8161c191681555b505060018360011b0184555b5050505050565b602081525f82518060208401528060208501604085015e5f604082850101526040601f19601f83011684010191505092915050565b80516020808301519190811015610246575f1960209190910360031b1b16919050565b60805160a05160c05160e05161010051610120516101405161016051610180516113016104635f395f8181608e0152818161030e015281816106f4015261099901525f610abe01525f610a8d01525f610cfd01525f610cd501525f610c3001525f610c5a01525f610c8401525f818161017c015281816101a0015281816101cf015281816104f901528181610528015281816107ed0152818161089a0152610afc01526113015ff3fe608060405234801561000f575f5ffd5b5060043610610085575f3560e01c806384b0196e1161005857806384b0196e1461012a5780639e317f1214610145578063b2ffbcbb14610164578063cab19cdc14610177575f5ffd5b8063174ca1ce14610089578063363fc53f146100cd57806339498ed8146100e257806372fec2d414610103575b5f5ffd5b6100b07f000000000000000000000000000000000000000000000000000000000000000081565b6040516001600160a01b0390911681526020015b60405180910390f35b6100e06100db366004610f6c565b61019e565b005b6100f56100f0366004610fab565b610405565b6040519081526020016100c4565b6100f57f6e9943df859533104c6022374821fb54b8718bf31a2409762cd50729f4f495fe81565b6101326104b5565b6040516100c49796959493929190610ffa565b6100f5610153366004611090565b60026020525f908152604090205481565b6100e06101723660046110ec565b6104f7565b6100b07f000000000000000000000000000000000000000000000000000000000000000081565b7f00000000000000000000000000000000000000000000000000000000000000006001600160a01b03166391d148547f00000000000000000000000000000000000000000000000000000000000000006001600160a01b031663926d7d7f6040518163ffffffff1660e01b8152600401602060405180830381865afa158015610229573d5f5f3e3d5ffd5b505050506040513d601f19601f8201168201806040525081019061024d919061116f565b6040516001600160e01b031960e084901b1681526004810191909152336024820152604401602060405180830381865afa15801561028d573d5f5f3e3d5ffd5b505050506040513d601f19601f820116820180604052508101906102b19190611186565b6102ce5760405163c64891a560e01b815260040160405180910390fd5b8115806102d9575080155b156102f757604051637c946ed760e01b815260040160405180910390fd5b6040516248cd7d60e51b8152600481018490525f907f00000000000000000000000000000000000000000000000000000000000000006001600160a01b031690630919afa09060240160c060405180830381865afa15801561035b573d5f5f3e3d5ffd5b505050506040513d601f19601f8201168201806040525081019061037f91906111b8565b905060018160a00151600481111561039957610399611240565b146103bf57604051631e9b4ec360e21b8152600481018590526024015b60405180910390fd5b828160200151857f362d6101ab1b9dcc9234188a2a9f21fc27563b9604906263bedfb04006c4da51856040516103f791815260200190565b60405180910390a450505050565b5f6104af7f6e9943df859533104c6022374821fb54b8718bf31a2409762cd50729f4f495fe835f013584602001358560400135866060013587608001358860a001358960c00135604051602001610494989796959493929190978852602088019690965260408701949094526060860192909252608085015260a084015260c083015260e08201526101000190565b60405160208183030381529060405280519060200120610a5a565b92915050565b5f6060805f5f5f60606104c6610a86565b6104ce610ab7565b604080515f80825260208201909252600f60f81b9b939a50919850469750309650945092509050565b7f00000000000000000000000000000000000000000000000000000000000000006001600160a01b03166391d148547f00000000000000000000000000000000000000000000000000000000000000006001600160a01b031663926d7d7f6040518163ffffffff1660e01b8152600401602060405180830381865afa158015610582573d5f5f3e3d5ffd5b505050506040513d601f19601f820116820180604052508101906105a6919061116f565b6040516001600160e01b031960e084901b1681526004810191909152336024820152604401602060405180830381865afa1580156105e6573d5f5f3e3d5ffd5b505050506040513d601f19601f8201168201806040525081019061060a9190611186565b6106275760405163c64891a560e01b815260040160405180910390fd5b8460c0013542111561065257604051637c06dd7560e11b815260c086013560048201526024016103b6565b84355f9081526002602052604090205460a0860135811461069357604051633ddd513160e01b815260a08701356004820152602481018290526044016103b6565b85604001358660200135036106bb576040516398532b0960e01b815260040160405180910390fd5b60808601356106dd57604051637c946ed760e01b815260040160405180910390fd5b6040516248cd7d60e51b8152863560048201525f907f00000000000000000000000000000000000000000000000000000000000000006001600160a01b031690630919afa09060240160c060405180830381865afa158015610741573d5f5f3e3d5ffd5b505050506040513d601f19601f8201168201806040525081019061076591906111b8565b9050866020013581602001511461079c5760405163a992598960e01b815287356004820152602088013560248201526044016103b6565b80606001518760600135146107d457606081810151604051631fbe047160e31b815291890135600483015260248201526044016103b6565b60405163c181b27360e01b8152602088013560048201527f00000000000000000000000000000000000000000000000000000000000000006001600160a01b03169063c181b27390602401602060405180830381865afa15801561083a573d5f5f3e3d5ffd5b505050506040513d601f19601f8201168201806040525081019061085e9190611186565b61088157604051636bed6a6760e01b8152602088013560048201526024016103b6565b6040805163c181b27360e01b81529088013560048201527f00000000000000000000000000000000000000000000000000000000000000006001600160a01b03169063c181b27390602401602060405180830381865afa1580156108e7573d5f5f3e3d5ffd5b505050506040513d601f19601f8201168201806040525081019061090b9190611186565b61092e5760408051636bed6a6760e01b81529088013560048201526024016103b6565b5f61093888610405565b905061094a8860200135828989610ae4565b61095a8860400135828787610ae4565b610965836001611254565b88355f818152600260205260409081902092909255815163fa2deb7f60e01b815260048101919091529089013560248201527f00000000000000000000000000000000000000000000000000000000000000006001600160a01b03169063fa2deb7f906044015f604051808303815f87803b1580156109e2575f5ffd5b505af11580156109f4573d5f5f3e3d5ffd5b5050505087604001358860200135895f01357f8bdfb6aed70d24fb4fd5d461facc3a41886ddfa5c8350184df857abb2dfeb9b68b606001358c60800135604051610a48929190918252602082015260400190565b60405180910390a45050505050505050565b5f6104af610a66610c24565b8360405161190160f01b8152600281019290925260228201526042902090565b6060610ab27f00000000000000000000000000000000000000000000000000000000000000005f610d4d565b905090565b6060610ab27f00000000000000000000000000000000000000000000000000000000000000006001610d4d565b604051634c11f34360e11b8152600481018590525f907f00000000000000000000000000000000000000000000000000000000000000006001600160a01b031690639823e68690602401602060405180830381865afa158015610b49573d5f5f3e3d5ffd5b505050506040513d601f19601f82011682018060405250810190610b6d9190611273565b90505f5f610bb08686868080601f0160208091040260200160405190810160405280939291908181526020018383808284375f92019190915250610df792505050565b5090925090505f816003811115610bc957610bc9611240565b141580610bdd57506001600160a01b038316155b80610bfa5750826001600160a01b0316826001600160a01b031614155b15610c1b5760405163252d5ca160e01b8152600481018890526024016103b6565b50505050505050565b5f306001600160a01b037f000000000000000000000000000000000000000000000000000000000000000016148015610c7c57507f000000000000000000000000000000000000000000000000000000000000000046145b15610ca657507f000000000000000000000000000000000000000000000000000000000000000090565b610ab2604080517f8b73c3c69bb8fe3d512ecc4cf759cc79239f7b179b0ffacaa9a75d522b39400f60208201527f0000000000000000000000000000000000000000000000000000000000000000918101919091527f000000000000000000000000000000000000000000000000000000000000000060608201524660808201523060a08201525f9060c00160405160208183030381529060405280519060200120905090565b606060ff8314610d6757610d6083610e40565b90506104af565b818054610d7390611299565b80601f0160208091040260200160405190810160405280929190818152602001828054610d9f90611299565b8015610dea5780601f10610dc157610100808354040283529160200191610dea565b820191905f5260205f20905b815481529060010190602001808311610dcd57829003601f168201915b5050505050905092915050565b5f5f5f8351604103610e2e576020840151604085015160608601515f1a610e2088828585610e7d565b955095509550505050610e39565b505081515f91506002905b9250925092565b60605f610e4c83610f45565b6040805160208082528183019092529192505f91906020820181803683375050509182525060208101929092525090565b5f80807f7fffffffffffffffffffffffffffffff5d576e7357a4501ddfe92f46681b20a0841115610eb657505f91506003905082610f3b565b604080515f808252602082018084528a905260ff891692820192909252606081018790526080810186905260019060a0016020604051602081039080840390855afa158015610f07573d5f5f3e3d5ffd5b5050604051601f1901519150506001600160a01b038116610f3257505f925060019150829050610f3b565b92505f91508190505b9450945094915050565b5f60ff8216601f8111156104af57604051632cd44ac360e21b815260040160405180910390fd5b5f5f5f60608486031215610f7e575f5ffd5b505081359360208301359350604090920135919050565b5f60e08284031215610fa5575f5ffd5b50919050565b5f60e08284031215610fbb575f5ffd5b610fc58383610f95565b9392505050565b5f81518084528060208401602086015e5f602082860101526020601f19601f83011685010191505092915050565b60ff60f81b8816815260e060208201525f61101860e0830189610fcc565b828103604084015261102a8189610fcc565b606084018890526001600160a01b038716608085015260a0840186905283810360c0850152845180825260208087019350909101905f5b8181101561107f578351835260209384019390920191600101611061565b50909b9a5050505050505050505050565b5f602082840312156110a0575f5ffd5b5035919050565b5f5f83601f8401126110b7575f5ffd5b50813567ffffffffffffffff8111156110ce575f5ffd5b6020830191508360208285010111156110e5575f5ffd5b9250929050565b5f5f5f5f5f6101208688031215611101575f5ffd5b61110b8787610f95565b945060e086013567ffffffffffffffff811115611126575f5ffd5b611132888289016110a7565b90955093505061010086013567ffffffffffffffff811115611152575f5ffd5b61115e888289016110a7565b969995985093965092949392505050565b5f6020828403121561117f575f5ffd5b5051919050565b5f60208284031215611196575f5ffd5b81518015158114610fc5575f5ffd5b8051600581106111b3575f5ffd5b919050565b5f60c08284031280156111c9575f5ffd5b5060405160c0810167ffffffffffffffff811182821017156111f957634e487b7160e01b5f52604160045260245ffd5b604090815283518252602080850151908301528381015190820152606080840151908201526080808401519082015261123460a084016111a5565b60a08201529392505050565b634e487b7160e01b5f52602160045260245ffd5b808201808211156104af57634e487b7160e01b5f52601160045260245ffd5b5f60208284031215611283575f5ffd5b81516001600160a01b0381168114610fc5575f5ffd5b600181811c908216806112ad57607f821691505b602082108103610fa557634e487b7160e01b5f52602260045260245ffdfea26469706673582212207044299145d955778d693f78e4d09bb8c05c1338fb7499f9d9a886c6c93cde8464736f6c63430008220033",
}

// CustodyManagerABI is the input ABI used to generate the binding from.
// Deprecated: Use CustodyManagerMetaData.ABI instead.
var CustodyManagerABI = CustodyManagerMetaData.ABI

// CustodyManagerBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use CustodyManagerMetaData.Bin instead.
var CustodyManagerBin = CustodyManagerMetaData.Bin

// DeployCustodyManager deploys a new Ethereum contract, binding an instance of CustodyManager to it.
func DeployCustodyManager(auth *bind.TransactOpts, backend bind.ContractBackend, orgs_ common.Address, passports_ common.Address) (common.Address, *types.Transaction, *CustodyManager, error) {
	parsed, err := CustodyManagerMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(CustodyManagerBin), backend, orgs_, passports_)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &CustodyManager{CustodyManagerCaller: CustodyManagerCaller{contract: contract}, CustodyManagerTransactor: CustodyManagerTransactor{contract: contract}, CustodyManagerFilterer: CustodyManagerFilterer{contract: contract}}, nil
}

// CustodyManager is an auto generated Go binding around an Ethereum contract.
type CustodyManager struct {
	CustodyManagerCaller     // Read-only binding to the contract
	CustodyManagerTransactor // Write-only binding to the contract
	CustodyManagerFilterer   // Log filterer for contract events
}

// CustodyManagerCaller is an auto generated read-only Go binding around an Ethereum contract.
type CustodyManagerCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// CustodyManagerTransactor is an auto generated write-only Go binding around an Ethereum contract.
type CustodyManagerTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// CustodyManagerFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type CustodyManagerFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// CustodyManagerSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type CustodyManagerSession struct {
	Contract     *CustodyManager   // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// CustodyManagerCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type CustodyManagerCallerSession struct {
	Contract *CustodyManagerCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts         // Call options to use throughout this session
}

// CustodyManagerTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type CustodyManagerTransactorSession struct {
	Contract     *CustodyManagerTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts         // Transaction auth options to use throughout this session
}

// CustodyManagerRaw is an auto generated low-level Go binding around an Ethereum contract.
type CustodyManagerRaw struct {
	Contract *CustodyManager // Generic contract binding to access the raw methods on
}

// CustodyManagerCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type CustodyManagerCallerRaw struct {
	Contract *CustodyManagerCaller // Generic read-only contract binding to access the raw methods on
}

// CustodyManagerTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type CustodyManagerTransactorRaw struct {
	Contract *CustodyManagerTransactor // Generic write-only contract binding to access the raw methods on
}

// NewCustodyManager creates a new instance of CustodyManager, bound to a specific deployed contract.
func NewCustodyManager(address common.Address, backend bind.ContractBackend) (*CustodyManager, error) {
	contract, err := bindCustodyManager(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &CustodyManager{CustodyManagerCaller: CustodyManagerCaller{contract: contract}, CustodyManagerTransactor: CustodyManagerTransactor{contract: contract}, CustodyManagerFilterer: CustodyManagerFilterer{contract: contract}}, nil
}

// NewCustodyManagerCaller creates a new read-only instance of CustodyManager, bound to a specific deployed contract.
func NewCustodyManagerCaller(address common.Address, caller bind.ContractCaller) (*CustodyManagerCaller, error) {
	contract, err := bindCustodyManager(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &CustodyManagerCaller{contract: contract}, nil
}

// NewCustodyManagerTransactor creates a new write-only instance of CustodyManager, bound to a specific deployed contract.
func NewCustodyManagerTransactor(address common.Address, transactor bind.ContractTransactor) (*CustodyManagerTransactor, error) {
	contract, err := bindCustodyManager(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &CustodyManagerTransactor{contract: contract}, nil
}

// NewCustodyManagerFilterer creates a new log filterer instance of CustodyManager, bound to a specific deployed contract.
func NewCustodyManagerFilterer(address common.Address, filterer bind.ContractFilterer) (*CustodyManagerFilterer, error) {
	contract, err := bindCustodyManager(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &CustodyManagerFilterer{contract: contract}, nil
}

// bindCustodyManager binds a generic wrapper to an already deployed contract.
func bindCustodyManager(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := CustodyManagerMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_CustodyManager *CustodyManagerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _CustodyManager.Contract.CustodyManagerCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_CustodyManager *CustodyManagerRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _CustodyManager.Contract.CustodyManagerTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_CustodyManager *CustodyManagerRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _CustodyManager.Contract.CustodyManagerTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_CustodyManager *CustodyManagerCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _CustodyManager.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_CustodyManager *CustodyManagerTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _CustodyManager.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_CustodyManager *CustodyManagerTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _CustodyManager.Contract.contract.Transact(opts, method, params...)
}

// HANDOVERTYPEHASH is a free data retrieval call binding the contract method 0x72fec2d4.
//
// Solidity: function HANDOVER_TYPEHASH() view returns(bytes32)
func (_CustodyManager *CustodyManagerCaller) HANDOVERTYPEHASH(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _CustodyManager.contract.Call(opts, &out, "HANDOVER_TYPEHASH")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// HANDOVERTYPEHASH is a free data retrieval call binding the contract method 0x72fec2d4.
//
// Solidity: function HANDOVER_TYPEHASH() view returns(bytes32)
func (_CustodyManager *CustodyManagerSession) HANDOVERTYPEHASH() ([32]byte, error) {
	return _CustodyManager.Contract.HANDOVERTYPEHASH(&_CustodyManager.CallOpts)
}

// HANDOVERTYPEHASH is a free data retrieval call binding the contract method 0x72fec2d4.
//
// Solidity: function HANDOVER_TYPEHASH() view returns(bytes32)
func (_CustodyManager *CustodyManagerCallerSession) HANDOVERTYPEHASH() ([32]byte, error) {
	return _CustodyManager.Contract.HANDOVERTYPEHASH(&_CustodyManager.CallOpts)
}

// Eip712Domain is a free data retrieval call binding the contract method 0x84b0196e.
//
// Solidity: function eip712Domain() view returns(bytes1 fields, string name, string version, uint256 chainId, address verifyingContract, bytes32 salt, uint256[] extensions)
func (_CustodyManager *CustodyManagerCaller) Eip712Domain(opts *bind.CallOpts) (struct {
	Fields            [1]byte
	Name              string
	Version           string
	ChainId           *big.Int
	VerifyingContract common.Address
	Salt              [32]byte
	Extensions        []*big.Int
}, error) {
	var out []interface{}
	err := _CustodyManager.contract.Call(opts, &out, "eip712Domain")

	outstruct := new(struct {
		Fields            [1]byte
		Name              string
		Version           string
		ChainId           *big.Int
		VerifyingContract common.Address
		Salt              [32]byte
		Extensions        []*big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Fields = *abi.ConvertType(out[0], new([1]byte)).(*[1]byte)
	outstruct.Name = *abi.ConvertType(out[1], new(string)).(*string)
	outstruct.Version = *abi.ConvertType(out[2], new(string)).(*string)
	outstruct.ChainId = *abi.ConvertType(out[3], new(*big.Int)).(**big.Int)
	outstruct.VerifyingContract = *abi.ConvertType(out[4], new(common.Address)).(*common.Address)
	outstruct.Salt = *abi.ConvertType(out[5], new([32]byte)).(*[32]byte)
	outstruct.Extensions = *abi.ConvertType(out[6], new([]*big.Int)).(*[]*big.Int)

	return *outstruct, err

}

// Eip712Domain is a free data retrieval call binding the contract method 0x84b0196e.
//
// Solidity: function eip712Domain() view returns(bytes1 fields, string name, string version, uint256 chainId, address verifyingContract, bytes32 salt, uint256[] extensions)
func (_CustodyManager *CustodyManagerSession) Eip712Domain() (struct {
	Fields            [1]byte
	Name              string
	Version           string
	ChainId           *big.Int
	VerifyingContract common.Address
	Salt              [32]byte
	Extensions        []*big.Int
}, error) {
	return _CustodyManager.Contract.Eip712Domain(&_CustodyManager.CallOpts)
}

// Eip712Domain is a free data retrieval call binding the contract method 0x84b0196e.
//
// Solidity: function eip712Domain() view returns(bytes1 fields, string name, string version, uint256 chainId, address verifyingContract, bytes32 salt, uint256[] extensions)
func (_CustodyManager *CustodyManagerCallerSession) Eip712Domain() (struct {
	Fields            [1]byte
	Name              string
	Version           string
	ChainId           *big.Int
	VerifyingContract common.Address
	Salt              [32]byte
	Extensions        []*big.Int
}, error) {
	return _CustodyManager.Contract.Eip712Domain(&_CustodyManager.CallOpts)
}

// HashHandover is a free data retrieval call binding the contract method 0x39498ed8.
//
// Solidity: function hashHandover((bytes32,bytes32,bytes32,uint256,bytes32,uint256,uint256) h) view returns(bytes32)
func (_CustodyManager *CustodyManagerCaller) HashHandover(opts *bind.CallOpts, h CustodyManagerHandover) ([32]byte, error) {
	var out []interface{}
	err := _CustodyManager.contract.Call(opts, &out, "hashHandover", h)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// HashHandover is a free data retrieval call binding the contract method 0x39498ed8.
//
// Solidity: function hashHandover((bytes32,bytes32,bytes32,uint256,bytes32,uint256,uint256) h) view returns(bytes32)
func (_CustodyManager *CustodyManagerSession) HashHandover(h CustodyManagerHandover) ([32]byte, error) {
	return _CustodyManager.Contract.HashHandover(&_CustodyManager.CallOpts, h)
}

// HashHandover is a free data retrieval call binding the contract method 0x39498ed8.
//
// Solidity: function hashHandover((bytes32,bytes32,bytes32,uint256,bytes32,uint256,uint256) h) view returns(bytes32)
func (_CustodyManager *CustodyManagerCallerSession) HashHandover(h CustodyManagerHandover) ([32]byte, error) {
	return _CustodyManager.Contract.HashHandover(&_CustodyManager.CallOpts, h)
}

// Nonces is a free data retrieval call binding the contract method 0x9e317f12.
//
// Solidity: function nonces(bytes32 passportId) view returns(uint256)
func (_CustodyManager *CustodyManagerCaller) Nonces(opts *bind.CallOpts, passportId [32]byte) (*big.Int, error) {
	var out []interface{}
	err := _CustodyManager.contract.Call(opts, &out, "nonces", passportId)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Nonces is a free data retrieval call binding the contract method 0x9e317f12.
//
// Solidity: function nonces(bytes32 passportId) view returns(uint256)
func (_CustodyManager *CustodyManagerSession) Nonces(passportId [32]byte) (*big.Int, error) {
	return _CustodyManager.Contract.Nonces(&_CustodyManager.CallOpts, passportId)
}

// Nonces is a free data retrieval call binding the contract method 0x9e317f12.
//
// Solidity: function nonces(bytes32 passportId) view returns(uint256)
func (_CustodyManager *CustodyManagerCallerSession) Nonces(passportId [32]byte) (*big.Int, error) {
	return _CustodyManager.Contract.Nonces(&_CustodyManager.CallOpts, passportId)
}

// Orgs is a free data retrieval call binding the contract method 0xcab19cdc.
//
// Solidity: function orgs() view returns(address)
func (_CustodyManager *CustodyManagerCaller) Orgs(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _CustodyManager.contract.Call(opts, &out, "orgs")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Orgs is a free data retrieval call binding the contract method 0xcab19cdc.
//
// Solidity: function orgs() view returns(address)
func (_CustodyManager *CustodyManagerSession) Orgs() (common.Address, error) {
	return _CustodyManager.Contract.Orgs(&_CustodyManager.CallOpts)
}

// Orgs is a free data retrieval call binding the contract method 0xcab19cdc.
//
// Solidity: function orgs() view returns(address)
func (_CustodyManager *CustodyManagerCallerSession) Orgs() (common.Address, error) {
	return _CustodyManager.Contract.Orgs(&_CustodyManager.CallOpts)
}

// Passports is a free data retrieval call binding the contract method 0x174ca1ce.
//
// Solidity: function passports() view returns(address)
func (_CustodyManager *CustodyManagerCaller) Passports(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _CustodyManager.contract.Call(opts, &out, "passports")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Passports is a free data retrieval call binding the contract method 0x174ca1ce.
//
// Solidity: function passports() view returns(address)
func (_CustodyManager *CustodyManagerSession) Passports() (common.Address, error) {
	return _CustodyManager.Contract.Passports(&_CustodyManager.CallOpts)
}

// Passports is a free data retrieval call binding the contract method 0x174ca1ce.
//
// Solidity: function passports() view returns(address)
func (_CustodyManager *CustodyManagerCallerSession) Passports() (common.Address, error) {
	return _CustodyManager.Contract.Passports(&_CustodyManager.CallOpts)
}

// RecordEvent is a paid mutator transaction binding the contract method 0x363fc53f.
//
// Solidity: function recordEvent(bytes32 passportId, bytes32 kind, bytes32 dataHash) returns()
func (_CustodyManager *CustodyManagerTransactor) RecordEvent(opts *bind.TransactOpts, passportId [32]byte, kind [32]byte, dataHash [32]byte) (*types.Transaction, error) {
	return _CustodyManager.contract.Transact(opts, "recordEvent", passportId, kind, dataHash)
}

// RecordEvent is a paid mutator transaction binding the contract method 0x363fc53f.
//
// Solidity: function recordEvent(bytes32 passportId, bytes32 kind, bytes32 dataHash) returns()
func (_CustodyManager *CustodyManagerSession) RecordEvent(passportId [32]byte, kind [32]byte, dataHash [32]byte) (*types.Transaction, error) {
	return _CustodyManager.Contract.RecordEvent(&_CustodyManager.TransactOpts, passportId, kind, dataHash)
}

// RecordEvent is a paid mutator transaction binding the contract method 0x363fc53f.
//
// Solidity: function recordEvent(bytes32 passportId, bytes32 kind, bytes32 dataHash) returns()
func (_CustodyManager *CustodyManagerTransactorSession) RecordEvent(passportId [32]byte, kind [32]byte, dataHash [32]byte) (*types.Transaction, error) {
	return _CustodyManager.Contract.RecordEvent(&_CustodyManager.TransactOpts, passportId, kind, dataHash)
}

// RecordHandover is a paid mutator transaction binding the contract method 0xb2ffbcbb.
//
// Solidity: function recordHandover((bytes32,bytes32,bytes32,uint256,bytes32,uint256,uint256) h, bytes sigFrom, bytes sigTo) returns()
func (_CustodyManager *CustodyManagerTransactor) RecordHandover(opts *bind.TransactOpts, h CustodyManagerHandover, sigFrom []byte, sigTo []byte) (*types.Transaction, error) {
	return _CustodyManager.contract.Transact(opts, "recordHandover", h, sigFrom, sigTo)
}

// RecordHandover is a paid mutator transaction binding the contract method 0xb2ffbcbb.
//
// Solidity: function recordHandover((bytes32,bytes32,bytes32,uint256,bytes32,uint256,uint256) h, bytes sigFrom, bytes sigTo) returns()
func (_CustodyManager *CustodyManagerSession) RecordHandover(h CustodyManagerHandover, sigFrom []byte, sigTo []byte) (*types.Transaction, error) {
	return _CustodyManager.Contract.RecordHandover(&_CustodyManager.TransactOpts, h, sigFrom, sigTo)
}

// RecordHandover is a paid mutator transaction binding the contract method 0xb2ffbcbb.
//
// Solidity: function recordHandover((bytes32,bytes32,bytes32,uint256,bytes32,uint256,uint256) h, bytes sigFrom, bytes sigTo) returns()
func (_CustodyManager *CustodyManagerTransactorSession) RecordHandover(h CustodyManagerHandover, sigFrom []byte, sigTo []byte) (*types.Transaction, error) {
	return _CustodyManager.Contract.RecordHandover(&_CustodyManager.TransactOpts, h, sigFrom, sigTo)
}

// CustodyManagerCustodyTransferredIterator is returned from FilterCustodyTransferred and is used to iterate over the raw logs and unpacked data for CustodyTransferred events raised by the CustodyManager contract.
type CustodyManagerCustodyTransferredIterator struct {
	Event *CustodyManagerCustodyTransferred // Event containing the contract specifics and raw log

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
func (it *CustodyManagerCustodyTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(CustodyManagerCustodyTransferred)
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
		it.Event = new(CustodyManagerCustodyTransferred)
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
func (it *CustodyManagerCustodyTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *CustodyManagerCustodyTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// CustodyManagerCustodyTransferred represents a CustodyTransferred event raised by the CustodyManager contract.
type CustodyManagerCustodyTransferred struct {
	PassportId [32]byte
	FromOrg    [32]byte
	ToOrg      [32]byte
	Quantity   *big.Int
	DataHash   [32]byte
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterCustodyTransferred is a free log retrieval operation binding the contract event 0x8bdfb6aed70d24fb4fd5d461facc3a41886ddfa5c8350184df857abb2dfeb9b6.
//
// Solidity: event CustodyTransferred(bytes32 indexed passportId, bytes32 indexed fromOrg, bytes32 indexed toOrg, uint256 quantity, bytes32 dataHash)
func (_CustodyManager *CustodyManagerFilterer) FilterCustodyTransferred(opts *bind.FilterOpts, passportId [][32]byte, fromOrg [][32]byte, toOrg [][32]byte) (*CustodyManagerCustodyTransferredIterator, error) {

	var passportIdRule []interface{}
	for _, passportIdItem := range passportId {
		passportIdRule = append(passportIdRule, passportIdItem)
	}
	var fromOrgRule []interface{}
	for _, fromOrgItem := range fromOrg {
		fromOrgRule = append(fromOrgRule, fromOrgItem)
	}
	var toOrgRule []interface{}
	for _, toOrgItem := range toOrg {
		toOrgRule = append(toOrgRule, toOrgItem)
	}

	logs, sub, err := _CustodyManager.contract.FilterLogs(opts, "CustodyTransferred", passportIdRule, fromOrgRule, toOrgRule)
	if err != nil {
		return nil, err
	}
	return &CustodyManagerCustodyTransferredIterator{contract: _CustodyManager.contract, event: "CustodyTransferred", logs: logs, sub: sub}, nil
}

// WatchCustodyTransferred is a free log subscription operation binding the contract event 0x8bdfb6aed70d24fb4fd5d461facc3a41886ddfa5c8350184df857abb2dfeb9b6.
//
// Solidity: event CustodyTransferred(bytes32 indexed passportId, bytes32 indexed fromOrg, bytes32 indexed toOrg, uint256 quantity, bytes32 dataHash)
func (_CustodyManager *CustodyManagerFilterer) WatchCustodyTransferred(opts *bind.WatchOpts, sink chan<- *CustodyManagerCustodyTransferred, passportId [][32]byte, fromOrg [][32]byte, toOrg [][32]byte) (event.Subscription, error) {

	var passportIdRule []interface{}
	for _, passportIdItem := range passportId {
		passportIdRule = append(passportIdRule, passportIdItem)
	}
	var fromOrgRule []interface{}
	for _, fromOrgItem := range fromOrg {
		fromOrgRule = append(fromOrgRule, fromOrgItem)
	}
	var toOrgRule []interface{}
	for _, toOrgItem := range toOrg {
		toOrgRule = append(toOrgRule, toOrgItem)
	}

	logs, sub, err := _CustodyManager.contract.WatchLogs(opts, "CustodyTransferred", passportIdRule, fromOrgRule, toOrgRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(CustodyManagerCustodyTransferred)
				if err := _CustodyManager.contract.UnpackLog(event, "CustodyTransferred", log); err != nil {
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

// ParseCustodyTransferred is a log parse operation binding the contract event 0x8bdfb6aed70d24fb4fd5d461facc3a41886ddfa5c8350184df857abb2dfeb9b6.
//
// Solidity: event CustodyTransferred(bytes32 indexed passportId, bytes32 indexed fromOrg, bytes32 indexed toOrg, uint256 quantity, bytes32 dataHash)
func (_CustodyManager *CustodyManagerFilterer) ParseCustodyTransferred(log types.Log) (*CustodyManagerCustodyTransferred, error) {
	event := new(CustodyManagerCustodyTransferred)
	if err := _CustodyManager.contract.UnpackLog(event, "CustodyTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// CustodyManagerEIP712DomainChangedIterator is returned from FilterEIP712DomainChanged and is used to iterate over the raw logs and unpacked data for EIP712DomainChanged events raised by the CustodyManager contract.
type CustodyManagerEIP712DomainChangedIterator struct {
	Event *CustodyManagerEIP712DomainChanged // Event containing the contract specifics and raw log

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
func (it *CustodyManagerEIP712DomainChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(CustodyManagerEIP712DomainChanged)
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
		it.Event = new(CustodyManagerEIP712DomainChanged)
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
func (it *CustodyManagerEIP712DomainChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *CustodyManagerEIP712DomainChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// CustodyManagerEIP712DomainChanged represents a EIP712DomainChanged event raised by the CustodyManager contract.
type CustodyManagerEIP712DomainChanged struct {
	Raw types.Log // Blockchain specific contextual infos
}

// FilterEIP712DomainChanged is a free log retrieval operation binding the contract event 0x0a6387c9ea3628b88a633bb4f3b151770f70085117a15f9bf3787cda53f13d31.
//
// Solidity: event EIP712DomainChanged()
func (_CustodyManager *CustodyManagerFilterer) FilterEIP712DomainChanged(opts *bind.FilterOpts) (*CustodyManagerEIP712DomainChangedIterator, error) {

	logs, sub, err := _CustodyManager.contract.FilterLogs(opts, "EIP712DomainChanged")
	if err != nil {
		return nil, err
	}
	return &CustodyManagerEIP712DomainChangedIterator{contract: _CustodyManager.contract, event: "EIP712DomainChanged", logs: logs, sub: sub}, nil
}

// WatchEIP712DomainChanged is a free log subscription operation binding the contract event 0x0a6387c9ea3628b88a633bb4f3b151770f70085117a15f9bf3787cda53f13d31.
//
// Solidity: event EIP712DomainChanged()
func (_CustodyManager *CustodyManagerFilterer) WatchEIP712DomainChanged(opts *bind.WatchOpts, sink chan<- *CustodyManagerEIP712DomainChanged) (event.Subscription, error) {

	logs, sub, err := _CustodyManager.contract.WatchLogs(opts, "EIP712DomainChanged")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(CustodyManagerEIP712DomainChanged)
				if err := _CustodyManager.contract.UnpackLog(event, "EIP712DomainChanged", log); err != nil {
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

// ParseEIP712DomainChanged is a log parse operation binding the contract event 0x0a6387c9ea3628b88a633bb4f3b151770f70085117a15f9bf3787cda53f13d31.
//
// Solidity: event EIP712DomainChanged()
func (_CustodyManager *CustodyManagerFilterer) ParseEIP712DomainChanged(log types.Log) (*CustodyManagerEIP712DomainChanged, error) {
	event := new(CustodyManagerEIP712DomainChanged)
	if err := _CustodyManager.contract.UnpackLog(event, "EIP712DomainChanged", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// CustodyManagerTraceEventIterator is returned from FilterTraceEvent and is used to iterate over the raw logs and unpacked data for TraceEvent events raised by the CustodyManager contract.
type CustodyManagerTraceEventIterator struct {
	Event *CustodyManagerTraceEvent // Event containing the contract specifics and raw log

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
func (it *CustodyManagerTraceEventIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(CustodyManagerTraceEvent)
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
		it.Event = new(CustodyManagerTraceEvent)
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
func (it *CustodyManagerTraceEventIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *CustodyManagerTraceEventIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// CustodyManagerTraceEvent represents a TraceEvent event raised by the CustodyManager contract.
type CustodyManagerTraceEvent struct {
	PassportId [32]byte
	Holder     [32]byte
	Kind       [32]byte
	DataHash   [32]byte
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterTraceEvent is a free log retrieval operation binding the contract event 0x362d6101ab1b9dcc9234188a2a9f21fc27563b9604906263bedfb04006c4da51.
//
// Solidity: event TraceEvent(bytes32 indexed passportId, bytes32 indexed holder, bytes32 indexed kind, bytes32 dataHash)
func (_CustodyManager *CustodyManagerFilterer) FilterTraceEvent(opts *bind.FilterOpts, passportId [][32]byte, holder [][32]byte, kind [][32]byte) (*CustodyManagerTraceEventIterator, error) {

	var passportIdRule []interface{}
	for _, passportIdItem := range passportId {
		passportIdRule = append(passportIdRule, passportIdItem)
	}
	var holderRule []interface{}
	for _, holderItem := range holder {
		holderRule = append(holderRule, holderItem)
	}
	var kindRule []interface{}
	for _, kindItem := range kind {
		kindRule = append(kindRule, kindItem)
	}

	logs, sub, err := _CustodyManager.contract.FilterLogs(opts, "TraceEvent", passportIdRule, holderRule, kindRule)
	if err != nil {
		return nil, err
	}
	return &CustodyManagerTraceEventIterator{contract: _CustodyManager.contract, event: "TraceEvent", logs: logs, sub: sub}, nil
}

// WatchTraceEvent is a free log subscription operation binding the contract event 0x362d6101ab1b9dcc9234188a2a9f21fc27563b9604906263bedfb04006c4da51.
//
// Solidity: event TraceEvent(bytes32 indexed passportId, bytes32 indexed holder, bytes32 indexed kind, bytes32 dataHash)
func (_CustodyManager *CustodyManagerFilterer) WatchTraceEvent(opts *bind.WatchOpts, sink chan<- *CustodyManagerTraceEvent, passportId [][32]byte, holder [][32]byte, kind [][32]byte) (event.Subscription, error) {

	var passportIdRule []interface{}
	for _, passportIdItem := range passportId {
		passportIdRule = append(passportIdRule, passportIdItem)
	}
	var holderRule []interface{}
	for _, holderItem := range holder {
		holderRule = append(holderRule, holderItem)
	}
	var kindRule []interface{}
	for _, kindItem := range kind {
		kindRule = append(kindRule, kindItem)
	}

	logs, sub, err := _CustodyManager.contract.WatchLogs(opts, "TraceEvent", passportIdRule, holderRule, kindRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(CustodyManagerTraceEvent)
				if err := _CustodyManager.contract.UnpackLog(event, "TraceEvent", log); err != nil {
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

// ParseTraceEvent is a log parse operation binding the contract event 0x362d6101ab1b9dcc9234188a2a9f21fc27563b9604906263bedfb04006c4da51.
//
// Solidity: event TraceEvent(bytes32 indexed passportId, bytes32 indexed holder, bytes32 indexed kind, bytes32 dataHash)
func (_CustodyManager *CustodyManagerFilterer) ParseTraceEvent(log types.Log) (*CustodyManagerTraceEvent, error) {
	event := new(CustodyManagerTraceEvent)
	if err := _CustodyManager.contract.UnpackLog(event, "TraceEvent", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
