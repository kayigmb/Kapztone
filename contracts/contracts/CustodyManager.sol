// SPDX-License-Identifier: MIT
pragma solidity ^0.8.28;

import {EIP712} from "@openzeppelin/contracts/utils/cryptography/EIP712.sol";
import {ECDSA} from "@openzeppelin/contracts/utils/cryptography/ECDSA.sol";
import {Managed} from "./Managed.sol";
import {OrgRegistry} from "./OrgRegistry.sol";
import {PassportRegistry} from "./PassportRegistry.sol";

/// @title CustodyManager
/// @notice Custody handovers. A handover counts only when the sender and the
///         receiver have both signed the same message; the relayer cannot
///         move a passport on its own.
contract CustodyManager is Managed, EIP712 {
    struct Handover {
        bytes32 passportId;
        bytes32 fromOrg;
        bytes32 toOrg;
        uint256 quantity; // what both parties agree is being handed over
        bytes32 dataHash; // hash of the off-chain record (location, photos)
        uint256 nonce; // the passport's handover count; makes each message single-use
        uint256 deadline; // unix time after which the signatures are void
    }

    bytes32 public constant HANDOVER_TYPEHASH = keccak256(
        "Handover(bytes32 passportId,bytes32 fromOrg,bytes32 toOrg,uint256 quantity,bytes32 dataHash,uint256 nonce,uint256 deadline)"
    );

    PassportRegistry public immutable passports;
    /// @notice Number of handovers recorded for a passport; the next message must carry this value.
    mapping(bytes32 passportId => uint256) public nonces;

    event CustodyTransferred(
        bytes32 indexed passportId, bytes32 indexed fromOrg, bytes32 indexed toOrg, uint256 quantity, bytes32 dataHash
    );
    event TraceEvent(bytes32 indexed passportId, bytes32 indexed holder, bytes32 indexed kind, bytes32 dataHash);

    error ZeroValue();
    error Expired(uint256 deadline);
    error BadNonce(uint256 given, uint256 expected);
    error NotHolder(bytes32 passportId, bytes32 fromOrg);
    error SameOrg();
    error OrgNotVerified(bytes32 orgId);
    error QuantityMismatch(uint256 given, uint256 held);
    error BadSignature(bytes32 orgId);
    error PassportNotActive(bytes32 passportId);

    constructor(OrgRegistry orgs_, PassportRegistry passports_) Managed(orgs_) EIP712("ChainTrack", "1") {
        passports = passports_;
    }

    /// @notice Move a passport from its holder to the receiver, given both signatures.
    function recordHandover(Handover calldata h, bytes calldata sigFrom, bytes calldata sigTo) external onlyRelayer {
        if (block.timestamp > h.deadline) revert Expired(h.deadline);
        uint256 expected = nonces[h.passportId];
        if (h.nonce != expected) revert BadNonce(h.nonce, expected);
        if (h.fromOrg == h.toOrg) revert SameOrg();
        if (h.dataHash == bytes32(0)) revert ZeroValue();

        PassportRegistry.Passport memory p = passports.passport(h.passportId);
        if (p.holder != h.fromOrg) revert NotHolder(h.passportId, h.fromOrg);
        if (h.quantity != p.remaining) revert QuantityMismatch(h.quantity, p.remaining);
        if (!orgs.isVerified(h.fromOrg)) revert OrgNotVerified(h.fromOrg);
        if (!orgs.isVerified(h.toOrg)) revert OrgNotVerified(h.toOrg);

        bytes32 digest = hashHandover(h);
        _requireSignedBy(h.fromOrg, digest, sigFrom);
        _requireSignedBy(h.toOrg, digest, sigTo);

        nonces[h.passportId] = expected + 1;
        passports.transferCustody(h.passportId, h.toOrg); // reverts unless the passport is active
        emit CustodyTransferred(h.passportId, h.fromOrg, h.toOrg, h.quantity, h.dataHash);
    }

    /// @notice Record a journey event that does not change the holder (export, import, inspection).
    function recordEvent(bytes32 passportId, bytes32 kind, bytes32 dataHash) external onlyRelayer {
        if (kind == bytes32(0) || dataHash == bytes32(0)) revert ZeroValue();
        PassportRegistry.Passport memory p = passports.passport(passportId);
        if (p.status != PassportRegistry.Status.Active) revert PassportNotActive(passportId);
        emit TraceEvent(passportId, p.holder, kind, dataHash);
    }

    /// @notice The exact value each party signs for a handover.
    function hashHandover(Handover calldata h) public view returns (bytes32) {
        return _hashTypedDataV4(
            keccak256(
                abi.encode(HANDOVER_TYPEHASH, h.passportId, h.fromOrg, h.toOrg, h.quantity, h.dataHash, h.nonce, h.deadline)
            )
        );
    }

    function _requireSignedBy(bytes32 orgId, bytes32 digest, bytes calldata signature) private view {
        address signer = orgs.signerOf(orgId);
        (address recovered, ECDSA.RecoverError err,) = ECDSA.tryRecover(digest, signature);
        if (err != ECDSA.RecoverError.NoError || signer == address(0) || recovered != signer) {
            revert BadSignature(orgId);
        }
    }
}
