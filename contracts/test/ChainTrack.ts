import { readFileSync } from "node:fs";
import { expect } from "chai";
import { network } from "hardhat";

const { ethers, networkHelpers } = await network.create();

// Ids and hashes on-chain are 32-byte values; tests derive them from names.
const id = (name: string) => ethers.id(name);
const ZERO = ethers.ZeroHash;

const COFFEE = id("coffee");
const WASHING = id("washing_station");
const SPLIT = id("split");
const PPM = 1_000_000n;
const WASHING_YIELD = 200_000n; // 0.2
const PERIOD = 2026n;
const CAPACITY = 1_000n;
// Certificate types are template evidence keys, hashed.
const LAB = id("lab_report");
const ORIGIN_CERT = id("certificate_of_origin");
const FOREVER = 2n ** 64n - 1n;

// The handover message is declared once, in a file the Go tests read too.
const vector = JSON.parse(readFileSync(new URL("./vectors/handover.json", import.meta.url), "utf8"));
const HANDOVER_TYPES = vector.types;

async function deploy() {
  const [admin, relayer, stranger, farmKey, exporterKey, certifierKey, spareKey] = await ethers.getSigners();

  const orgs = await ethers.deployContract("OrgRegistry", [admin.address, relayer.address]);
  const facilities = await ethers.deployContract("FacilityRegistry", [orgs.target]);
  const passports = await ethers.deployContract("PassportRegistry", [orgs.target, facilities.target]);
  const custody = await ethers.deployContract("CustodyManager", [orgs.target, passports.target]);
  const attestations = await ethers.deployContract("AttestationRegistry", [orgs.target]);
  await facilities.setOriginator(passports.target);
  await passports.setCustodian(custody.target);

  // Three organisations: a farm and an exporter (verified), a certifier (verified + accredited).
  const FARM = id("org:farm"), EXPORTER = id("org:exporter"), CERTIFIER = id("org:certifier");
  const asRelayer = { orgs: orgs.connect(relayer), facilities: facilities.connect(relayer), passports: passports.connect(relayer), custody: custody.connect(relayer), attestations: attestations.connect(relayer) };
  await asRelayer.orgs.register(FARM, farmKey.address);
  await asRelayer.orgs.register(EXPORTER, exporterKey.address);
  await asRelayer.orgs.register(CERTIFIER, certifierKey.address);
  for (const org of [FARM, EXPORTER, CERTIFIER]) await asRelayer.orgs.setVerified(org, true);
  await asRelayer.orgs.setAccreditation(CERTIFIER, LAB, 0n, FOREVER);

  const PLOT = id("facility:plot");
  await asRelayer.facilities.register(PLOT, FARM, id("plot record"), CAPACITY);
  await asRelayer.passports.setMaxYield(COFFEE, WASHING, WASHING_YIELD);
  await asRelayer.passports.setMaxYield(COFFEE, SPLIT, PPM); // splitting loses nothing

  const domain = {
    name: "ChainTrack",
    version: "1",
    chainId: (await ethers.provider.getNetwork()).chainId,
    verifyingContract: await custody.getAddress(),
  };

  /** A handover of `passportId`, valid for an hour unless overridden. */
  const handover = async (passportId: string, quantity: bigint, overrides: Record<string, unknown> = {}) => ({
    passportId,
    fromOrg: FARM,
    toOrg: EXPORTER,
    quantity,
    dataHash: id("handover record"),
    nonce: await custody.nonces(passportId),
    deadline: BigInt(await networkHelpers.time.latest()) + 3600n,
    ...overrides,
  });
  const sign = (key: typeof farmKey, h: object) => key.signTypedData(domain, HANDOVER_TYPES, h);

  /** Issues `quantity` of coffee from the plot and returns the passport id. */
  const issue = async (name: string, quantity: bigint, period = PERIOD) => {
    await asRelayer.passports.issue(id(name), COFFEE, PLOT, period, quantity, id(name + " record"));
    return id(name);
  };

  return {
    admin, relayer, stranger, farmKey, exporterKey, certifierKey, spareKey,
    orgs, facilities, passports, custody, attestations, asRelayer,
    FARM, EXPORTER, CERTIFIER, PLOT, domain, handover, sign, issue,
  };
}

const setup = () => networkHelpers.loadFixture(deploy);

