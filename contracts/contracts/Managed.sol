// SPDX-License-Identifier: MIT
pragma solidity ^0.8.28;

import {OrgRegistry} from "./OrgRegistry.sol";

/// @notice Shared access rules: who the admin and relayer are is decided in
///         one place, the OrgRegistry.
abstract contract Managed {
    OrgRegistry public immutable orgs;

    error NotRelayer();
    error NotAdmin();

    constructor(OrgRegistry orgs_) {
        orgs = orgs_;
    }

    modifier onlyRelayer() {
        if (!orgs.hasRole(orgs.RELAYER_ROLE(), msg.sender)) revert NotRelayer();
        _;
    }

    modifier onlyAdmin() {
        if (!orgs.hasRole(orgs.DEFAULT_ADMIN_ROLE(), msg.sender)) revert NotAdmin();
        _;
    }
}
