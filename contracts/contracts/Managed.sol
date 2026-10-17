// SPDX-License-Identifier: MIT
pragma solidity ^0.8.28;

import {OrgRegistry} from "./OrgRegistry.sol";

/// @notice Shared access rules: who the admin, operator and relayer are is
///         decided in one place, the OrgRegistry.
abstract contract Managed {
    OrgRegistry public immutable orgs;

    error NotRelayer();
    error NotAdmin();
    error NotOperator();

    constructor(OrgRegistry orgs_) {
        orgs = orgs_;
    }

    modifier onlyRelayer() {
        if (!orgs.hasRole(orgs.RELAYER_ROLE(), msg.sender)) revert NotRelayer();
        _;
    }

    modifier onlyOperator() {
        if (!orgs.hasRole(orgs.OPERATOR_ROLE(), msg.sender)) revert NotOperator();
        _;
    }

    modifier onlyAdmin() {
        if (!orgs.hasRole(orgs.DEFAULT_ADMIN_ROLE(), msg.sender)) revert NotAdmin();
        _;
    }
}
