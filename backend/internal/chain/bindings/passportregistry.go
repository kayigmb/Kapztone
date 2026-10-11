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

// PassportRegistryOutput is an auto generated low-level Go binding around an user-defined struct.
type PassportRegistryOutput struct {
	Id       [32]byte
	Quantity *big.Int
	DataHash [32]byte
}

// PassportRegistryPassport is an auto generated low-level Go binding around an user-defined struct.
type PassportRegistryPassport struct {
	Commodity [32]byte
	Holder    [32]byte
	Quantity  *big.Int
	Remaining *big.Int
	DataHash  [32]byte
	Status    uint8
}

// PassportRegistryMetaData contains all meta data concerning the PassportRegistry contract.
var PassportRegistryMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"contractOrgRegistry\",\"name\":\"orgs_\",\"type\":\"address\"},{\"internalType\":\"contractFacilityRegistry\",\"name\":\"facilities_\",\"type\":\"address\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"passportId\",\"type\":\"bytes32\"}],\"name\":\"AlreadyRevoked\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"CustodianAlreadySet\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"passportId\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"requested\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"remaining\",\"type\":\"uint256\"}],\"name\":\"InsufficientQuantity\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"LengthMismatch\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"outputs\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"maxOutputs\",\"type\":\"uint256\"}],\"name\":\"MassBalanceExceeded\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"passportId\",\"type\":\"bytes32\"}],\"name\":\"MixedInputs\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotAdmin\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotCustodian\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotRelayer\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"passportId\",\"type\":\"bytes32\"}],\"name\":\"PassportExists\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"passportId\",\"type\":\"bytes32\"}],\"name\":\"PassportNotActive\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"passportId\",\"type\":\"bytes32\"}],\"name\":\"PassportUnknown\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"commodity\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"stage\",\"type\":\"bytes32\"}],\"name\":\"YieldNotSet\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"ZeroValue\",\"type\":\"error\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"custodian\",\"type\":\"address\"}],\"name\":\"CustodianSet\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"passportId\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"reasonHash\",\"type\":\"bytes32\"}],\"name\":\"Flagged\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"holder\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"stage\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"bytes32[]\",\"name\":\"inputs\",\"type\":\"bytes32[]\"},{\"indexed\":false,\"internalType\":\"uint256[]\",\"name\":\"inputsUsed\",\"type\":\"uint256[]\"},{\"indexed\":false,\"internalType\":\"bytes32[]\",\"name\":\"outputs\",\"type\":\"bytes32[]\"},{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"dataHash\",\"type\":\"bytes32\"}],\"name\":\"LotTransformed\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"commodity\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"stage\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"partsPerMillion\",\"type\":\"uint256\"}],\"name\":\"MaxYieldSet\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"passportId\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"holder\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"facilityId\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"commodity\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"period\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"quantity\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"dataHash\",\"type\":\"bytes32\"}],\"name\":\"PassportIssued\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"passportId\",\"type\":\"bytes32\"}],\"name\":\"PassportRetired\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"passportId\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"reasonHash\",\"type\":\"bytes32\"}],\"name\":\"Revoked\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"YIELD_SCALE\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"custodian\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"facilities\",\"outputs\":[{\"internalType\":\"contractFacilityRegistry\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"passportId\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"reasonHash\",\"type\":\"bytes32\"}],\"name\":\"flag\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"passportId\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"commodity\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"facilityId\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"period\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"quantity\",\"type\":\"uint256\"},{\"internalType\":\"bytes32\",\"name\":\"dataHash\",\"type\":\"bytes32\"}],\"name\":\"issue\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"commodity\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"stage\",\"type\":\"bytes32\"}],\"name\":\"maxYield\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"orgs\",\"outputs\":[{\"internalType\":\"contractOrgRegistry\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"passportId\",\"type\":\"bytes32\"}],\"name\":\"passport\",\"outputs\":[{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"commodity\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"holder\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"quantity\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"remaining\",\"type\":\"uint256\"},{\"internalType\":\"bytes32\",\"name\":\"dataHash\",\"type\":\"bytes32\"},{\"internalType\":\"enumPassportRegistry.Status\",\"name\":\"status\",\"type\":\"uint8\"}],\"internalType\":\"structPassportRegistry.Passport\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"passportId\",\"type\":\"bytes32\"}],\"name\":\"retire\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"passportId\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"reasonHash\",\"type\":\"bytes32\"}],\"name\":\"revoke\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"custodian_\",\"type\":\"address\"}],\"name\":\"setCustodian\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"commodity\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"stage\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"partsPerMillion\",\"type\":\"uint256\"}],\"name\":\"setMaxYield\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"passportId\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"newHolder\",\"type\":\"bytes32\"}],\"name\":\"transferCustody\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32[]\",\"name\":\"inputs\",\"type\":\"bytes32[]\"},{\"internalType\":\"uint256[]\",\"name\":\"inputsUsed\",\"type\":\"uint256[]\"},{\"internalType\":\"bytes32\",\"name\":\"stage\",\"type\":\"bytes32\"},{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"id\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"quantity\",\"type\":\"uint256\"},{\"internalType\":\"bytes32\",\"name\":\"dataHash\",\"type\":\"bytes32\"}],\"internalType\":\"structPassportRegistry.Output[]\",\"name\":\"outputs\",\"type\":\"tuple[]\"},{\"internalType\":\"bytes32\",\"name\":\"dataHash\",\"type\":\"bytes32\"}],\"name\":\"transform\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
	Bin: "0x60c060405234801561000f575f5ffd5b50604051611a97380380611a9783398101604081905261002e9161005c565b6001600160a01b039182166080521660a052610094565b6001600160a01b0381168114610059575f5ffd5b50565b5f5f6040838503121561006d575f5ffd5b825161007881610045565b602084015190925061008981610045565b809150509250929050565b60805160a05161197961011e5f395f818161021e015261046f01525f81816101f7015281816102f2015281816103210152818161054901528181610578015281816106d301528181610702015281816108a6015281816108d501528181610a9b01528181610aca01528181610c5201528181610c8101528181610e180152610e4701526119795ff3fe608060405234801561000f575f5ffd5b50600436106100e5575f3560e01c80638cb52bb611610088578063cab19cdc11610063578063cab19cdc146101f2578063e1050ef914610219578063e6f93e5714610240578063fa2deb7f14610253575f5ffd5b80638cb52bb6146101b95780638db91b46146101cc578063c2664610146101df575f5ffd5b80631f2b6f23116100c35780631f2b6f231461015f578063375b74c314610172578063403f37311461019c5780636267228f146101af575f5ffd5b80630919afa0146100e95780630a90f4f114610112578063179868c11461014a575b5f5ffd5b6100fc6100f736600461155f565b610266565b604051610109919061158a565b60405180910390f35b61013c6101203660046115eb565b600260209081525f928352604080842090915290825290205481565b604051908152602001610109565b61015d61015836600461160b565b6102f0565b005b61015d61016d36600461155f565b610547565b5f54610184906001600160a01b031681565b6040516001600160a01b039091168152602001610109565b61015d6101aa36600461164a565b6106d1565b61013c620f424081565b61015d6101c73660046116bf565b6108a4565b61015d6101da3660046117a1565b610a99565b61015d6101ed3660046115eb565b610c50565b6101847f000000000000000000000000000000000000000000000000000000000000000081565b6101847f000000000000000000000000000000000000000000000000000000000000000081565b61015d61024e3660046115eb565b610e16565b61015d6102613660046115eb565b610f8e565b61026e61151f565b61027782610fca565b6040518060c00160405290815f820154815260200160018201548152602001600282015481526020016003820154815260200160048201548152602001600582015f9054906101000a900460ff1660048111156102d6576102d6611576565b60048111156102e7576102e7611576565b90525092915050565b7f00000000000000000000000000000000000000000000000000000000000000006001600160a01b03166391d148547f00000000000000000000000000000000000000000000000000000000000000006001600160a01b031663926d7d7f6040518163ffffffff1660e01b8152600401602060405180830381865afa15801561037b573d5f5f3e3d5ffd5b505050506040513d601f19601f8201168201806040525081019061039f91906117ca565b6040516001600160e01b031960e084901b1681526004810191909152336024820152604401602060405180830381865afa1580156103df573d5f5f3e3d5ffd5b505050506040513d601f19601f8201168201806040525081019061040391906117e1565b6104205760405163c64891a560e01b815260040160405180910390fd5b84158061042b575081155b1561044957604051637c946ed760e01b815260040160405180910390fd5b6040516331e5d2e560e11b81526004810185905260248101849052604481018390525f907f00000000000000000000000000000000000000000000000000000000000000006001600160a01b0316906363cba5ca906064016020604051808303815f875af11580156104bd573d5f5f3e3d5ffd5b505050506040513d601f19601f820116820180604052508101906104e191906117ca565b90506104f08787838686611017565b6040805187815260208101869052908101849052606081018390528590829089907f8dcb03e40f2cc0d665279dd4894896cf458a980d9b7c67c903692ecbd826f33a9060800160405180910390a450505050505050565b7f00000000000000000000000000000000000000000000000000000000000000006001600160a01b03166391d148547f00000000000000000000000000000000000000000000000000000000000000006001600160a01b031663926d7d7f6040518163ffffffff1660e01b8152600401602060405180830381865afa1580156105d2573d5f5f3e3d5ffd5b505050506040513d601f19601f820116820180604052508101906105f691906117ca565b6040516001600160e01b031960e084901b1681526004810191909152336024820152604401602060405180830381865afa158015610636573d5f5f3e3d5ffd5b505050506040513d601f19601f8201168201806040525081019061065a91906117e1565b6106775760405163c64891a560e01b815260040160405180910390fd5b600361068282611137565b600501805460ff1916600183600481111561069f5761069f611576565b021790555060405181907fde3bd82e297099d837b8a48b5678f7a5ce1c8331c96f54d323394cd8e3f7bbaa905f90a250565b7f00000000000000000000000000000000000000000000000000000000000000006001600160a01b03166391d148547f00000000000000000000000000000000000000000000000000000000000000006001600160a01b031663a217fddf6040518163ffffffff1660e01b8152600401602060405180830381865afa15801561075c573d5f5f3e3d5ffd5b505050506040513d601f19601f8201168201806040525081019061078091906117ca565b6040516001600160e01b031960e084901b1681526004810191909152336024820152604401602060405180830381865afa1580156107c0573d5f5f3e3d5ffd5b505050506040513d601f19601f820116820180604052508101906107e491906117e1565b61080157604051637bfa4b9f60e01b815260040160405180910390fd5b5f546001600160a01b03161561082a5760405163a49cb83d60e01b815260040160405180910390fd5b6001600160a01b03811661085157604051637c946ed760e01b815260040160405180910390fd5b5f80546001600160a01b0319166001600160a01b0383169081179091556040519081527fb88c20a211c5d7677ba2a26c317d8ae6b25aa492016dc8ceca2469761d063d809060200160405180910390a150565b7f00000000000000000000000000000000000000000000000000000000000000006001600160a01b03166391d148547f00000000000000000000000000000000000000000000000000000000000000006001600160a01b031663926d7d7f6040518163ffffffff1660e01b8152600401602060405180830381865afa15801561092f573d5f5f3e3d5ffd5b505050506040513d601f19601f8201168201806040525081019061095391906117ca565b6040516001600160e01b031960e084901b1681526004810191909152336024820152604401602060405180830381865afa158015610993573d5f5f3e3d5ffd5b505050506040513d601f19601f820116820180604052508101906109b791906117e1565b6109d45760405163c64891a560e01b815260040160405180910390fd5b8615806109df575081155b156109fd57604051637c946ed760e01b815260040160405180910390fd5b868514610a20576040516001621398b960e31b0319815260040160405180910390fd5b5f60605f5f610a318c8c8c8c61117f565b91955092509050610a46878784878c86611322565b9250505085827fccc473428e0fc7f33fe2dccee21361ab5de16247ae912e847eb07e98e99c661e8c8c8c8c878a604051610a8596959493929190611830565b60405180910390a350505050505050505050565b7f00000000000000000000000000000000000000000000000000000000000000006001600160a01b03166391d148547f00000000000000000000000000000000000000000000000000000000000000006001600160a01b031663a217fddf6040518163ffffffff1660e01b8152600401602060405180830381865afa158015610b24573d5f5f3e3d5ffd5b505050506040513d601f19601f82011682018060405250810190610b4891906117ca565b6040516001600160e01b031960e084901b1681526004810191909152336024820152604401602060405180830381865afa158015610b88573d5f5f3e3d5ffd5b505050506040513d601f19601f82011682018060405250810190610bac91906117e1565b610bc957604051637bfa4b9f60e01b815260040160405180910390fd5b821580610bd4575081155b80610bdd575080155b15610bfb57604051637c946ed760e01b815260040160405180910390fd5b5f8381526002602090815260408083208584528252918290208390559051828152839185917f7aa7b02a23d0a067ba64c3dfd437ba883a6bad048794658cead64a3305c044cb910160405180910390a3505050565b7f00000000000000000000000000000000000000000000000000000000000000006001600160a01b03166391d148547f00000000000000000000000000000000000000000000000000000000000000006001600160a01b031663a217fddf6040518163ffffffff1660e01b8152600401602060405180830381865afa158015610cdb573d5f5f3e3d5ffd5b505050506040513d601f19601f82011682018060405250810190610cff91906117ca565b6040516001600160e01b031960e084901b1681526004810191909152336024820152604401602060405180830381865afa158015610d3f573d5f5f3e3d5ffd5b505050506040513d601f19601f82011682018060405250810190610d6391906117e1565b610d8057604051637bfa4b9f60e01b815260040160405180910390fd5b5f610d8a83610fca565b90506004600582015460ff166004811115610da757610da7611576565b03610dcd576040516390315de160e01b8152600481018490526024015b60405180910390fd5b60058101805460ff1916600417905560405182815283907fa75610f599cea10d1a6bb6332878af5fffa58f57b227cba0d0229fa90a63cc519060200160405180910390a2505050565b7f00000000000000000000000000000000000000000000000000000000000000006001600160a01b03166391d148547f00000000000000000000000000000000000000000000000000000000000000006001600160a01b031663926d7d7f6040518163ffffffff1660e01b8152600401602060405180830381865afa158015610ea1573d5f5f3e3d5ffd5b505050506040513d601f19601f82011682018060405250810190610ec591906117ca565b6040516001600160e01b031960e084901b1681526004810191909152336024820152604401602060405180830381865afa158015610f05573d5f5f3e3d5ffd5b505050506040513d601f19601f82011682018060405250810190610f2991906117e1565b610f465760405163c64891a560e01b815260040160405180910390fd5b610f4f82610fca565b50817fb2f54fcb033e122b6e806410edf3a737bf74e4ba51301f0d2d3927fb269437a182604051610f8291815260200190565b60405180910390a25050565b5f546001600160a01b03163314610fb85760405163918cb1dd60e01b815260040160405180910390fd5b80610fc283611137565b600101555050565b5f81815260016020526040812090600582015460ff166004811115610ff157610ff1611576565b0361101257604051630d55cb5960e41b815260048101839052602401610dc4565b919050565b841580611022575080155b1561104057604051637c946ed760e01b815260040160405180910390fd5b5f8581526001602052604081206005015460ff16600481111561106557611065611576565b1461108657604051636565a01760e11b815260048101869052602401610dc4565b6040518060c00160405280858152602001848152602001838152602001838152602001828152602001600160048111156110c2576110c2611576565b90525f8681526001602081815260409283902084518155908401518183015591830151600283015560608301516003830155608083015160048084019190915560a08401516005840180549193909260ff1990921691849081111561112957611129611576565b021790555050505050505050565b5f61114182610fca565b90506001600582015460ff16600481111561115e5761115e611576565b1461101257604051631e9b4ec360e21b815260048101839052602401610dc4565b5f5f5f5f60015f89895f818110611198576111986118a5565b9050602002013581526020019081526020015f20905080600101549350805f015492505f5b87811015611316575f6111e78a8a848181106111db576111db6118a5565b90506020020135611137565b90508581600101541415806111fd575080548514155b1561123757898983818110611214576112146118a5565b9050602002013560405163cf5394dd60e01b8152600401610dc491815260200190565b5f88888481811061124a5761124a6118a5565b905060200201359050805f0361127357604051637c946ed760e01b815260040160405180910390fd5b81600301548111156112ca578a8a84818110611291576112916118a5565b600385015460405163ee70308f60e01b815260209092029390930135600482015260248101849052604481019290925250606401610dc4565b80826003015f8282546112dd91906118cd565b909155505060038201545f036112fd5760058201805460ff191660021790555b61130781866118e6565b945050508060010190506111bd565b50509450945094915050565b5f84815260026020908152604080832085845290915281205460609181900361136857604051631e802fb160e21b81526004810187905260248101859052604401610dc4565b5f8767ffffffffffffffff811115611382576113826118f9565b6040519080825280602002602001820160405280156113ab578160200160208202803683370190505b5092505f5b888110156114ba578989828181106113ca576113ca6118a5565b905060600201602001355f036113f357604051637c946ed760e01b815260040160405180910390fd5b898982818110611405576114056118a5565b905060600201602001358261141a91906118e6565b915089898281811061142e5761142e6118a5565b9050606002015f0135848281518110611449576114496118a5565b6020026020010181815250506114b28a8a8381811061146a5761146a6118a5565b9050606002015f013589898d8d86818110611487576114876118a5565b905060600201602001358e8e878181106114a3576114a36118a5565b90506060020160400135611017565b6001016113b0565b506114c5828561190d565b6114d2620f42408361190d565b11156115135780620f42406114e7848761190d565b6114f19190611924565b60405163d522a57f60e01b815260048101929092526024820152604401610dc4565b50509695505050505050565b6040518060c001604052805f81526020015f81526020015f81526020015f81526020015f81526020015f600481111561155a5761155a611576565b905290565b5f6020828403121561156f575f5ffd5b5035919050565b634e487b7160e01b5f52602160045260245ffd5b5f60c082019050825182526020830151602083015260408301516040830152606083015160608301526080830151608083015260a0830151600581106115de57634e487b7160e01b5f52602160045260245ffd5b8060a08401525092915050565b5f5f604083850312156115fc575f5ffd5b50508035926020909101359150565b5f5f5f5f5f5f60c08789031215611620575f5ffd5b505084359660208601359650604086013595606081013595506080810135945060a0013592509050565b5f6020828403121561165a575f5ffd5b81356001600160a01b0381168114611670575f5ffd5b9392505050565b5f5f83601f840112611687575f5ffd5b50813567ffffffffffffffff81111561169e575f5ffd5b6020830191508360208260051b85010111156116b8575f5ffd5b9250929050565b5f5f5f5f5f5f5f5f60a0898b0312156116d6575f5ffd5b883567ffffffffffffffff8111156116ec575f5ffd5b6116f88b828c01611677565b909950975050602089013567ffffffffffffffff811115611717575f5ffd5b6117238b828c01611677565b90975095505060408901359350606089013567ffffffffffffffff811115611749575f5ffd5b8901601f81018b13611759575f5ffd5b803567ffffffffffffffff81111561176f575f5ffd5b8b6020606083028401011115611783575f5ffd5b989b979a509598949793966020909601959394608001359392505050565b5f5f5f606084860312156117b3575f5ffd5b505081359360208301359350604090920135919050565b5f602082840312156117da575f5ffd5b5051919050565b5f602082840312156117f1575f5ffd5b81518015158114611670575f5ffd5b8183525f6001600160fb1b03831115611817575f5ffd5b8260051b80836020870137939093016020019392505050565b608081525f61184360808301888a611800565b8281036020840152611856818789611800565b8381036040850152855180825260208088019350909101905f5b8181101561188e578351835260209384019390920191600101611870565b505060609390930193909352509695505050505050565b634e487b7160e01b5f52603260045260245ffd5b634e487b7160e01b5f52601160045260245ffd5b818103818111156118e0576118e06118b9565b92915050565b808201808211156118e0576118e06118b9565b634e487b7160e01b5f52604160045260245ffd5b80820281158282048414176118e0576118e06118b9565b5f8261193e57634e487b7160e01b5f52601260045260245ffd5b50049056fea2646970667358221220d0444140a98d72a0204afae9ac05efa44fe831e92b9831bc62fdf3928f28271364736f6c63430008220033",
}