describe("OrgRegistry", function () {
  it("registers an organisation unverified, then admin verifies and accredits it", async function () {
    const { orgs, asRelayer, spareKey } = await setup();
    const NEW = id("org:new");

    await expect(asRelayer.orgs.register(NEW, spareKey.address)).to.emit(orgs, "OrgRegistered").withArgs(NEW, spareKey.address);
    expect(await orgs.isVerified(NEW)).to.equal(false);
    expect(await orgs.signerOf(NEW)).to.equal(spareKey.address);
    expect(await orgs.orgOfSigner(spareKey.address)).to.equal(NEW);

    await expect(asRelayer.orgs.setVerified(NEW, true)).to.emit(orgs, "OrgVerified").withArgs(NEW, true);
    expect(await orgs.isVerified(NEW)).to.equal(true);
    // Accreditation is per certificate type, and only counts for a verified organisation.
    await expect(asRelayer.orgs.setAccreditation(NEW, LAB, 0n, FOREVER)).to.emit(orgs, "IssuerAccredited").withArgs(NEW, LAB, 0n, FOREVER);
    expect(await orgs.isAccredited(NEW, LAB)).to.equal(true);
    expect(await orgs.isAccredited(NEW, ORIGIN_CERT)).to.equal(false);
    await asRelayer.orgs.setVerified(NEW, false);
    expect(await orgs.isAccredited(NEW, LAB)).to.equal(false);
  });

  it("accreditation counts only inside its window, and validTo 0 removes it", async function () {
    const { orgs, asRelayer, CERTIFIER } = await setup();
    const now = BigInt(await networkHelpers.time.latest());

    await asRelayer.orgs.setAccreditation(CERTIFIER, ORIGIN_CERT, now + 100n, now + 200n);
    expect(await orgs.isAccredited(CERTIFIER, ORIGIN_CERT)).to.equal(false); // not started
    await networkHelpers.time.increaseTo(now + 150n);
    expect(await orgs.isAccredited(CERTIFIER, ORIGIN_CERT)).to.equal(true);
    await networkHelpers.time.increaseTo(now + 201n);
    expect(await orgs.isAccredited(CERTIFIER, ORIGIN_CERT)).to.equal(false); // ended

    expect(await orgs.isAccredited(CERTIFIER, LAB)).to.equal(true);
    await expect(asRelayer.orgs.setAccreditation(CERTIFIER, LAB, 0n, 0n)).to.emit(orgs, "IssuerAccredited").withArgs(CERTIFIER, LAB, 0n, 0n);
    expect(await orgs.isAccredited(CERTIFIER, LAB)).to.equal(false);
  });

  it("refuses a bad accreditation window or an empty type", async function () {
    const { orgs, asRelayer, CERTIFIER } = await setup();
    await expect(asRelayer.orgs.setAccreditation(CERTIFIER, LAB, 10n, 10n)).to.be.revertedWithCustomError(orgs, "BadValidity").withArgs(10n, 10n);
    await expect(asRelayer.orgs.setAccreditation(CERTIFIER, LAB, 10n, 5n)).to.be.revertedWithCustomError(orgs, "BadValidity");
    await expect(asRelayer.orgs.setAccreditation(CERTIFIER, ZERO, 0n, FOREVER)).to.be.revertedWithCustomError(orgs, "ZeroValue");
  });

  it("only the operator verifies or accredits; only the admin grants roles", async function () {
    const { orgs, admin, relayer, stranger, FARM } = await setup();
    const OPERATOR = await orgs.OPERATOR_ROLE();
    expect(await orgs.hasRole(OPERATOR, relayer.address)).to.equal(true);
    expect(await orgs.hasRole(OPERATOR, admin.address)).to.equal(false);
    // The operator cannot hand out roles.
    await expect(orgs.connect(relayer).grantRole(OPERATOR, stranger.address)).to.be.revertedWithCustomError(orgs, "AccessControlUnauthorizedAccount");
    // The admin can, and the new operator can then act.
    await orgs.grantRole(OPERATOR, stranger.address);
    await expect(orgs.connect(stranger).setVerified(FARM, false)).to.emit(orgs, "OrgVerified").withArgs(FARM, false);
    await orgs.revokeRole(OPERATOR, stranger.address);
    await expect(orgs.connect(stranger).setVerified(FARM, true)).to.be.revertedWithCustomError(orgs, "AccessControlUnauthorizedAccount");
  });

  it("the admin cannot verify or accredit; only the relayer can register", async function () {
    const { orgs, asRelayer, stranger, spareKey, FARM } = await setup();
    await expect(orgs.setVerified(FARM, false)).to.be.revertedWithCustomError(orgs, "AccessControlUnauthorizedAccount");
    await expect(orgs.setAccreditation(FARM, LAB, 0n, FOREVER)).to.be.revertedWithCustomError(orgs, "AccessControlUnauthorizedAccount");
    await expect(orgs.connect(stranger).register(id("x"), spareKey.address)).to.be.revertedWithCustomError(orgs, "AccessControlUnauthorizedAccount");
    await expect(orgs.register(id("x"), spareKey.address)).to.be.revertedWithCustomError(orgs, "AccessControlUnauthorizedAccount");
  });

  it("rejects duplicates, unknown organisations and empty values", async function () {
    const { orgs, asRelayer, farmKey, spareKey, FARM } = await setup();
    await expect(asRelayer.orgs.register(FARM, spareKey.address)).to.be.revertedWithCustomError(orgs, "OrgExists");
    await expect(asRelayer.orgs.register(id("x"), farmKey.address)).to.be.revertedWithCustomError(orgs, "SignerInUse");
    await expect(asRelayer.orgs.register(ZERO, spareKey.address)).to.be.revertedWithCustomError(orgs, "ZeroValue");
    await expect(asRelayer.orgs.register(id("x"), ethers.ZeroAddress)).to.be.revertedWithCustomError(orgs, "ZeroValue");
    await expect(asRelayer.orgs.setVerified(id("ghost"), true)).to.be.revertedWithCustomError(orgs, "OrgUnknown");
    await expect(asRelayer.orgs.setAccreditation(id("ghost"), LAB, 0n, FOREVER)).to.be.revertedWithCustomError(orgs, "OrgUnknown");
    await expect(ethers.deployContract("OrgRegistry", [ethers.ZeroAddress, spareKey.address])).to.be.revertedWithCustomError(orgs, "ZeroValue");
  });

  it("admin can rotate a signing key; the old key stops counting", async function () {
    const { orgs, asRelayer, farmKey, spareKey, exporterKey, FARM } = await setup();
    await expect(asRelayer.orgs.setSigner(FARM, spareKey.address)).to.emit(orgs, "OrgSignerChanged").withArgs(FARM, spareKey.address);
    expect(await orgs.signerOf(FARM)).to.equal(spareKey.address);
    expect(await orgs.orgOfSigner(farmKey.address)).to.equal(ZERO);

    await expect(asRelayer.orgs.setSigner(FARM, exporterKey.address)).to.be.revertedWithCustomError(orgs, "SignerInUse");
    await expect(asRelayer.orgs.setSigner(id("ghost"), farmKey.address)).to.be.revertedWithCustomError(orgs, "OrgUnknown");
    await expect(orgs.setSigner(FARM, farmKey.address)).to.be.revertedWithCustomError(orgs, "AccessControlUnauthorizedAccount");
  });
});

