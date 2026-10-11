// SPDX-License-Identifier: MIT
pragma solidity ^0.8.28;

import {Managed} from "./Managed.sol";
import {OrgRegistry} from "./OrgRegistry.sol";

/// @title FacilityRegistry
/// @notice Origin facilities and how much each may originate per period.
///         This is what stops goods from nowhere being "mixed in": a passport
///         can only be issued against capacity that is still free.
contract FacilityRegistry is Managed {
    struct Facility {
        bytes32 orgId;
        bytes32 dataHash; // hash of the off-chain record (location, licence)
        uint256 capacityPerPeriod; // in the commodity's smallest unit
        bool suspended;
    }

    mapping(bytes32 facilityId => Facility) private _facilities;
    /// @notice Quantity already originated by a facility in a period.
    mapping(bytes32 facilityId => mapping(uint256 period => uint256)) public originatedInPeriod;
    /// @notice The only contract allowed to consume capacity (the PassportRegistry).
    address public originator;

    event FacilityRegistered(bytes32 indexed facilityId, bytes32 indexed orgId, bytes32 dataHash, uint256 capacityPerPeriod);
    event FacilitySuspended(bytes32 indexed facilityId, bool suspended);
    event OriginatorSet(address originator);

    error ZeroValue();
    error OrgNotVerified(bytes32 orgId);
    error FacilityExists(bytes32 facilityId);
    error FacilityUnknown(bytes32 facilityId);
    error FacilityIsSuspended(bytes32 facilityId);
    error CapacityExceeded(bytes32 facilityId, uint256 period, uint256 requested, uint256 available);
    error OriginatorAlreadySet();
    error NotOriginator();

    constructor(OrgRegistry orgs_) Managed(orgs_) {}

    /// @notice One-time wiring: name the PassportRegistry.
    function setOriginator(address originator_) external onlyAdmin {
        if (originator != address(0)) revert OriginatorAlreadySet();
        if (originator_ == address(0)) revert ZeroValue();
        originator = originator_;
        emit OriginatorSet(originator_);
    }

    function register(bytes32 facilityId, bytes32 orgId, bytes32 dataHash, uint256 capacityPerPeriod)
        external
        onlyRelayer
    {
        if (facilityId == bytes32(0) || dataHash == bytes32(0) || capacityPerPeriod == 0) revert ZeroValue();
        if (_facilities[facilityId].orgId != bytes32(0)) revert FacilityExists(facilityId);
        if (!orgs.isVerified(orgId)) revert OrgNotVerified(orgId);

        _facilities[facilityId] = Facility(orgId, dataHash, capacityPerPeriod, false);
        emit FacilityRegistered(facilityId, orgId, dataHash, capacityPerPeriod);
    }

    function setSuspended(bytes32 facilityId, bool suspended) external onlyAdmin {
        _known(facilityId).suspended = suspended;
        emit FacilitySuspended(facilityId, suspended);
    }

    /// @notice Take `quantity` from the facility's free capacity for `period`
    ///         and return the organisation that owns the facility.
    /// @dev Reverts if the facility is suspended, its organisation is no longer
    ///      verified, or the capacity would be exceeded.
    function consumeCapacity(bytes32 facilityId, uint256 period, uint256 quantity) external returns (bytes32 orgId) {
        if (msg.sender != originator) revert NotOriginator();
        Facility storage f = _known(facilityId);
        if (f.suspended) revert FacilityIsSuspended(facilityId);
        if (!orgs.isVerified(f.orgId)) revert OrgNotVerified(f.orgId);

        uint256 used = originatedInPeriod[facilityId][period];
        uint256 available = f.capacityPerPeriod - used;
        if (quantity > available) revert CapacityExceeded(facilityId, period, quantity, available);

        originatedInPeriod[facilityId][period] = used + quantity;
        return f.orgId;
    }

    function facility(bytes32 facilityId) external view returns (Facility memory) {
        return _known(facilityId);
    }

    function _known(bytes32 facilityId) private view returns (Facility storage f) {
        f = _facilities[facilityId];
        if (f.orgId == bytes32(0)) revert FacilityUnknown(facilityId);
    }
}
