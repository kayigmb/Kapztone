# The dev chain: anvil, then the same deployment as scripts/deploy.ts (keep
# the two in step). Anvil's accounts match Hardhat's, so the admin key's
# deploys land on the addresses in .env; the script stops if they don't.
set -eu

anvil --host 0.0.0.0 --chain-id 31337 --silent &
until cast chain-id >/dev/null 2>&1; do sleep 0.5; done

key=$CHAIN_ADMIN_KEY
admin=$(cast wallet address --private-key "$key")
relayer=$(cast wallet address --private-key "$CHAIN_RELAYER_KEY")

# deploy <contract> <constructor signature> <args...>: prints its address.
deploy() {
	name=$1 sig=$2
	shift 2
	cast compute-address "$admin" --nonce "$(cast nonce "$admin")" | awk '{print $NF}'
	cast send --private-key "$key" --create "$(cat "/contracts/$name.hex")$(cast abi-encode "$sig" "$@" | cut -c3-)" >/dev/null
}

orgs=$(deploy OrgRegistry 'c(address,address)' "$admin" "$relayer")
facilities=$(deploy FacilityRegistry 'c(address)' "$orgs")
passports=$(deploy PassportRegistry 'c(address,address)' "$orgs" "$facilities")
custody=$(deploy CustodyManager 'c(address,address)' "$orgs" "$passports")
attestations=$(deploy AttestationRegistry 'c(address)' "$orgs")

# One-time wiring: who may use up capacity, and who may change a holder.
cast send --private-key "$key" "$facilities" 'setOriginator(address)' "$passports" >/dev/null
cast send --private-key "$key" "$passports" 'setCustodian(address)' "$custody" >/dev/null

# check <name> <address in .env> <deployed address>
check() {
	echo "$1=$3"
	if [ -n "$2" ] && [ "$(echo "$2" | tr A-F a-f)" != "$(echo "$3" | tr A-F a-f)" ]; then
		echo "$1 in .env is $2, but the contract is at $3" >&2
		exit 1
	fi
}
check ORG_REGISTRY_ADDRESS "${ORG_REGISTRY_ADDRESS:-}" "$orgs"
check FACILITY_REGISTRY_ADDRESS "${FACILITY_REGISTRY_ADDRESS:-}" "$facilities"
check PASSPORT_REGISTRY_ADDRESS "${PASSPORT_REGISTRY_ADDRESS:-}" "$passports"
check CUSTODY_MANAGER_ADDRESS "${CUSTODY_MANAGER_ADDRESS:-}" "$custody"
check ATTESTATION_REGISTRY_ADDRESS "${ATTESTATION_REGISTRY_ADDRESS:-}" "$attestations"

touch /tmp/deployed
wait