describe("FacilityRegistry", function () {
  it("registers a facility for a verified organisation", async function () {
    const { facilities, asRelayer, EXPORTER } = await setup();
    const SITE = id("facility:site");
    await expect(asRelayer.facilities.register(SITE, EXPORTER, id("site record"), 50n))
      .to.emit(facilities, "FacilityRegistered").withArgs(SITE, EXPORTER, id("site record"), 50n);
    const f = await facilities.facility(SITE);
    expect([f.orgId, f.capacityPerPeriod, f.suspended]).to.deep.equal([EXPORTER, 50n, false]);
  });

  it("rejects an unverified organisation, duplicates, empty values and non-relayers", async function () {
    const { facilities, orgs, asRelayer, spareKey, stranger, FARM, PLOT } = await setup();
    await asRelayer.orgs.register(id("org:pending"), spareKey.address);
    await expect(asRelayer.facilities.register(id("f"), id("org:pending"), id("r"), 1n)).to.be.revertedWithCustomError(facilities, "OrgNotVerified");
    await expect(asRelayer.facilities.register(PLOT, FARM, id("r"), 1n)).to.be.revertedWithCustomError(facilities, "FacilityExists");
    await expect(asRelayer.facilities.register(id("f"), FARM, id("r"), 0n)).to.be.revertedWithCustomError(facilities, "ZeroValue");
    await expect(asRelayer.facilities.register(ZERO, FARM, id("r"), 1n)).to.be.revertedWithCustomError(facilities, "ZeroValue");
    await expect(asRelayer.facilities.register(id("f"), FARM, ZERO, 1n)).to.be.revertedWithCustomError(facilities, "ZeroValue");
    await expect(facilities.connect(stranger).register(id("f"), FARM, id("r"), 1n)).to.be.revertedWithCustomError(facilities, "NotRelayer");
    await expect(facilities.facility(id("ghost"))).to.be.revertedWithCustomError(facilities, "FacilityUnknown");
    expect(await orgs.isVerified(FARM)).to.equal(true);
  });

  it("only admin can suspend, and only once is the originator wired", async function () {
    const { facilities, asRelayer, stranger, PLOT } = await setup();
    await expect(facilities.setSuspended(PLOT, true)).to.be.revertedWithCustomError(facilities, "NotOperator");
    await expect(asRelayer.facilities.setSuspended(id("ghost"), true)).to.be.revertedWithCustomError(facilities, "FacilityUnknown");
    await expect(asRelayer.facilities.setSuspended(PLOT, true)).to.emit(facilities, "FacilitySuspended").withArgs(PLOT, true);

    await expect(facilities.setOriginator(stranger.address)).to.be.revertedWithCustomError(facilities, "OriginatorAlreadySet");
    await expect(asRelayer.facilities.setOriginator(stranger.address)).to.be.revertedWithCustomError(facilities, "NotAdmin");
    // Nobody but the PassportRegistry can use up capacity.
    await expect(facilities.consumeCapacity(PLOT, PERIOD, 1n)).to.be.revertedWithCustomError(facilities, "NotOriginator");

    const fresh = await ethers.deployContract("FacilityRegistry", [await facilities.orgs()]);
    await expect(fresh.setOriginator(ethers.ZeroAddress)).to.be.revertedWithCustomError(fresh, "ZeroValue");
  });
});

