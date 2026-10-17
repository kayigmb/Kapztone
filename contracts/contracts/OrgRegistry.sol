// SPDX-License-Identifier: MIT
pragma solidity ^0.8.28;

import {AccessControl} from "@openzeppelin/contracts/access/AccessControl.sol";

/// @title OrgRegistry
/// @notice Verified organisations, their signing addresses and accredited
///         issuers. It is also the platform's role book: every other contract
///         asks it who the admin and the relayer are.
/// @dev Organisations never send transactions. The relayer sends them; an
///      organisation is known by the address of the key the platform holds
///      for it, and proves consent by signature (see CustodyManager).
contract OrgRegistry is AccessControl {
    /// @notice May submit transactions on behalf of organisations.
    bytes32 public constant RELAYER_ROLE = keccak256("RELAYER_ROLE");
    /// @notice Routine platform decisions: verify organisations, accredit
    ///         issuers, set yields, suspend facilities, revoke passports.
    ///         The admin only grants roles and does one-time wiring, so its
    ///         key can stay offline.
    bytes32 public constant OPERATOR_ROLE = keccak256("OPERATOR_ROLE");

    struct Org {
        address signer;
        bool verified;
    }

    /// @notice When an issuer may issue one type of certificate.
    struct Window {
        uint64 validFrom;
        uint64 validTo;
    }

    mapping(bytes32 orgId => Org) private _orgs;
    mapping(bytes32 orgId => mapping(bytes32 certType => Window)) private _accreditations;
    mapping(address signer => bytes32 orgId) public orgOfSigner;

    event OrgRegistered(bytes32 indexed orgId, address indexed signer);
    event OrgSignerChanged(bytes32 indexed orgId, address indexed signer);
    event OrgVerified(bytes32 indexed orgId, bool verified);
    event IssuerAccredited(bytes32 indexed orgId, bytes32 indexed certType, uint64 validFrom, uint64 validTo);

    error ZeroValue();
    error OrgExists(bytes32 orgId);
    error OrgUnknown(bytes32 orgId);
    error SignerInUse(address signer);
    error BadValidity(uint64 validFrom, uint64 validTo);

    constructor(address admin, address relayer) {
        if (admin == address(0) || relayer == address(0)) revert ZeroValue();
        _grantRole(DEFAULT_ADMIN_ROLE, admin);
        _grantRole(RELAYER_ROLE, relayer);
        // The relayer process sends the routine decisions too.
        _grantRole(OPERATOR_ROLE, relayer);
    }

    /// @notice Record an organisation and the address of its signing key.
    ///         It starts unverified.
    function register(bytes32 orgId, address signer) external onlyRole(RELAYER_ROLE) {
        if (orgId == bytes32(0)) revert ZeroValue();
        if (_orgs[orgId].signer != address(0)) revert OrgExists(orgId);
        _bindSigner(orgId, signer);
        emit OrgRegistered(orgId, signer);
    }

    /// @notice Replace an organisation's signing key, e.g. after a key rotation.
    function setSigner(bytes32 orgId, address signer) external onlyRole(OPERATOR_ROLE) {
        address old = _known(orgId).signer;
        delete orgOfSigner[old];
        _bindSigner(orgId, signer);
        emit OrgSignerChanged(orgId, signer);
    }

    /// @notice Mark an organisation as having passed (or lost) KYB.
    function setVerified(bytes32 orgId, bool verified) external onlyRole(OPERATOR_ROLE) {
        _known(orgId).verified = verified;
        emit OrgVerified(orgId, verified);
    }

    /// @notice Allow an organisation to issue one type of certificate inside a
    ///         window, or stop it (validTo 0).
    function setAccreditation(bytes32 orgId, bytes32 certType, uint64 validFrom, uint64 validTo)
        external
        onlyRole(OPERATOR_ROLE)
    {
        _known(orgId);
        if (certType == bytes32(0)) revert ZeroValue();
        if (validTo == 0) {
            delete _accreditations[orgId][certType];
        } else {
            if (validTo <= validFrom) revert BadValidity(validFrom, validTo);
            _accreditations[orgId][certType] = Window(validFrom, validTo);
        }
        emit IssuerAccredited(orgId, certType, validFrom, validTo);
    }

    function isVerified(bytes32 orgId) external view returns (bool) {
        return _orgs[orgId].verified;
    }

    /// @notice An issuer must be verified and inside its window for this type.
    function isAccredited(bytes32 orgId, bytes32 certType) external view returns (bool) {
        Window storage w = _accreditations[orgId][certType];
        return _orgs[orgId].verified && block.timestamp >= w.validFrom && block.timestamp <= w.validTo;
    }

    /// @notice The address whose signature counts as this organisation's; zero if unknown.
    function signerOf(bytes32 orgId) external view returns (address) {
        return _orgs[orgId].signer;
    }

    function _known(bytes32 orgId) private view returns (Org storage o) {
        o = _orgs[orgId];
        if (o.signer == address(0)) revert OrgUnknown(orgId);
    }

    function _bindSigner(bytes32 orgId, address signer) private {
        if (signer == address(0)) revert ZeroValue();
        if (orgOfSigner[signer] != bytes32(0)) revert SignerInUse(signer);
        _orgs[orgId].signer = signer;
        orgOfSigner[signer] = orgId;
    }
}