// PassportRegistryABI is the input ABI used to generate the binding from.
// Deprecated: Use PassportRegistryMetaData.ABI instead.
var PassportRegistryABI = PassportRegistryMetaData.ABI

// PassportRegistryBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use PassportRegistryMetaData.Bin instead.
var PassportRegistryBin = PassportRegistryMetaData.Bin

// DeployPassportRegistry deploys a new Ethereum contract, binding an instance of PassportRegistry to it.
func DeployPassportRegistry(auth *bind.TransactOpts, backend bind.ContractBackend, orgs_ common.Address, facilities_ common.Address) (common.Address, *types.Transaction, *PassportRegistry, error) {
	parsed, err := PassportRegistryMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(PassportRegistryBin), backend, orgs_, facilities_)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &PassportRegistry{PassportRegistryCaller: PassportRegistryCaller{contract: contract}, PassportRegistryTransactor: PassportRegistryTransactor{contract: contract}, PassportRegistryFilterer: PassportRegistryFilterer{contract: contract}}, nil
}

// PassportRegistry is an auto generated Go binding around an Ethereum contract.
type PassportRegistry struct {
	PassportRegistryCaller     // Read-only binding to the contract
	PassportRegistryTransactor // Write-only binding to the contract
	PassportRegistryFilterer   // Log filterer for contract events
}