describe("PassportRegistry: issue", function () {
  it("issues a passport to the facility's organisation and counts it against capacity", async function () {
    const { passports, facilities, asRelayer, FARM, PLOT } = await setup();
    const P = id("lot-1");
    await expect(asRelayer.passports.issue(P, COFFEE, PLOT, PERIOD, 400n, id("lot-1 record")))
      .to.emit(passports, "PassportIssued").withArgs(P, FARM, PLOT, COFFEE, PERIOD, 400n, id("lot-1 record"));

    const p = await passports.passport(P);
    expect([p.commodity, p.holder, p.quantity, p.remaining, p.dataHash, p.status]).to.deep.equal([COFFEE, FARM, 400n, 400n, id("lot-1 record"), 1n]);
    expect(await facilities.originatedInPeriod(PLOT, PERIOD)).to.equal(400n);
  });

  it("blocks issuing beyond the facility's capacity for the period", async function () {
    const { facilities, issue, asRelayer, PLOT } = await setup();
    await issue("lot-1", 600n);
    await issue("lot-2", 400n); // exactly at capacity
    await expect(asRelayer.passports.issue(id("lot-3"), COFFEE, PLOT, PERIOD, 1n, id("r")))
      .to.be.revertedWithCustomError(facilities, "CapacityExceeded").withArgs(PLOT, PERIOD, 1n, 0n);
    // A new period has its full capacity again.
    await issue("lot-4", CAPACITY, PERIOD + 1n);
  });

  it("blocks a suspended facility, an unverified owner and an unknown facility", async function () {
    const { facilities, orgs, asRelayer, PLOT, FARM } = await setup();
    const tryIssue = (name: string, facility = PLOT) => asRelayer.passports.issue(id(name), COFFEE, facility, PERIOD, 1n, id("r"));

    await asRelayer.facilities.setSuspended(PLOT, true);
    await expect(tryIssue("a")).to.be.revertedWithCustomError(facilities, "FacilityIsSuspended");
    await asRelayer.facilities.setSuspended(PLOT, false);

    await asRelayer.orgs.setVerified(FARM, false);
    await expect(tryIssue("b")).to.be.revertedWithCustomError(facilities, "OrgNotVerified");
    await asRelayer.orgs.setVerified(FARM, true);

    await expect(tryIssue("c", id("ghost"))).to.be.revertedWithCustomError(facilities, "FacilityUnknown");
  });

  it("rejects a reused id, empty values and non-relayers", async function () {
    const { passports, issue, asRelayer, stranger, PLOT } = await setup();
    const P = await issue("lot-1", 10n);
    await expect(asRelayer.passports.issue(P, COFFEE, PLOT, PERIOD, 10n, id("r"))).to.be.revertedWithCustomError(passports, "PassportExists");
    await expect(asRelayer.passports.issue(id("x"), COFFEE, PLOT, PERIOD, 0n, id("r"))).to.be.revertedWithCustomError(passports, "ZeroValue");
    await expect(asRelayer.passports.issue(id("x"), ZERO, PLOT, PERIOD, 1n, id("r"))).to.be.revertedWithCustomError(passports, "ZeroValue");
    await expect(asRelayer.passports.issue(id("x"), COFFEE, PLOT, PERIOD, 1n, ZERO)).to.be.revertedWithCustomError(passports, "ZeroValue");
    await expect(asRelayer.passports.issue(ZERO, COFFEE, PLOT, PERIOD, 1n, id("r"))).to.be.revertedWithCustomError(passports, "ZeroValue");
    await expect(passports.connect(stranger).issue(id("x"), COFFEE, PLOT, PERIOD, 1n, id("r"))).to.be.revertedWithCustomError(passports, "NotRelayer");
    await expect(passports.passport(id("ghost"))).to.be.revertedWithCustomError(passports, "PassportUnknown");
  });
});

describe("PassportRegistry: transform and mass balance", function () {
  const out = (name: string, quantity: bigint) => ({ id: id(name), quantity, dataHash: id(name + " record") });

  it("processes cherries into parchment within the stage's yield", async function () {
    const { passports, issue, asRelayer, FARM } = await setup();
    const P = await issue("cherries", 1000n);

    await expect(asRelayer.passports.transform([P], [1000n], WASHING, [out("parchment", 200n)], id("wash record")))
      .to.emit(passports, "LotTransformed").withArgs(FARM, WASHING, [P], [1000n], [id("parchment")], id("wash record"));

    const input = await passports.passport(P);
    expect([input.remaining, input.status]).to.deep.equal([0n, 2n]); // Consumed
    const output = await passports.passport(id("parchment"));
    expect([output.holder, output.commodity, output.quantity, output.status]).to.deep.equal([FARM, COFFEE, 200n, 1n]);
  });

  it("blocks outputs above inputs × max yield, by even one unit", async function () {
    const { passports, issue, asRelayer } = await setup();
    const P = await issue("cherries", 1000n);
    await expect(asRelayer.passports.transform([P], [1000n], WASHING, [out("parchment", 201n)], id("r")))
      .to.be.revertedWithCustomError(passports, "MassBalanceExceeded").withArgs(201n, 200n);
    // Rounding never helps the caller: 7 in at 0.2 allows 1 out, not 2.
    await expect(asRelayer.passports.transform([P], [7n], WASHING, [out("p2", 2n)], id("r")))
      .to.be.revertedWithCustomError(passports, "MassBalanceExceeded").withArgs(2n, 1n);
    await asRelayer.passports.transform([P], [7n], WASHING, [out("p1", 1n)], id("r"));
  });

  it("splits one lot and merges several, tracking what remains", async function () {
    const { passports, issue, asRelayer } = await setup();
    const A = await issue("a", 300n), B = await issue("b", 200n);

    // Split: 100 of A into two lots; A keeps 200 and stays active.
    await asRelayer.passports.transform([A], [100n], SPLIT, [out("a1", 60n), out("a2", 40n)], id("r"));
    const a = await passports.passport(A);
    expect([a.remaining, a.status]).to.deep.equal([200n, 1n]);

    // Merge: the rest of A and all of B into one lot.
    await asRelayer.passports.transform([A, B], [200n, 200n], SPLIT, [out("ab", 400n)], id("r"));
    expect((await passports.passport(A)).status).to.equal(2n);
    expect((await passports.passport(B)).status).to.equal(2n);

    // A consumed passport cannot be used again.
    await expect(asRelayer.passports.transform([A], [1n], SPLIT, [out("again", 1n)], id("r")))
      .to.be.revertedWithCustomError(passports, "PassportNotActive");
  });

  it("rejects using more than remains, mixed holders, unset yields and malformed calls", async function () {
    const { passports, orgs, issue, asRelayer, exporterKey, farmKey, handover, sign, stranger, EXPORTER } = await setup();
    const A = await issue("a", 100n), B = await issue("b", 100n);
    const t = asRelayer.passports.transform;

    await expect(t([A], [101n], SPLIT, [out("x", 1n)], id("r"))).to.be.revertedWithCustomError(passports, "InsufficientQuantity").withArgs(A, 101n, 100n);
    await expect(t([A], [0n], SPLIT, [out("x", 1n)], id("r"))).to.be.revertedWithCustomError(passports, "ZeroValue");
    await expect(t([A], [10n], SPLIT, [out("x", 0n)], id("r"))).to.be.revertedWithCustomError(passports, "ZeroValue");
    await expect(t([A], [10n], id("roasting"), [out("x", 1n)], id("r"))).to.be.revertedWithCustomError(passports, "YieldNotSet");
    await expect(t([A, B], [10n], SPLIT, [out("x", 1n)], id("r"))).to.be.revertedWithCustomError(passports, "LengthMismatch");
    await expect(t([], [], SPLIT, [out("x", 1n)], id("r"))).to.be.revertedWithCustomError(passports, "ZeroValue");
    await expect(t([A], [10n], SPLIT, [], id("r"))).to.be.revertedWithCustomError(passports, "ZeroValue");
    await expect(t([id("ghost")], [1n], SPLIT, [out("x", 1n)], id("r"))).to.be.revertedWithCustomError(passports, "PassportUnknown");
    await expect(t([A], [10n], SPLIT, [{ id: B, quantity: 10n, dataHash: id("r") }], id("r"))).to.be.revertedWithCustomError(passports, "PassportExists");
    await expect(passports.connect(stranger).transform([A], [10n], SPLIT, [out("x", 1n)], id("r"))).to.be.revertedWithCustomError(passports, "NotRelayer");

    // Lots held by different organisations cannot be merged.
    const h = await handover(B, 100n);
    await asRelayer.custody.recordHandover(h, await sign(farmKey, h), await sign(exporterKey, h));
    await expect(t([A, B], [10n, 10n], SPLIT, [out("x", 20n)], id("r"))).to.be.revertedWithCustomError(passports, "MixedInputs").withArgs(B);
    expect(await orgs.isVerified(EXPORTER)).to.equal(true);
  });

  it("only admin sets yields", async function () {
    const { passports, asRelayer } = await setup();
    await expect(asRelayer.passports.setMaxYield(COFFEE, id("roasting"), 850_000n)).to.emit(passports, "MaxYieldSet").withArgs(COFFEE, id("roasting"), 850_000n);
    await expect(passports.setMaxYield(COFFEE, WASHING, PPM)).to.be.revertedWithCustomError(passports, "NotOperator");
    await expect(asRelayer.passports.setMaxYield(COFFEE, WASHING, 0n)).to.be.revertedWithCustomError(passports, "ZeroValue");
    await expect(asRelayer.passports.setMaxYield(ZERO, WASHING, 1n)).to.be.revertedWithCustomError(passports, "ZeroValue");
    await expect(asRelayer.passports.setMaxYield(COFFEE, ZERO, 1n)).to.be.revertedWithCustomError(passports, "ZeroValue");
  });
});

