// SPDX-License-Identifier: MIT
pragma solidity ^0.8.28;

import {Managed} from "./Managed.sol";
import {OrgRegistry} from "./OrgRegistry.sol";
import {FacilityRegistry} from "./FacilityRegistry.sol";

/// @title PassportRegistry
/// @notice The passports themselves: issued at an origin facility within its
///         capacity, transformed under mass balance, retired or revoked.
/// @dev Quantities are whole numbers in the commodity's smallest unit.
contract PassportRegistry is Managed {
    enum Status {
        None,
        Active,
        Consumed, // fully used as input to a transform
        Retired, // reached its final buyer
        Revoked
    }

    struct Passport {
        bytes32 commodity;
        bytes32 holder; // organisation currently holding the goods
        uint256 quantity; // as issued or created
        uint256 remaining; // not yet used in a transform
        bytes32 dataHash; // hash of the off-chain record
        Status status;
    }

    struct Output {
        bytes32 id;
        uint256 quantity;
        bytes32 dataHash;
    }

    /// @notice Yields are fractions written as parts per million: 0.2 is 200_000.
    uint256 public constant YIELD_SCALE = 1_000_000;

    FacilityRegistry public immutable facilities;
    /// @notice The only contract allowed to change a passport's holder (the CustodyManager).
    address public custodian;

    mapping(bytes32 passportId => Passport) private _passports;
    /// @notice Most output allowed per unit of input at a stage, in parts per million. 0 = not set.
    mapping(bytes32 commodity => mapping(bytes32 stage => uint256)) public maxYield;

    event PassportIssued(
        bytes32 indexed passportId, bytes32 indexed holder, bytes32 indexed facilityId, bytes32 commodity, uint256 period, uint256 quantity, bytes32 dataHash
    );
    event LotTransformed(
        bytes32 indexed holder, bytes32 indexed stage, bytes32[] inputs, uint256[] inputsUsed, bytes32[] outputs, bytes32 dataHash
    );
    event PassportRetired(bytes32 indexed passportId);
    event Revoked(bytes32 indexed passportId, bytes32 reasonHash);
    event Flagged(bytes32 indexed passportId, bytes32 reasonHash);
    event MaxYieldSet(bytes32 indexed commodity, bytes32 indexed stage, uint256 partsPerMillion);
    event CustodianSet(address custodian);

    error ZeroValue();
    error LengthMismatch();
    error PassportExists(bytes32 passportId);
    error PassportUnknown(bytes32 passportId);
    error PassportNotActive(bytes32 passportId);
    error AlreadyRevoked(bytes32 passportId);
    error InsufficientQuantity(bytes32 passportId, uint256 requested, uint256 remaining);
    error MixedInputs(bytes32 passportId);
    error YieldNotSet(bytes32 commodity, bytes32 stage);
    error MassBalanceExceeded(uint256 outputs, uint256 maxOutputs);
    error CustodianAlreadySet();
    error NotCustodian();

    constructor(OrgRegistry orgs_, FacilityRegistry facilities_) Managed(orgs_) {
        facilities = facilities_;
    }

    /// @notice One-time wiring: name the CustodyManager.
    function setCustodian(address custodian_) external onlyAdmin {
        if (custodian != address(0)) revert CustodianAlreadySet();
        if (custodian_ == address(0)) revert ZeroValue();
        custodian = custodian_;
        emit CustodianSet(custodian_);
    }

    function setMaxYield(bytes32 commodity, bytes32 stage, uint256 partsPerMillion) external onlyOperator {
        if (commodity == bytes32(0) || stage == bytes32(0) || partsPerMillion == 0) revert ZeroValue();
        maxYield[commodity][stage] = partsPerMillion;
        emit MaxYieldSet(commodity, stage, partsPerMillion);
    }

    /// @notice Issue an origin passport. The facility's organisation becomes
    ///         the holder, and the quantity counts against the facility's
    ///         capacity for the period.
    function issue(bytes32 passportId, bytes32 commodity, bytes32 facilityId, uint256 period, uint256 quantity, bytes32 dataHash)
        external
        onlyRelayer
    {
        if (commodity == bytes32(0) || quantity == 0) revert ZeroValue();
        bytes32 holder = facilities.consumeCapacity(facilityId, period, quantity);
        _create(passportId, commodity, holder, quantity, dataHash);
        emit PassportIssued(passportId, holder, facilityId, commodity, period, quantity, dataHash);
    }

    /// @notice Turn (parts of) input passports into new output passports:
    ///         processing, splitting or merging. All inputs must be active,
    ///         held by the same organisation and of the same commodity.
    ///         Mass balance: total outputs ≤ total inputs used × the stage's max yield.
    function transform(
        bytes32[] calldata inputs,
        uint256[] calldata inputsUsed,
        bytes32 stage,
        Output[] calldata outputs,
        bytes32 dataHash
    ) external onlyRelayer {
        if (inputs.length == 0 || outputs.length == 0) revert ZeroValue();
        if (inputs.length != inputsUsed.length) revert LengthMismatch();

        // Scoped so the intermediate values leave the stack before the event is emitted.
        bytes32 holder;
        bytes32[] memory outputIds;
        {
            bytes32 commodity;
            uint256 totalIn;
            (holder, commodity, totalIn) = _useInputs(inputs, inputsUsed);
            outputIds = _createOutputs(outputs, commodity, holder, stage, totalIn);
        }
        emit LotTransformed(holder, stage, inputs, inputsUsed, outputIds, dataHash);
    }

    /// @dev Takes the used quantities out of the inputs; an input used up is marked Consumed.
    function _useInputs(bytes32[] calldata inputs, uint256[] calldata inputsUsed)
        private
        returns (bytes32 holder, bytes32 commodity, uint256 totalIn)
    {
        Passport storage first = _passports[inputs[0]];
        holder = first.holder;
        commodity = first.commodity;

        for (uint256 i; i < inputs.length; ++i) {
            Passport storage p = _active(inputs[i]);
            if (p.holder != holder || p.commodity != commodity) revert MixedInputs(inputs[i]);
            uint256 used = inputsUsed[i];
            if (used == 0) revert ZeroValue();
            if (used > p.remaining) revert InsufficientQuantity(inputs[i], used, p.remaining);

            p.remaining -= used;
            if (p.remaining == 0) p.status = Status.Consumed;
            totalIn += used;
        }
    }

    /// @dev Creates the outputs and enforces mass balance against what was used.
    function _createOutputs(Output[] calldata outputs, bytes32 commodity, bytes32 holder, bytes32 stage, uint256 totalIn)
        private
        returns (bytes32[] memory ids)
    {
        uint256 limit = maxYield[commodity][stage];
        if (limit == 0) revert YieldNotSet(commodity, stage);

        uint256 totalOut;
        ids = new bytes32[](outputs.length);
        for (uint256 i; i < outputs.length; ++i) {
            if (outputs[i].quantity == 0) revert ZeroValue();
            totalOut += outputs[i].quantity;
            ids[i] = outputs[i].id;
            _create(outputs[i].id, commodity, holder, outputs[i].quantity, outputs[i].dataHash);
        }

        // Compared without dividing, so rounding can never favour the caller.
        if (totalOut * YIELD_SCALE > totalIn * limit) {
            revert MassBalanceExceeded(totalOut, (totalIn * limit) / YIELD_SCALE);
        }
    }

    /// @notice Close a passport that has reached its final buyer.
    function retire(bytes32 passportId) external onlyRelayer {
        _active(passportId).status = Status.Retired;
        emit PassportRetired(passportId);
    }

    /// @notice Revoke a passport for good, whatever its state.
    function revoke(bytes32 passportId, bytes32 reasonHash) external onlyOperator {
        Passport storage p = _known(passportId);
        if (p.status == Status.Revoked) revert AlreadyRevoked(passportId);
        p.status = Status.Revoked;
        emit Revoked(passportId, reasonHash);
    }

    /// @notice Record that a concern was raised about a passport.
    function flag(bytes32 passportId, bytes32 reasonHash) external onlyRelayer {
        _known(passportId);
        emit Flagged(passportId, reasonHash);
    }

    /// @notice Hand an active passport to another organisation. Only the
    ///         CustodyManager may call this, after checking both signatures.
    function transferCustody(bytes32 passportId, bytes32 newHolder) external {
        if (msg.sender != custodian) revert NotCustodian();
        _active(passportId).holder = newHolder;
    }

    function passport(bytes32 passportId) external view returns (Passport memory) {
        return _known(passportId);
    }

    function _create(bytes32 passportId, bytes32 commodity, bytes32 holder, uint256 quantity, bytes32 dataHash) private {
        if (passportId == bytes32(0) || dataHash == bytes32(0)) revert ZeroValue();
        if (_passports[passportId].status != Status.None) revert PassportExists(passportId);
        _passports[passportId] = Passport(commodity, holder, quantity, quantity, dataHash, Status.Active);
    }

    function _known(bytes32 passportId) private view returns (Passport storage p) {
        p = _passports[passportId];
        if (p.status == Status.None) revert PassportUnknown(passportId);
    }

    function _active(bytes32 passportId) private view returns (Passport storage p) {
        p = _known(passportId);
        if (p.status != Status.Active) revert PassportNotActive(passportId);
    }
}
