// SPDX-License-Identifier: MIT
pragma solidity ^0.8.28;

import {Managed} from "./Managed.sol";
import {OrgRegistry} from "./OrgRegistry.sol";

/// @title AttestationRegistry
/// @notice Certificates attached to a passport or a facility by accredited
///         issuers. Only the document's hash is stored; a certificate altered
///         after issue no longer matches it.
contract AttestationRegistry is Managed {
    struct Attestation {
        bytes32 issuerOrg;
        bytes32 subject; // passport or facility id
        bytes32 docHash;
        uint64 validFrom;
        uint64 validTo;
        bool revoked;
    }

    mapping(bytes32 attestationId => Attestation) private _attestations;

    event AttestationIssued(
        bytes32 indexed attestationId, bytes32 indexed issuerOrg, bytes32 indexed subject, bytes32 docHash, uint64 validFrom, uint64 validTo
    );
    event Revoked(bytes32 indexed attestationId);

    error ZeroValue();
    error NotAccredited(bytes32 orgId);
    error AttestationExists(bytes32 attestationId);
    error AttestationUnknown(bytes32 attestationId);
    error BadValidity(uint64 validFrom, uint64 validTo);
    error AlreadyRevoked(bytes32 attestationId);

    constructor(OrgRegistry orgs_) Managed(orgs_) {}

    function issue(bytes32 attestationId, bytes32 issuerOrg, bytes32 subject, bytes32 docHash, uint64 validFrom, uint64 validTo)
        external
        onlyRelayer
    {
        if (attestationId == bytes32(0) || subject == bytes32(0) || docHash == bytes32(0)) revert ZeroValue();
        if (_attestations[attestationId].issuerOrg != bytes32(0)) revert AttestationExists(attestationId);
        if (!orgs.isAccredited(issuerOrg)) revert NotAccredited(issuerOrg);
        if (validTo <= validFrom) revert BadValidity(validFrom, validTo);

        _attestations[attestationId] = Attestation(issuerOrg, subject, docHash, validFrom, validTo, false);
        emit AttestationIssued(attestationId, issuerOrg, subject, docHash, validFrom, validTo);
    }

    /// @notice Revocation is permanent.
    function revoke(bytes32 attestationId) external onlyRelayer {
        Attestation storage a = _known(attestationId);
        if (a.revoked) revert AlreadyRevoked(attestationId);
        a.revoked = true;
        emit Revoked(attestationId);
    }

    /// @notice True while the attestation is unrevoked and inside its validity window.
    function isValid(bytes32 attestationId) external view returns (bool) {
        Attestation storage a = _attestations[attestationId];
        return a.issuerOrg != bytes32(0) && !a.revoked && block.timestamp >= a.validFrom && block.timestamp <= a.validTo;
    }

    function attestation(bytes32 attestationId) external view returns (Attestation memory) {
        return _known(attestationId);
    }

    function _known(bytes32 attestationId) private view returns (Attestation storage a) {
        a = _attestations[attestationId];
        if (a.issuerOrg == bytes32(0)) revert AttestationUnknown(attestationId);
    }
}