describe("PassportRegistry: retire, revoke, flag", function () {
  it("retires an active passport once", async function () {
    const { passports, issue, asRelayer } = await setup();
    const P = await issue("lot", 10n);
    await expect(asRelayer.passports.retire(P)).to.emit(passports, "PassportRetired").withArgs(P);
    await expect(asRelayer.passports.retire(P)).to.be.revertedWithCustomError(passports, "PassportNotActive");
    await expect(passports.retire(P)).to.be.revertedWithCustomError(passports, "NotRelayer");
  });

  it("revocation is permanent and stops every further use", async function () {
    const { passports, issue, asRelayer, farmKey, exporterKey, handover, sign } = await setup();
    const P = await issue("lot", 10n);
    await expect(passports.revoke(P, id("fraud"))).to.be.revertedWithCustomError(passports, "NotOperator");
    await expect(asRelayer.passports.revoke(P, id("fraud"))).to.emit(passports, "Revoked").withArgs(P, id("fraud"));
    expect((await passports.passport(P)).status).to.equal(4n);

    await expect(asRelayer.passports.revoke(P, id("fraud"))).to.be.revertedWithCustomError(passports, "AlreadyRevoked");
    await expect(asRelayer.passports.revoke(id("ghost"), id("fraud"))).to.be.revertedWithCustomError(passports, "PassportUnknown");
    await expect(asRelayer.passports.retire(P)).to.be.revertedWithCustomError(passports, "PassportNotActive");
    await expect(asRelayer.passports.transform([P], [1n], SPLIT, [{ id: id("x"), quantity: 1n, dataHash: id("r") }], id("r")))
      .to.be.revertedWithCustomError(passports, "PassportNotActive");
    const h = await handover(P, 10n);
    await expect(asRelayer.custody.recordHandover(h, await sign(farmKey, h), await sign(exporterKey, h)))
      .to.be.revertedWithCustomError(passports, "PassportNotActive");
  });

  it("flags a known passport", async function () {
    const { passports, issue, asRelayer } = await setup();
    const P = await issue("lot", 10n);
    await expect(asRelayer.passports.flag(P, id("concern"))).to.emit(passports, "Flagged").withArgs(P, id("concern"));
    await expect(asRelayer.passports.flag(id("ghost"), id("concern"))).to.be.revertedWithCustomError(passports, "PassportUnknown");
    await expect(passports.flag(P, id("concern"))).to.be.revertedWithCustomError(passports, "NotRelayer");
  });

  it("only the custody manager can change a holder, and it is wired once", async function () {
    const { passports, issue, asRelayer, stranger, EXPORTER } = await setup();
    const P = await issue("lot", 10n);
    await expect(asRelayer.passports.transferCustody(P, EXPORTER)).to.be.revertedWithCustomError(passports, "NotCustodian");
    await expect(passports.setCustodian(stranger.address)).to.be.revertedWithCustomError(passports, "CustodianAlreadySet");
    await expect(asRelayer.passports.setCustodian(stranger.address)).to.be.revertedWithCustomError(passports, "NotAdmin");
    const fresh = await ethers.deployContract("PassportRegistry", [await passports.orgs(), await passports.facilities()]);
    await expect(fresh.setCustodian(ethers.ZeroAddress)).to.be.revertedWithCustomError(fresh, "ZeroValue");
  });
});

