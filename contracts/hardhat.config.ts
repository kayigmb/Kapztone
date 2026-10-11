import hardhatToolboxMochaEthersPlugin from "@nomicfoundation/hardhat-toolbox-mocha-ethers";
import { configVariable, defineConfig } from "hardhat/config";

export default defineConfig({
  plugins: [hardhatToolboxMochaEthersPlugin],
  solidity: {
    version: "0.8.34",
    settings: {
      optimizer: { enabled: true, runs: 200 },
      // Conservative target: avoids opcodes newer than Cancun, which the target
      // testnet may not have yet. Confirm against Amoy before deploying.
      evmVersion: "cancun",
    },
  },
  networks: {
    // The chain .env points at: a local node (`make chain`) or a testnet.
    target: {
      type: "http",
      chainType: "l1",
      url: configVariable("CHAIN_RPC_URL"),
      accounts: [configVariable("CHAIN_PLATFORM_KEY")],
    },
  },
});