// PassportRegistryCaller is an auto generated read-only Go binding around an Ethereum contract.
type PassportRegistryCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// PassportRegistryTransactor is an auto generated write-only Go binding around an Ethereum contract.
type PassportRegistryTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// PassportRegistryFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type PassportRegistryFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// PassportRegistrySession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type PassportRegistrySession struct {
	Contract     *PassportRegistry // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// PassportRegistryCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type PassportRegistryCallerSession struct {
	Contract *PassportRegistryCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts           // Call options to use throughout this session
}

// PassportRegistryTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type PassportRegistryTransactorSession struct {
	Contract     *PassportRegistryTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts           // Transaction auth options to use throughout this session
}

// PassportRegistryRaw is an auto generated low-level Go binding around an Ethereum contract.
type PassportRegistryRaw struct {
	Contract *PassportRegistry // Generic contract binding to access the raw methods on
}

// PassportRegistryCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type PassportRegistryCallerRaw struct {
	Contract *PassportRegistryCaller // Generic read-only contract binding to access the raw methods on
}

// PassportRegistryTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type PassportRegistryTransactorRaw struct {
	Contract *PassportRegistryTransactor // Generic write-only contract binding to access the raw methods on
}

// NewPassportRegistry creates a new instance of PassportRegistry, bound to a specific deployed contract.
func NewPassportRegistry(address common.Address, backend bind.ContractBackend) (*PassportRegistry, error) {
	contract, err := bindPassportRegistry(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &PassportRegistry{PassportRegistryCaller: PassportRegistryCaller{contract: contract}, PassportRegistryTransactor: PassportRegistryTransactor{contract: contract}, PassportRegistryFilterer: PassportRegistryFilterer{contract: contract}}, nil
}

// NewPassportRegistryCaller creates a new read-only instance of PassportRegistry, bound to a specific deployed contract.
func NewPassportRegistryCaller(address common.Address, caller bind.ContractCaller) (*PassportRegistryCaller, error) {
	contract, err := bindPassportRegistry(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &PassportRegistryCaller{contract: contract}, nil
}

// NewPassportRegistryTransactor creates a new write-only instance of PassportRegistry, bound to a specific deployed contract.
func NewPassportRegistryTransactor(address common.Address, transactor bind.ContractTransactor) (*PassportRegistryTransactor, error) {
	contract, err := bindPassportRegistry(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &PassportRegistryTransactor{contract: contract}, nil
}

// NewPassportRegistryFilterer creates a new log filterer instance of PassportRegistry, bound to a specific deployed contract.
func NewPassportRegistryFilterer(address common.Address, filterer bind.ContractFilterer) (*PassportRegistryFilterer, error) {
	contract, err := bindPassportRegistry(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &PassportRegistryFilterer{contract: contract}, nil
}

// bindPassportRegistry binds a generic wrapper to an already deployed contract.
func bindPassportRegistry(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := PassportRegistryMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_PassportRegistry *PassportRegistryRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _PassportRegistry.Contract.PassportRegistryCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_PassportRegistry *PassportRegistryRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _PassportRegistry.Contract.PassportRegistryTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_PassportRegistry *PassportRegistryRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _PassportRegistry.Contract.PassportRegistryTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_PassportRegistry *PassportRegistryCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _PassportRegistry.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_PassportRegistry *PassportRegistryTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _PassportRegistry.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_PassportRegistry *PassportRegistryTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _PassportRegistry.Contract.contract.Transact(opts, method, params...)
}

// YIELDSCALE is a free data retrieval call binding the contract method 0x6267228f.
//
// Solidity: function YIELD_SCALE() view returns(uint256)
func (_PassportRegistry *PassportRegistryCaller) YIELDSCALE(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _PassportRegistry.contract.Call(opts, &out, "YIELD_SCALE")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// YIELDSCALE is a free data retrieval call binding the contract method 0x6267228f.
//
// Solidity: function YIELD_SCALE() view returns(uint256)
func (_PassportRegistry *PassportRegistrySession) YIELDSCALE() (*big.Int, error) {
	return _PassportRegistry.Contract.YIELDSCALE(&_PassportRegistry.CallOpts)
}

// YIELDSCALE is a free data retrieval call binding the contract method 0x6267228f.
//
// Solidity: function YIELD_SCALE() view returns(uint256)
func (_PassportRegistry *PassportRegistryCallerSession) YIELDSCALE() (*big.Int, error) {
	return _PassportRegistry.Contract.YIELDSCALE(&_PassportRegistry.CallOpts)
}

// Custodian is a free data retrieval call binding the contract method 0x375b74c3.
//
// Solidity: function custodian() view returns(address)
func (_PassportRegistry *PassportRegistryCaller) Custodian(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _PassportRegistry.contract.Call(opts, &out, "custodian")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Custodian is a free data retrieval call binding the contract method 0x375b74c3.
//
// Solidity: function custodian() view returns(address)
func (_PassportRegistry *PassportRegistrySession) Custodian() (common.Address, error) {
	return _PassportRegistry.Contract.Custodian(&_PassportRegistry.CallOpts)
}

// Custodian is a free data retrieval call binding the contract method 0x375b74c3.
//
// Solidity: function custodian() view returns(address)
func (_PassportRegistry *PassportRegistryCallerSession) Custodian() (common.Address, error) {
	return _PassportRegistry.Contract.Custodian(&_PassportRegistry.CallOpts)
}

// Facilities is a free data retrieval call binding the contract method 0xe1050ef9.
//
// Solidity: function facilities() view returns(address)
func (_PassportRegistry *PassportRegistryCaller) Facilities(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _PassportRegistry.contract.Call(opts, &out, "facilities")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Facilities is a free data retrieval call binding the contract method 0xe1050ef9.
//
// Solidity: function facilities() view returns(address)
func (_PassportRegistry *PassportRegistrySession) Facilities() (common.Address, error) {
	return _PassportRegistry.Contract.Facilities(&_PassportRegistry.CallOpts)
}

// Facilities is a free data retrieval call binding the contract method 0xe1050ef9.
//
// Solidity: function facilities() view returns(address)
func (_PassportRegistry *PassportRegistryCallerSession) Facilities() (common.Address, error) {
	return _PassportRegistry.Contract.Facilities(&_PassportRegistry.CallOpts)
}

// MaxYield is a free data retrieval call binding the contract method 0x0a90f4f1.
//
// Solidity: function maxYield(bytes32 commodity, bytes32 stage) view returns(uint256)
func (_PassportRegistry *PassportRegistryCaller) MaxYield(opts *bind.CallOpts, commodity [32]byte, stage [32]byte) (*big.Int, error) {
	var out []interface{}
	err := _PassportRegistry.contract.Call(opts, &out, "maxYield", commodity, stage)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// MaxYield is a free data retrieval call binding the contract method 0x0a90f4f1.
//
// Solidity: function maxYield(bytes32 commodity, bytes32 stage) view returns(uint256)
func (_PassportRegistry *PassportRegistrySession) MaxYield(commodity [32]byte, stage [32]byte) (*big.Int, error) {
	return _PassportRegistry.Contract.MaxYield(&_PassportRegistry.CallOpts, commodity, stage)
}

// MaxYield is a free data retrieval call binding the contract method 0x0a90f4f1.
//
// Solidity: function maxYield(bytes32 commodity, bytes32 stage) view returns(uint256)
func (_PassportRegistry *PassportRegistryCallerSession) MaxYield(commodity [32]byte, stage [32]byte) (*big.Int, error) {
	return _PassportRegistry.Contract.MaxYield(&_PassportRegistry.CallOpts, commodity, stage)
}

// Orgs is a free data retrieval call binding the contract method 0xcab19cdc.
//
// Solidity: function orgs() view returns(address)
func (_PassportRegistry *PassportRegistryCaller) Orgs(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _PassportRegistry.contract.Call(opts, &out, "orgs")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Orgs is a free data retrieval call binding the contract method 0xcab19cdc.
//
// Solidity: function orgs() view returns(address)
func (_PassportRegistry *PassportRegistrySession) Orgs() (common.Address, error) {
	return _PassportRegistry.Contract.Orgs(&_PassportRegistry.CallOpts)
}

// Orgs is a free data retrieval call binding the contract method 0xcab19cdc.
//
// Solidity: function orgs() view returns(address)
func (_PassportRegistry *PassportRegistryCallerSession) Orgs() (common.Address, error) {
	return _PassportRegistry.Contract.Orgs(&_PassportRegistry.CallOpts)
}

// Passport is a free data retrieval call binding the contract method 0x0919afa0.
//
// Solidity: function passport(bytes32 passportId) view returns((bytes32,bytes32,uint256,uint256,bytes32,uint8))
func (_PassportRegistry *PassportRegistryCaller) Passport(opts *bind.CallOpts, passportId [32]byte) (PassportRegistryPassport, error) {
	var out []interface{}
	err := _PassportRegistry.contract.Call(opts, &out, "passport", passportId)

	if err != nil {
		return *new(PassportRegistryPassport), err
	}

	out0 := *abi.ConvertType(out[0], new(PassportRegistryPassport)).(*PassportRegistryPassport)

	return out0, err

}

// Passport is a free data retrieval call binding the contract method 0x0919afa0.
//
// Solidity: function passport(bytes32 passportId) view returns((bytes32,bytes32,uint256,uint256,bytes32,uint8))
func (_PassportRegistry *PassportRegistrySession) Passport(passportId [32]byte) (PassportRegistryPassport, error) {
	return _PassportRegistry.Contract.Passport(&_PassportRegistry.CallOpts, passportId)
}

// Passport is a free data retrieval call binding the contract method 0x0919afa0.
//
// Solidity: function passport(bytes32 passportId) view returns((bytes32,bytes32,uint256,uint256,bytes32,uint8))
func (_PassportRegistry *PassportRegistryCallerSession) Passport(passportId [32]byte) (PassportRegistryPassport, error) {
	return _PassportRegistry.Contract.Passport(&_PassportRegistry.CallOpts, passportId)
}

// Flag is a paid mutator transaction binding the contract method 0xe6f93e57.
//
// Solidity: function flag(bytes32 passportId, bytes32 reasonHash) returns()
func (_PassportRegistry *PassportRegistryTransactor) Flag(opts *bind.TransactOpts, passportId [32]byte, reasonHash [32]byte) (*types.Transaction, error) {
	return _PassportRegistry.contract.Transact(opts, "flag", passportId, reasonHash)
}

// Flag is a paid mutator transaction binding the contract method 0xe6f93e57.
//
// Solidity: function flag(bytes32 passportId, bytes32 reasonHash) returns()
func (_PassportRegistry *PassportRegistrySession) Flag(passportId [32]byte, reasonHash [32]byte) (*types.Transaction, error) {
	return _PassportRegistry.Contract.Flag(&_PassportRegistry.TransactOpts, passportId, reasonHash)
}

// Flag is a paid mutator transaction binding the contract method 0xe6f93e57.
//
// Solidity: function flag(bytes32 passportId, bytes32 reasonHash) returns()
func (_PassportRegistry *PassportRegistryTransactorSession) Flag(passportId [32]byte, reasonHash [32]byte) (*types.Transaction, error) {
	return _PassportRegistry.Contract.Flag(&_PassportRegistry.TransactOpts, passportId, reasonHash)
}

// Issue is a paid mutator transaction binding the contract method 0x179868c1.
//
// Solidity: function issue(bytes32 passportId, bytes32 commodity, bytes32 facilityId, uint256 period, uint256 quantity, bytes32 dataHash) returns()
func (_PassportRegistry *PassportRegistryTransactor) Issue(opts *bind.TransactOpts, passportId [32]byte, commodity [32]byte, facilityId [32]byte, period *big.Int, quantity *big.Int, dataHash [32]byte) (*types.Transaction, error) {
	return _PassportRegistry.contract.Transact(opts, "issue", passportId, commodity, facilityId, period, quantity, dataHash)
}

// Issue is a paid mutator transaction binding the contract method 0x179868c1.
//
// Solidity: function issue(bytes32 passportId, bytes32 commodity, bytes32 facilityId, uint256 period, uint256 quantity, bytes32 dataHash) returns()
func (_PassportRegistry *PassportRegistrySession) Issue(passportId [32]byte, commodity [32]byte, facilityId [32]byte, period *big.Int, quantity *big.Int, dataHash [32]byte) (*types.Transaction, error) {
	return _PassportRegistry.Contract.Issue(&_PassportRegistry.TransactOpts, passportId, commodity, facilityId, period, quantity, dataHash)
}

// Issue is a paid mutator transaction binding the contract method 0x179868c1.
//
// Solidity: function issue(bytes32 passportId, bytes32 commodity, bytes32 facilityId, uint256 period, uint256 quantity, bytes32 dataHash) returns()
func (_PassportRegistry *PassportRegistryTransactorSession) Issue(passportId [32]byte, commodity [32]byte, facilityId [32]byte, period *big.Int, quantity *big.Int, dataHash [32]byte) (*types.Transaction, error) {
	return _PassportRegistry.Contract.Issue(&_PassportRegistry.TransactOpts, passportId, commodity, facilityId, period, quantity, dataHash)
}

// Retire is a paid mutator transaction binding the contract method 0x1f2b6f23.
//
// Solidity: function retire(bytes32 passportId) returns()
func (_PassportRegistry *PassportRegistryTransactor) Retire(opts *bind.TransactOpts, passportId [32]byte) (*types.Transaction, error) {
	return _PassportRegistry.contract.Transact(opts, "retire", passportId)
}

// Retire is a paid mutator transaction binding the contract method 0x1f2b6f23.
//
// Solidity: function retire(bytes32 passportId) returns()
func (_PassportRegistry *PassportRegistrySession) Retire(passportId [32]byte) (*types.Transaction, error) {
	return _PassportRegistry.Contract.Retire(&_PassportRegistry.TransactOpts, passportId)
}

// Retire is a paid mutator transaction binding the contract method 0x1f2b6f23.
//
// Solidity: function retire(bytes32 passportId) returns()
func (_PassportRegistry *PassportRegistryTransactorSession) Retire(passportId [32]byte) (*types.Transaction, error) {
	return _PassportRegistry.Contract.Retire(&_PassportRegistry.TransactOpts, passportId)
}

// Revoke is a paid mutator transaction binding the contract method 0xc2664610.
//
// Solidity: function revoke(bytes32 passportId, bytes32 reasonHash) returns()
func (_PassportRegistry *PassportRegistryTransactor) Revoke(opts *bind.TransactOpts, passportId [32]byte, reasonHash [32]byte) (*types.Transaction, error) {
	return _PassportRegistry.contract.Transact(opts, "revoke", passportId, reasonHash)
}

// Revoke is a paid mutator transaction binding the contract method 0xc2664610.
//
// Solidity: function revoke(bytes32 passportId, bytes32 reasonHash) returns()
func (_PassportRegistry *PassportRegistrySession) Revoke(passportId [32]byte, reasonHash [32]byte) (*types.Transaction, error) {
	return _PassportRegistry.Contract.Revoke(&_PassportRegistry.TransactOpts, passportId, reasonHash)
}

// Revoke is a paid mutator transaction binding the contract method 0xc2664610.
//
// Solidity: function revoke(bytes32 passportId, bytes32 reasonHash) returns()
func (_PassportRegistry *PassportRegistryTransactorSession) Revoke(passportId [32]byte, reasonHash [32]byte) (*types.Transaction, error) {
	return _PassportRegistry.Contract.Revoke(&_PassportRegistry.TransactOpts, passportId, reasonHash)
}

// SetCustodian is a paid mutator transaction binding the contract method 0x403f3731.
//
// Solidity: function setCustodian(address custodian_) returns()
func (_PassportRegistry *PassportRegistryTransactor) SetCustodian(opts *bind.TransactOpts, custodian_ common.Address) (*types.Transaction, error) {
	return _PassportRegistry.contract.Transact(opts, "setCustodian", custodian_)
}

// SetCustodian is a paid mutator transaction binding the contract method 0x403f3731.
//
// Solidity: function setCustodian(address custodian_) returns()
func (_PassportRegistry *PassportRegistrySession) SetCustodian(custodian_ common.Address) (*types.Transaction, error) {
	return _PassportRegistry.Contract.SetCustodian(&_PassportRegistry.TransactOpts, custodian_)
}

// SetCustodian is a paid mutator transaction binding the contract method 0x403f3731.
//
// Solidity: function setCustodian(address custodian_) returns()
func (_PassportRegistry *PassportRegistryTransactorSession) SetCustodian(custodian_ common.Address) (*types.Transaction, error) {
	return _PassportRegistry.Contract.SetCustodian(&_PassportRegistry.TransactOpts, custodian_)
}

// SetMaxYield is a paid mutator transaction binding the contract method 0x8db91b46.
//
// Solidity: function setMaxYield(bytes32 commodity, bytes32 stage, uint256 partsPerMillion) returns()
func (_PassportRegistry *PassportRegistryTransactor) SetMaxYield(opts *bind.TransactOpts, commodity [32]byte, stage [32]byte, partsPerMillion *big.Int) (*types.Transaction, error) {
	return _PassportRegistry.contract.Transact(opts, "setMaxYield", commodity, stage, partsPerMillion)
}

// SetMaxYield is a paid mutator transaction binding the contract method 0x8db91b46.
//
// Solidity: function setMaxYield(bytes32 commodity, bytes32 stage, uint256 partsPerMillion) returns()
func (_PassportRegistry *PassportRegistrySession) SetMaxYield(commodity [32]byte, stage [32]byte, partsPerMillion *big.Int) (*types.Transaction, error) {
	return _PassportRegistry.Contract.SetMaxYield(&_PassportRegistry.TransactOpts, commodity, stage, partsPerMillion)
}

// SetMaxYield is a paid mutator transaction binding the contract method 0x8db91b46.
//
// Solidity: function setMaxYield(bytes32 commodity, bytes32 stage, uint256 partsPerMillion) returns()
func (_PassportRegistry *PassportRegistryTransactorSession) SetMaxYield(commodity [32]byte, stage [32]byte, partsPerMillion *big.Int) (*types.Transaction, error) {
	return _PassportRegistry.Contract.SetMaxYield(&_PassportRegistry.TransactOpts, commodity, stage, partsPerMillion)
}

// TransferCustody is a paid mutator transaction binding the contract method 0xfa2deb7f.
//
// Solidity: function transferCustody(bytes32 passportId, bytes32 newHolder) returns()
func (_PassportRegistry *PassportRegistryTransactor) TransferCustody(opts *bind.TransactOpts, passportId [32]byte, newHolder [32]byte) (*types.Transaction, error) {
	return _PassportRegistry.contract.Transact(opts, "transferCustody", passportId, newHolder)
}

// TransferCustody is a paid mutator transaction binding the contract method 0xfa2deb7f.
//
// Solidity: function transferCustody(bytes32 passportId, bytes32 newHolder) returns()
func (_PassportRegistry *PassportRegistrySession) TransferCustody(passportId [32]byte, newHolder [32]byte) (*types.Transaction, error) {
	return _PassportRegistry.Contract.TransferCustody(&_PassportRegistry.TransactOpts, passportId, newHolder)
}

// TransferCustody is a paid mutator transaction binding the contract method 0xfa2deb7f.
//
// Solidity: function transferCustody(bytes32 passportId, bytes32 newHolder) returns()
func (_PassportRegistry *PassportRegistryTransactorSession) TransferCustody(passportId [32]byte, newHolder [32]byte) (*types.Transaction, error) {
	return _PassportRegistry.Contract.TransferCustody(&_PassportRegistry.TransactOpts, passportId, newHolder)
}

// Transform is a paid mutator transaction binding the contract method 0x8cb52bb6.
//
// Solidity: function transform(bytes32[] inputs, uint256[] inputsUsed, bytes32 stage, (bytes32,uint256,bytes32)[] outputs, bytes32 dataHash) returns()
func (_PassportRegistry *PassportRegistryTransactor) Transform(opts *bind.TransactOpts, inputs [][32]byte, inputsUsed []*big.Int, stage [32]byte, outputs []PassportRegistryOutput, dataHash [32]byte) (*types.Transaction, error) {
	return _PassportRegistry.contract.Transact(opts, "transform", inputs, inputsUsed, stage, outputs, dataHash)
}

// Transform is a paid mutator transaction binding the contract method 0x8cb52bb6.
//
// Solidity: function transform(bytes32[] inputs, uint256[] inputsUsed, bytes32 stage, (bytes32,uint256,bytes32)[] outputs, bytes32 dataHash) returns()
func (_PassportRegistry *PassportRegistrySession) Transform(inputs [][32]byte, inputsUsed []*big.Int, stage [32]byte, outputs []PassportRegistryOutput, dataHash [32]byte) (*types.Transaction, error) {
	return _PassportRegistry.Contract.Transform(&_PassportRegistry.TransactOpts, inputs, inputsUsed, stage, outputs, dataHash)
}

// Transform is a paid mutator transaction binding the contract method 0x8cb52bb6.
//
// Solidity: function transform(bytes32[] inputs, uint256[] inputsUsed, bytes32 stage, (bytes32,uint256,bytes32)[] outputs, bytes32 dataHash) returns()
func (_PassportRegistry *PassportRegistryTransactorSession) Transform(inputs [][32]byte, inputsUsed []*big.Int, stage [32]byte, outputs []PassportRegistryOutput, dataHash [32]byte) (*types.Transaction, error) {
	return _PassportRegistry.Contract.Transform(&_PassportRegistry.TransactOpts, inputs, inputsUsed, stage, outputs, dataHash)
}

// PassportRegistryCustodianSetIterator is returned from FilterCustodianSet and is used to iterate over the raw logs and unpacked data for CustodianSet events raised by the PassportRegistry contract.
type PassportRegistryCustodianSetIterator struct {
	Event *PassportRegistryCustodianSet // Event containing the contract specifics and raw log

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
func (it *PassportRegistryCustodianSetIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(PassportRegistryCustodianSet)
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
		it.Event = new(PassportRegistryCustodianSet)
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
func (it *PassportRegistryCustodianSetIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *PassportRegistryCustodianSetIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// PassportRegistryCustodianSet represents a CustodianSet event raised by the PassportRegistry contract.
type PassportRegistryCustodianSet struct {
	Custodian common.Address
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterCustodianSet is a free log retrieval operation binding the contract event 0xb88c20a211c5d7677ba2a26c317d8ae6b25aa492016dc8ceca2469761d063d80.
//
// Solidity: event CustodianSet(address custodian)
func (_PassportRegistry *PassportRegistryFilterer) FilterCustodianSet(opts *bind.FilterOpts) (*PassportRegistryCustodianSetIterator, error) {

	logs, sub, err := _PassportRegistry.contract.FilterLogs(opts, "CustodianSet")
	if err != nil {
		return nil, err
	}
	return &PassportRegistryCustodianSetIterator{contract: _PassportRegistry.contract, event: "CustodianSet", logs: logs, sub: sub}, nil
}

// WatchCustodianSet is a free log subscription operation binding the contract event 0xb88c20a211c5d7677ba2a26c317d8ae6b25aa492016dc8ceca2469761d063d80.
//
// Solidity: event CustodianSet(address custodian)
func (_PassportRegistry *PassportRegistryFilterer) WatchCustodianSet(opts *bind.WatchOpts, sink chan<- *PassportRegistryCustodianSet) (event.Subscription, error) {

	logs, sub, err := _PassportRegistry.contract.WatchLogs(opts, "CustodianSet")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(PassportRegistryCustodianSet)
				if err := _PassportRegistry.contract.UnpackLog(event, "CustodianSet", log); err != nil {
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

// ParseCustodianSet is a log parse operation binding the contract event 0xb88c20a211c5d7677ba2a26c317d8ae6b25aa492016dc8ceca2469761d063d80.
//
// Solidity: event CustodianSet(address custodian)
func (_PassportRegistry *PassportRegistryFilterer) ParseCustodianSet(log types.Log) (*PassportRegistryCustodianSet, error) {
	event := new(PassportRegistryCustodianSet)
	if err := _PassportRegistry.contract.UnpackLog(event, "CustodianSet", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// PassportRegistryFlaggedIterator is returned from FilterFlagged and is used to iterate over the raw logs and unpacked data for Flagged events raised by the PassportRegistry contract.
type PassportRegistryFlaggedIterator struct {
	Event *PassportRegistryFlagged // Event containing the contract specifics and raw log

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
func (it *PassportRegistryFlaggedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(PassportRegistryFlagged)
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
		it.Event = new(PassportRegistryFlagged)
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
func (it *PassportRegistryFlaggedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *PassportRegistryFlaggedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// PassportRegistryFlagged represents a Flagged event raised by the PassportRegistry contract.
type PassportRegistryFlagged struct {
	PassportId [32]byte
	ReasonHash [32]byte
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterFlagged is a free log retrieval operation binding the contract event 0xb2f54fcb033e122b6e806410edf3a737bf74e4ba51301f0d2d3927fb269437a1.
//
// Solidity: event Flagged(bytes32 indexed passportId, bytes32 reasonHash)
func (_PassportRegistry *PassportRegistryFilterer) FilterFlagged(opts *bind.FilterOpts, passportId [][32]byte) (*PassportRegistryFlaggedIterator, error) {

	var passportIdRule []interface{}
	for _, passportIdItem := range passportId {
		passportIdRule = append(passportIdRule, passportIdItem)
	}

	logs, sub, err := _PassportRegistry.contract.FilterLogs(opts, "Flagged", passportIdRule)
	if err != nil {
		return nil, err
	}
	return &PassportRegistryFlaggedIterator{contract: _PassportRegistry.contract, event: "Flagged", logs: logs, sub: sub}, nil
}

// WatchFlagged is a free log subscription operation binding the contract event 0xb2f54fcb033e122b6e806410edf3a737bf74e4ba51301f0d2d3927fb269437a1.
//
// Solidity: event Flagged(bytes32 indexed passportId, bytes32 reasonHash)
func (_PassportRegistry *PassportRegistryFilterer) WatchFlagged(opts *bind.WatchOpts, sink chan<- *PassportRegistryFlagged, passportId [][32]byte) (event.Subscription, error) {

	var passportIdRule []interface{}
	for _, passportIdItem := range passportId {
		passportIdRule = append(passportIdRule, passportIdItem)
	}

	logs, sub, err := _PassportRegistry.contract.WatchLogs(opts, "Flagged", passportIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(PassportRegistryFlagged)
				if err := _PassportRegistry.contract.UnpackLog(event, "Flagged", log); err != nil {
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

// ParseFlagged is a log parse operation binding the contract event 0xb2f54fcb033e122b6e806410edf3a737bf74e4ba51301f0d2d3927fb269437a1.
//
// Solidity: event Flagged(bytes32 indexed passportId, bytes32 reasonHash)
func (_PassportRegistry *PassportRegistryFilterer) ParseFlagged(log types.Log) (*PassportRegistryFlagged, error) {
	event := new(PassportRegistryFlagged)
	if err := _PassportRegistry.contract.UnpackLog(event, "Flagged", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// PassportRegistryLotTransformedIterator is returned from FilterLotTransformed and is used to iterate over the raw logs and unpacked data for LotTransformed events raised by the PassportRegistry contract.
type PassportRegistryLotTransformedIterator struct {
	Event *PassportRegistryLotTransformed // Event containing the contract specifics and raw log

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
func (it *PassportRegistryLotTransformedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(PassportRegistryLotTransformed)
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
		it.Event = new(PassportRegistryLotTransformed)
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
func (it *PassportRegistryLotTransformedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *PassportRegistryLotTransformedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// PassportRegistryLotTransformed represents a LotTransformed event raised by the PassportRegistry contract.
type PassportRegistryLotTransformed struct {
	Holder     [32]byte
	Stage      [32]byte
	Inputs     [][32]byte
	InputsUsed []*big.Int
	Outputs    [][32]byte
	DataHash   [32]byte
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterLotTransformed is a free log retrieval operation binding the contract event 0xccc473428e0fc7f33fe2dccee21361ab5de16247ae912e847eb07e98e99c661e.
//
// Solidity: event LotTransformed(bytes32 indexed holder, bytes32 indexed stage, bytes32[] inputs, uint256[] inputsUsed, bytes32[] outputs, bytes32 dataHash)
func (_PassportRegistry *PassportRegistryFilterer) FilterLotTransformed(opts *bind.FilterOpts, holder [][32]byte, stage [][32]byte) (*PassportRegistryLotTransformedIterator, error) {

	var holderRule []interface{}
	for _, holderItem := range holder {
		holderRule = append(holderRule, holderItem)
	}
	var stageRule []interface{}
	for _, stageItem := range stage {
		stageRule = append(stageRule, stageItem)
	}

	logs, sub, err := _PassportRegistry.contract.FilterLogs(opts, "LotTransformed", holderRule, stageRule)
	if err != nil {
		return nil, err
	}
	return &PassportRegistryLotTransformedIterator{contract: _PassportRegistry.contract, event: "LotTransformed", logs: logs, sub: sub}, nil
}

// WatchLotTransformed is a free log subscription operation binding the contract event 0xccc473428e0fc7f33fe2dccee21361ab5de16247ae912e847eb07e98e99c661e.
//
// Solidity: event LotTransformed(bytes32 indexed holder, bytes32 indexed stage, bytes32[] inputs, uint256[] inputsUsed, bytes32[] outputs, bytes32 dataHash)
func (_PassportRegistry *PassportRegistryFilterer) WatchLotTransformed(opts *bind.WatchOpts, sink chan<- *PassportRegistryLotTransformed, holder [][32]byte, stage [][32]byte) (event.Subscription, error) {

	var holderRule []interface{}
	for _, holderItem := range holder {
		holderRule = append(holderRule, holderItem)
	}
	var stageRule []interface{}
	for _, stageItem := range stage {
		stageRule = append(stageRule, stageItem)
	}

	logs, sub, err := _PassportRegistry.contract.WatchLogs(opts, "LotTransformed", holderRule, stageRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(PassportRegistryLotTransformed)
				if err := _PassportRegistry.contract.UnpackLog(event, "LotTransformed", log); err != nil {
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

// ParseLotTransformed is a log parse operation binding the contract event 0xccc473428e0fc7f33fe2dccee21361ab5de16247ae912e847eb07e98e99c661e.
//
// Solidity: event LotTransformed(bytes32 indexed holder, bytes32 indexed stage, bytes32[] inputs, uint256[] inputsUsed, bytes32[] outputs, bytes32 dataHash)
func (_PassportRegistry *PassportRegistryFilterer) ParseLotTransformed(log types.Log) (*PassportRegistryLotTransformed, error) {
	event := new(PassportRegistryLotTransformed)
	if err := _PassportRegistry.contract.UnpackLog(event, "LotTransformed", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// PassportRegistryMaxYieldSetIterator is returned from FilterMaxYieldSet and is used to iterate over the raw logs and unpacked data for MaxYieldSet events raised by the PassportRegistry contract.
type PassportRegistryMaxYieldSetIterator struct {
	Event *PassportRegistryMaxYieldSet // Event containing the contract specifics and raw log

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
func (it *PassportRegistryMaxYieldSetIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(PassportRegistryMaxYieldSet)
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
		it.Event = new(PassportRegistryMaxYieldSet)
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
func (it *PassportRegistryMaxYieldSetIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *PassportRegistryMaxYieldSetIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// PassportRegistryMaxYieldSet represents a MaxYieldSet event raised by the PassportRegistry contract.
type PassportRegistryMaxYieldSet struct {
	Commodity       [32]byte
	Stage           [32]byte
	PartsPerMillion *big.Int
	Raw             types.Log // Blockchain specific contextual infos
}

// FilterMaxYieldSet is a free log retrieval operation binding the contract event 0x7aa7b02a23d0a067ba64c3dfd437ba883a6bad048794658cead64a3305c044cb.
//
// Solidity: event MaxYieldSet(bytes32 indexed commodity, bytes32 indexed stage, uint256 partsPerMillion)
func (_PassportRegistry *PassportRegistryFilterer) FilterMaxYieldSet(opts *bind.FilterOpts, commodity [][32]byte, stage [][32]byte) (*PassportRegistryMaxYieldSetIterator, error) {

	var commodityRule []interface{}
	for _, commodityItem := range commodity {
		commodityRule = append(commodityRule, commodityItem)
	}
	var stageRule []interface{}
	for _, stageItem := range stage {
		stageRule = append(stageRule, stageItem)
	}

	logs, sub, err := _PassportRegistry.contract.FilterLogs(opts, "MaxYieldSet", commodityRule, stageRule)
	if err != nil {
		return nil, err
	}
	return &PassportRegistryMaxYieldSetIterator{contract: _PassportRegistry.contract, event: "MaxYieldSet", logs: logs, sub: sub}, nil
}

// WatchMaxYieldSet is a free log subscription operation binding the contract event 0x7aa7b02a23d0a067ba64c3dfd437ba883a6bad048794658cead64a3305c044cb.
//
// Solidity: event MaxYieldSet(bytes32 indexed commodity, bytes32 indexed stage, uint256 partsPerMillion)
func (_PassportRegistry *PassportRegistryFilterer) WatchMaxYieldSet(opts *bind.WatchOpts, sink chan<- *PassportRegistryMaxYieldSet, commodity [][32]byte, stage [][32]byte) (event.Subscription, error) {

	var commodityRule []interface{}
	for _, commodityItem := range commodity {
		commodityRule = append(commodityRule, commodityItem)
	}
	var stageRule []interface{}
	for _, stageItem := range stage {
		stageRule = append(stageRule, stageItem)
	}

	logs, sub, err := _PassportRegistry.contract.WatchLogs(opts, "MaxYieldSet", commodityRule, stageRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(PassportRegistryMaxYieldSet)
				if err := _PassportRegistry.contract.UnpackLog(event, "MaxYieldSet", log); err != nil {
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

// ParseMaxYieldSet is a log parse operation binding the contract event 0x7aa7b02a23d0a067ba64c3dfd437ba883a6bad048794658cead64a3305c044cb.
//
// Solidity: event MaxYieldSet(bytes32 indexed commodity, bytes32 indexed stage, uint256 partsPerMillion)
func (_PassportRegistry *PassportRegistryFilterer) ParseMaxYieldSet(log types.Log) (*PassportRegistryMaxYieldSet, error) {
	event := new(PassportRegistryMaxYieldSet)
	if err := _PassportRegistry.contract.UnpackLog(event, "MaxYieldSet", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// PassportRegistryPassportIssuedIterator is returned from FilterPassportIssued and is used to iterate over the raw logs and unpacked data for PassportIssued events raised by the PassportRegistry contract.
type PassportRegistryPassportIssuedIterator struct {
	Event *PassportRegistryPassportIssued // Event containing the contract specifics and raw log

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
func (it *PassportRegistryPassportIssuedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(PassportRegistryPassportIssued)
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
		it.Event = new(PassportRegistryPassportIssued)
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
func (it *PassportRegistryPassportIssuedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *PassportRegistryPassportIssuedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// PassportRegistryPassportIssued represents a PassportIssued event raised by the PassportRegistry contract.
type PassportRegistryPassportIssued struct {
	PassportId [32]byte
	Holder     [32]byte
	FacilityId [32]byte
	Commodity  [32]byte
	Period     *big.Int
	Quantity   *big.Int
	DataHash   [32]byte
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterPassportIssued is a free log retrieval operation binding the contract event 0x8dcb03e40f2cc0d665279dd4894896cf458a980d9b7c67c903692ecbd826f33a.
//
// Solidity: event PassportIssued(bytes32 indexed passportId, bytes32 indexed holder, bytes32 indexed facilityId, bytes32 commodity, uint256 period, uint256 quantity, bytes32 dataHash)
func (_PassportRegistry *PassportRegistryFilterer) FilterPassportIssued(opts *bind.FilterOpts, passportId [][32]byte, holder [][32]byte, facilityId [][32]byte) (*PassportRegistryPassportIssuedIterator, error) {

	var passportIdRule []interface{}
	for _, passportIdItem := range passportId {
		passportIdRule = append(passportIdRule, passportIdItem)
	}
	var holderRule []interface{}
	for _, holderItem := range holder {
		holderRule = append(holderRule, holderItem)
	}
	var facilityIdRule []interface{}
	for _, facilityIdItem := range facilityId {
		facilityIdRule = append(facilityIdRule, facilityIdItem)
	}

	logs, sub, err := _PassportRegistry.contract.FilterLogs(opts, "PassportIssued", passportIdRule, holderRule, facilityIdRule)
	if err != nil {
		return nil, err
	}
	return &PassportRegistryPassportIssuedIterator{contract: _PassportRegistry.contract, event: "PassportIssued", logs: logs, sub: sub}, nil
}

// WatchPassportIssued is a free log subscription operation binding the contract event 0x8dcb03e40f2cc0d665279dd4894896cf458a980d9b7c67c903692ecbd826f33a.
//
// Solidity: event PassportIssued(bytes32 indexed passportId, bytes32 indexed holder, bytes32 indexed facilityId, bytes32 commodity, uint256 period, uint256 quantity, bytes32 dataHash)
func (_PassportRegistry *PassportRegistryFilterer) WatchPassportIssued(opts *bind.WatchOpts, sink chan<- *PassportRegistryPassportIssued, passportId [][32]byte, holder [][32]byte, facilityId [][32]byte) (event.Subscription, error) {

	var passportIdRule []interface{}
	for _, passportIdItem := range passportId {
		passportIdRule = append(passportIdRule, passportIdItem)
	}
	var holderRule []interface{}
	for _, holderItem := range holder {
		holderRule = append(holderRule, holderItem)
	}
	var facilityIdRule []interface{}
	for _, facilityIdItem := range facilityId {
		facilityIdRule = append(facilityIdRule, facilityIdItem)
	}

	logs, sub, err := _PassportRegistry.contract.WatchLogs(opts, "PassportIssued", passportIdRule, holderRule, facilityIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(PassportRegistryPassportIssued)
				if err := _PassportRegistry.contract.UnpackLog(event, "PassportIssued", log); err != nil {
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

// ParsePassportIssued is a log parse operation binding the contract event 0x8dcb03e40f2cc0d665279dd4894896cf458a980d9b7c67c903692ecbd826f33a.
//
// Solidity: event PassportIssued(bytes32 indexed passportId, bytes32 indexed holder, bytes32 indexed facilityId, bytes32 commodity, uint256 period, uint256 quantity, bytes32 dataHash)
func (_PassportRegistry *PassportRegistryFilterer) ParsePassportIssued(log types.Log) (*PassportRegistryPassportIssued, error) {
	event := new(PassportRegistryPassportIssued)
	if err := _PassportRegistry.contract.UnpackLog(event, "PassportIssued", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// PassportRegistryPassportRetiredIterator is returned from FilterPassportRetired and is used to iterate over the raw logs and unpacked data for PassportRetired events raised by the PassportRegistry contract.
type PassportRegistryPassportRetiredIterator struct {
	Event *PassportRegistryPassportRetired // Event containing the contract specifics and raw log

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
func (it *PassportRegistryPassportRetiredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(PassportRegistryPassportRetired)
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
		it.Event = new(PassportRegistryPassportRetired)
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
func (it *PassportRegistryPassportRetiredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *PassportRegistryPassportRetiredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// PassportRegistryPassportRetired represents a PassportRetired event raised by the PassportRegistry contract.
type PassportRegistryPassportRetired struct {
	PassportId [32]byte
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterPassportRetired is a free log retrieval operation binding the contract event 0xde3bd82e297099d837b8a48b5678f7a5ce1c8331c96f54d323394cd8e3f7bbaa.
//
// Solidity: event PassportRetired(bytes32 indexed passportId)
func (_PassportRegistry *PassportRegistryFilterer) FilterPassportRetired(opts *bind.FilterOpts, passportId [][32]byte) (*PassportRegistryPassportRetiredIterator, error) {

	var passportIdRule []interface{}
	for _, passportIdItem := range passportId {
		passportIdRule = append(passportIdRule, passportIdItem)
	}

	logs, sub, err := _PassportRegistry.contract.FilterLogs(opts, "PassportRetired", passportIdRule)
	if err != nil {
		return nil, err
	}
	return &PassportRegistryPassportRetiredIterator{contract: _PassportRegistry.contract, event: "PassportRetired", logs: logs, sub: sub}, nil
}

// WatchPassportRetired is a free log subscription operation binding the contract event 0xde3bd82e297099d837b8a48b5678f7a5ce1c8331c96f54d323394cd8e3f7bbaa.
//
// Solidity: event PassportRetired(bytes32 indexed passportId)
func (_PassportRegistry *PassportRegistryFilterer) WatchPassportRetired(opts *bind.WatchOpts, sink chan<- *PassportRegistryPassportRetired, passportId [][32]byte) (event.Subscription, error) {

	var passportIdRule []interface{}
	for _, passportIdItem := range passportId {
		passportIdRule = append(passportIdRule, passportIdItem)
	}

	logs, sub, err := _PassportRegistry.contract.WatchLogs(opts, "PassportRetired", passportIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(PassportRegistryPassportRetired)
				if err := _PassportRegistry.contract.UnpackLog(event, "PassportRetired", log); err != nil {
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

// ParsePassportRetired is a log parse operation binding the contract event 0xde3bd82e297099d837b8a48b5678f7a5ce1c8331c96f54d323394cd8e3f7bbaa.
//
// Solidity: event PassportRetired(bytes32 indexed passportId)
func (_PassportRegistry *PassportRegistryFilterer) ParsePassportRetired(log types.Log) (*PassportRegistryPassportRetired, error) {
	event := new(PassportRegistryPassportRetired)
	if err := _PassportRegistry.contract.UnpackLog(event, "PassportRetired", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// PassportRegistryRevokedIterator is returned from FilterRevoked and is used to iterate over the raw logs and unpacked data for Revoked events raised by the PassportRegistry contract.
type PassportRegistryRevokedIterator struct {
	Event *PassportRegistryRevoked // Event containing the contract specifics and raw log

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
func (it *PassportRegistryRevokedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(PassportRegistryRevoked)
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
		it.Event = new(PassportRegistryRevoked)
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
func (it *PassportRegistryRevokedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *PassportRegistryRevokedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// PassportRegistryRevoked represents a Revoked event raised by the PassportRegistry contract.
type PassportRegistryRevoked struct {
	PassportId [32]byte
	ReasonHash [32]byte
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterRevoked is a free log retrieval operation binding the contract event 0xa75610f599cea10d1a6bb6332878af5fffa58f57b227cba0d0229fa90a63cc51.
//
// Solidity: event Revoked(bytes32 indexed passportId, bytes32 reasonHash)
func (_PassportRegistry *PassportRegistryFilterer) FilterRevoked(opts *bind.FilterOpts, passportId [][32]byte) (*PassportRegistryRevokedIterator, error) {

	var passportIdRule []interface{}
	for _, passportIdItem := range passportId {
		passportIdRule = append(passportIdRule, passportIdItem)
	}

	logs, sub, err := _PassportRegistry.contract.FilterLogs(opts, "Revoked", passportIdRule)
	if err != nil {
		return nil, err
	}
	return &PassportRegistryRevokedIterator{contract: _PassportRegistry.contract, event: "Revoked", logs: logs, sub: sub}, nil
}

// WatchRevoked is a free log subscription operation binding the contract event 0xa75610f599cea10d1a6bb6332878af5fffa58f57b227cba0d0229fa90a63cc51.
//
// Solidity: event Revoked(bytes32 indexed passportId, bytes32 reasonHash)
func (_PassportRegistry *PassportRegistryFilterer) WatchRevoked(opts *bind.WatchOpts, sink chan<- *PassportRegistryRevoked, passportId [][32]byte) (event.Subscription, error) {

	var passportIdRule []interface{}
	for _, passportIdItem := range passportId {
		passportIdRule = append(passportIdRule, passportIdItem)
	}

	logs, sub, err := _PassportRegistry.contract.WatchLogs(opts, "Revoked", passportIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(PassportRegistryRevoked)
				if err := _PassportRegistry.contract.UnpackLog(event, "Revoked", log); err != nil {
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

// ParseRevoked is a log parse operation binding the contract event 0xa75610f599cea10d1a6bb6332878af5fffa58f57b227cba0d0229fa90a63cc51.
//
// Solidity: event Revoked(bytes32 indexed passportId, bytes32 reasonHash)
func (_PassportRegistry *PassportRegistryFilterer) ParseRevoked(log types.Log) (*PassportRegistryRevoked, error) {
	event := new(PassportRegistryRevoked)
	if err := _PassportRegistry.contract.UnpackLog(event, "Revoked", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