describe("CustodyManager: dual-signed handover", function () {
  it("moves custody when sender and receiver have both signed", async function () {
    const { passports, custody, issue, asRelayer, farmKey, exporterKey, handover, sign, FARM, EXPORTER } = await setup();
    const P = await issue("lot", 500n);
    const h = await handover(P, 500n);

    await expect(asRelayer.custody.recordHandover(h, await sign(farmKey, h), await sign(exporterKey, h)))
      .to.emit(custody, "CustodyTransferred").withArgs(P, FARM, EXPORTER, 500n, h.dataHash);
    expect((await passports.passport(P)).holder).to.equal(EXPORTER);
    expect(await custody.nonces(P)).to.equal(1n);
  });

  it("refuses a handover signed by only one party, or by the wrong keys", async function () {
    const { custody, issue, asRelayer, farmKey, exporterKey, spareKey, handover, sign } = await setup();
    const P = await issue("lot", 500n);
    const h = await handover(P, 500n);
    const fromSig = await sign(farmKey, h), toSig = await sign(exporterKey, h);
    const record = asRelayer.custody.recordHandover;

    // The sender signing twice is not the receiver agreeing.
    await expect(record(h, fromSig, fromSig)).to.be.revertedWithCustomError(custody, "BadSignature").withArgs(h.toOrg);
    await expect(record(h, toSig, toSig)).to.be.revertedWithCustomError(custody, "BadSignature").withArgs(h.fromOrg);
    await expect(record(h, await sign(spareKey, h), toSig)).to.be.revertedWithCustomError(custody, "BadSignature");
    await expect(record(h, fromSig, "0x1234")).to.be.revertedWithCustomError(custody, "BadSignature");
    await expect(record(h, "0x", toSig)).to.be.revertedWithCustomError(custody, "BadSignature");
  });

  it("refuses a message altered after signing", async function () {
    const { custody, orgs, issue, asRelayer, farmKey, exporterKey, certifierKey, handover, sign, CERTIFIER } = await setup();
    const P = await issue("lot", 500n);
    const h = await handover(P, 500n);
    const fromSig = await sign(farmKey, h), toSig = await sign(exporterKey, h);

    // Redirect the goods to a different (verified) organisation.
    const redirected = { ...h, toOrg: CERTIFIER };
    await expect(asRelayer.custody.recordHandover(redirected, fromSig, await sign(certifierKey, redirected)))
      .to.be.revertedWithCustomError(custody, "BadSignature").withArgs(h.fromOrg);
    // Change the record the parties agreed on.
    const rewritten = { ...h, dataHash: id("different record") };
    await expect(asRelayer.custody.recordHandover(rewritten, fromSig, toSig)).to.be.revertedWithCustomError(custody, "BadSignature");
    expect(await orgs.isVerified(CERTIFIER)).to.equal(true);
  });

  it("refuses a replayed handover", async function () {
    const { custody, issue, asRelayer, farmKey, exporterKey, handover, sign, FARM, EXPORTER } = await setup();
    const P = await issue("lot", 500n);
    const h = await handover(P, 500n);
    const fromSig = await sign(farmKey, h), toSig = await sign(exporterKey, h);
    await asRelayer.custody.recordHandover(h, fromSig, toSig);

    await expect(asRelayer.custody.recordHandover(h, fromSig, toSig)).to.be.revertedWithCustomError(custody, "BadNonce").withArgs(0n, 1n);

    // Hand it back, then try the first pair of signatures again.
    const back = await handover(P, 500n, { fromOrg: EXPORTER, toOrg: FARM });
    await asRelayer.custody.recordHandover(back, await sign(exporterKey, back), await sign(farmKey, back));
    await expect(asRelayer.custody.recordHandover(h, fromSig, toSig)).to.be.revertedWithCustomError(custody, "BadNonce").withArgs(0n, 2n);
  });

  it("refuses expired signatures, a non-holder, quantity mismatch, self-handover and unverified parties", async function () {
    const { custody, passports, orgs, issue, asRelayer, farmKey, exporterKey, certifierKey, handover, sign, stranger, FARM, EXPORTER, CERTIFIER } = await setup();
    const P = await issue("lot", 500n);
    const both = async (h: Awaited<ReturnType<typeof handover>>, a = farmKey, b = exporterKey) =>
      [h, await sign(a, h), await sign(b, h)] as const;
    const record = asRelayer.custody.recordHandover;

    const late = await handover(P, 500n, { deadline: BigInt(await networkHelpers.time.latest()) - 1n });
    await expect(record(...(await both(late)))).to.be.revertedWithCustomError(custody, "Expired");

    // The exporter does not hold this lot, even though both "parties" sign.
    const notHolder = await handover(P, 500n, { fromOrg: EXPORTER, toOrg: CERTIFIER });
    await expect(record(...(await both(notHolder, exporterKey, certifierKey)))).to.be.revertedWithCustomError(custody, "NotHolder");

    const short = await handover(P, 499n);
    await expect(record(...(await both(short)))).to.be.revertedWithCustomError(custody, "QuantityMismatch").withArgs(499n, 500n);

    const self = await handover(P, 500n, { toOrg: FARM });
    await expect(record(...(await both(self, farmKey, farmKey)))).to.be.revertedWithCustomError(custody, "SameOrg");

    const noRecord = await handover(P, 500n, { dataHash: ZERO });
    await expect(record(...(await both(noRecord)))).to.be.revertedWithCustomError(custody, "ZeroValue");

    const ok = await handover(P, 500n);
    await asRelayer.orgs.setVerified(EXPORTER, false);
    await expect(record(...(await both(ok)))).to.be.revertedWithCustomError(custody, "OrgNotVerified").withArgs(EXPORTER);
    await asRelayer.orgs.setVerified(EXPORTER, true);
    await asRelayer.orgs.setVerified(FARM, false);
    await expect(record(...(await both(ok)))).to.be.revertedWithCustomError(custody, "OrgNotVerified").withArgs(FARM);
    await asRelayer.orgs.setVerified(FARM, true);

    // Valid signatures still need the relayer to submit them.
    await expect(custody.connect(stranger).recordHandover(...(await both(ok)))).to.be.revertedWithCustomError(custody, "NotRelayer");
    await expect(record({ ...ok, passportId: id("ghost") }, "0x", "0x")).to.be.revertedWithCustomError(passports, "PassportUnknown");
  });

  it("a rotated key signs from then on; the old key no longer does", async function () {
    const { custody, orgs, issue, asRelayer, farmKey, exporterKey, spareKey, handover, sign, FARM } = await setup();
    const P = await issue("lot", 500n);
    await asRelayer.orgs.setSigner(FARM, spareKey.address);
    const h = await handover(P, 500n);
    await expect(asRelayer.custody.recordHandover(h, await sign(farmKey, h), await sign(exporterKey, h)))
      .to.be.revertedWithCustomError(custody, "BadSignature").withArgs(FARM);
    await asRelayer.custody.recordHandover(h, await sign(spareKey, h), await sign(exporterKey, h));
  });

  it("hashHandover matches what a wallet library signs", async function () {
    const { custody, issue, handover, domain } = await setup();
    const h = await handover(await issue("lot", 5n), 5n);
    expect(await custody.hashHandover(h)).to.equal(ethers.TypedDataEncoder.hash(domain, HANDOVER_TYPES, h));
  });

  it("the shared vector holds: same digest, signature and domain the backend must produce", async function () {
    const { custody, domain } = await setup();
    const wallet = new ethers.Wallet(vector.privateKey);
    expect(wallet.address).to.equal(vector.signer);
    expect(ethers.TypedDataEncoder.hash(vector.domain, vector.types, vector.message)).to.equal(vector.digest);
    expect(await wallet.signTypedData(vector.domain, vector.types, vector.message)).to.equal(vector.signature);

    // The contract's own domain name and version are the vector's.
    const onChain = await custody.eip712Domain();
    expect([onChain.name, onChain.version]).to.deep.equal([vector.domain.name, vector.domain.version]);
    expect([domain.name, domain.version]).to.deep.equal([vector.domain.name, vector.domain.version]);
    // And the contract's type hash is the hash of the vector's field list.
    expect(await custody.HANDOVER_TYPEHASH()).to.equal(ethers.id(ethers.TypedDataEncoder.from(vector.types).encodeType("Handover")));
  });

  it("records journey events for an active passport only", async function () {
    const { custody, passports, issue, asRelayer, FARM } = await setup();
    const P = await issue("lot", 5n);
    const EXPORT = id("EXPORT");
    await expect(asRelayer.custody.recordEvent(P, EXPORT, id("export record")))
      .to.emit(custody, "TraceEvent").withArgs(P, FARM, EXPORT, id("export record"));

    await expect(asRelayer.custody.recordEvent(P, ZERO, id("r"))).to.be.revertedWithCustomError(custody, "ZeroValue");
    await expect(asRelayer.custody.recordEvent(P, EXPORT, ZERO)).to.be.revertedWithCustomError(custody, "ZeroValue");
    await expect(custody.recordEvent(P, EXPORT, id("r"))).to.be.revertedWithCustomError(custody, "NotRelayer");
    await expect(asRelayer.custody.recordEvent(id("ghost"), EXPORT, id("r"))).to.be.revertedWithCustomError(passports, "PassportUnknown");
    await asRelayer.passports.retire(P);
    await expect(asRelayer.custody.recordEvent(P, EXPORT, id("r"))).to.be.revertedWithCustomError(custody, "PassportNotActive");
  });
});

