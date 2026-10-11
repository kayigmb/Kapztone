// Deploys the five contracts to the network .env points at (CHAIN_RPC_URL,
// CHAIN_PLATFORM_KEY) and prints the address lines to paste into .env.
//
//   make deploy
//
// The deploying account becomes both admin and relayer.
import { network } from "hardhat";

const { ethers } = await network.create();
const [platform] = await ethers.getSigners();
const { chainId } = await ethers.provider.getNetwork();
console.error(`deploying to chain ${chainId} as ${platform.address}`);

const orgs = await ethers.deployContract("OrgRegistry", [platform.address, platform.address]);
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
