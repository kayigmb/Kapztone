// Deploys the five contracts to the network .env points at (CHAIN_RPC_URL)
// and prints the address lines to paste into .env.
//
//   make deploy
//
// The deploying account (CHAIN_ADMIN_KEY) becomes the admin: it only grants
// roles and does the one-time wiring below, so its key can stay offline
// afterwards. The account of CHAIN_RELAYER_KEY becomes relayer and operator:
// the API and relayer use it for everything else.
import { network } from "hardhat";

const { ethers } = await network.create();
const [admin] = await ethers.getSigners();
const relayerKey = process.env.CHAIN_RELAYER_KEY;
if (!relayerKey) throw new Error("CHAIN_RELAYER_KEY is not set: add it to .env (see .env.example)");
const relayer = new ethers.Wallet(relayerKey).address;
const { chainId } = await ethers.provider.getNetwork();
console.error(`deploying to chain ${chainId} as admin ${admin.address}; relayer and operator ${relayer}`);
if (relayer === admin.address) console.error("warning: admin and relayer are the same key; use two keys outside a local chain");

const orgs = await ethers.deployContract("OrgRegistry", [admin.address, relayer]);
const facilities = await ethers.deployContract("FacilityRegistry", [orgs.target]);
const passports = await ethers.deployContract("PassportRegistry", [orgs.target, facilities.target]);
const custody = await ethers.deployContract("CustodyManager", [orgs.target, passports.target]);
const attestations = await ethers.deployContract("AttestationRegistry", [orgs.target]);
for (const c of [orgs, facilities, passports, custody, attestations]) await c.waitForDeployment();

// One-time wiring: who may use up capacity, and who may change a holder.
await (await facilities.setOriginator(passports.target)).wait();
await (await passports.setCustodian(custody.target)).wait();

console.log(`ORG_REGISTRY_ADDRESS=${orgs.target}`);
console.log(`FACILITY_REGISTRY_ADDRESS=${facilities.target}`);
console.log(`PASSPORT_REGISTRY_ADDRESS=${passports.target}`);
console.log(`CUSTODY_MANAGER_ADDRESS=${custody.target}`);
console.log(`ATTESTATION_REGISTRY_ADDRESS=${attestations.target}`);