describe("AttestationRegistry", function () {
  const A = id("cert-1");
  const window = async () => {
    const now = BigInt(await networkHelpers.time.latest());
    return { from: now - 10n, to: now + 1000n };
  };

  it("an accredited issuer attaches a certificate, valid inside its window", async function () {
    const { attestations, issue, asRelayer, CERTIFIER } = await setup();
    const P = await issue("lot", 5n);
    const { from, to } = await window();

    await expect(asRelayer.attestations.issue(A, CERTIFIER, P, LAB, id("cert pdf"), from, to))
      .to.emit(attestations, "AttestationIssued").withArgs(A, CERTIFIER, P, LAB, id("cert pdf"), from, to);
    expect(await attestations.isValid(A)).to.equal(true);
    const a = await attestations.attestation(A);
    expect([a.issuerOrg, a.subject, a.certType, a.docHash, a.revoked]).to.deep.equal([CERTIFIER, P, LAB, id("cert pdf"), false]);

    await networkHelpers.time.increaseTo(to + 1n);
    expect(await attestations.isValid(A)).to.equal(false); // expired
  });

  it("is not yet valid before its window opens", async function () {
    const { attestations, asRelayer, CERTIFIER } = await setup();
    const { to } = await window();
    await asRelayer.attestations.issue(A, CERTIFIER, id("subject"), LAB, id("doc"), to - 100n, to);
    expect(await attestations.isValid(A)).to.equal(false);
    expect(await attestations.isValid(id("ghost"))).to.equal(false);
  });

  it("only issuers accredited for that certificate type can issue", async function () {
    const { attestations, orgs, asRelayer, stranger, FARM, CERTIFIER } = await setup();
    const { from, to } = await window();
    const issueAs = (org: string, type = LAB) => asRelayer.attestations.issue(A, org, id("subject"), type, id("doc"), from, to);

    await expect(issueAs(FARM)).to.be.revertedWithCustomError(attestations, "NotAccredited").withArgs(FARM, LAB);
    // Accredited for lab reports, not for certificates of origin.
    await expect(issueAs(CERTIFIER, ORIGIN_CERT)).to.be.revertedWithCustomError(attestations, "NotAccredited").withArgs(CERTIFIER, ORIGIN_CERT);
    // Outside the accreditation window: before it opens, then after it ends.
    const now = BigInt(await networkHelpers.time.latest());
    await asRelayer.orgs.setAccreditation(CERTIFIER, ORIGIN_CERT, now + 100n, now + 200n);
    await expect(issueAs(CERTIFIER, ORIGIN_CERT)).to.be.revertedWithCustomError(attestations, "NotAccredited");
    await networkHelpers.time.increaseTo(now + 201n);
    await expect(issueAs(CERTIFIER, ORIGIN_CERT)).to.be.revertedWithCustomError(attestations, "NotAccredited");
    // Removed.
    await asRelayer.orgs.setAccreditation(CERTIFIER, LAB, 0n, 0n);
    await expect(issueAs(CERTIFIER)).to.be.revertedWithCustomError(attestations, "NotAccredited");
    await asRelayer.orgs.setAccreditation(CERTIFIER, LAB, 0n, FOREVER);
    await asRelayer.orgs.setVerified(CERTIFIER, false); // accredited but no longer verified
    await expect(issueAs(CERTIFIER)).to.be.revertedWithCustomError(attestations, "NotAccredited");
    await expect(attestations.connect(stranger).issue(A, CERTIFIER, id("subject"), LAB, id("doc"), from, to)).to.be.revertedWithCustomError(attestations, "NotRelayer");
  });

  it("rejects duplicates, empty values and a window that ends before it starts", async function () {
    const { attestations, asRelayer, CERTIFIER } = await setup();
    const { from, to } = await window();
    const issue = asRelayer.attestations.issue;
    await issue(A, CERTIFIER, id("subject"), LAB, id("doc"), from, to);
    await expect(issue(A, CERTIFIER, id("subject"), LAB, id("doc"), from, to)).to.be.revertedWithCustomError(attestations, "AttestationExists");
    await expect(issue(id("b"), CERTIFIER, id("subject"), LAB, id("doc"), to, from)).to.be.revertedWithCustomError(attestations, "BadValidity");
    await expect(issue(id("b"), CERTIFIER, id("subject"), LAB, id("doc"), to, to)).to.be.revertedWithCustomError(attestations, "BadValidity");
    await expect(issue(ZERO, CERTIFIER, id("subject"), LAB, id("doc"), from, to)).to.be.revertedWithCustomError(attestations, "ZeroValue");
    await expect(issue(id("b"), CERTIFIER, ZERO, LAB, id("doc"), from, to)).to.be.revertedWithCustomError(attestations, "ZeroValue");
    await expect(issue(id("b"), CERTIFIER, id("subject"), LAB, ZERO, from, to)).to.be.revertedWithCustomError(attestations, "ZeroValue");
    await expect(issue(id("b"), CERTIFIER, id("subject"), ZERO, id("doc"), from, to)).to.be.revertedWithCustomError(attestations, "ZeroValue");
  });

  it("revocation is permanent", async function () {
    const { attestations, asRelayer, CERTIFIER } = await setup();
    const { from, to } = await window();
    await asRelayer.attestations.issue(A, CERTIFIER, id("subject"), LAB, id("doc"), from, to);

    await expect(attestations.revoke(A)).to.be.revertedWithCustomError(attestations, "NotRelayer");
    await expect(asRelayer.attestations.revoke(A)).to.emit(attestations, "Revoked").withArgs(A);
    expect(await attestations.isValid(A)).to.equal(false);
    expect((await attestations.attestation(A)).revoked).to.equal(true);
    await expect(asRelayer.attestations.revoke(A)).to.be.revertedWithCustomError(attestations, "AlreadyRevoked");
    await expect(asRelayer.attestations.revoke(id("ghost"))).to.be.revertedWithCustomError(attestations, "AttestationUnknown");
    await expect(attestations.attestation(id("ghost"))).to.be.revertedWithCustomError(attestations, "AttestationUnknown");
  });
});
