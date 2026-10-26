#!/usr/bin/env python3
"""Combine the separate Keycloak definitions for a startup realm import."""

import argparse
import json
from pathlib import Path


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--check", action="store_true", help="validate without writing")
    args = parser.parse_args()

    directory = Path(__file__).resolve().parent
    realm = json.loads((directory / "guild-banker-realm.json").read_text())
    clients = json.loads((directory / "guild-banker-clients.json").read_text())
    if realm.get("realm") != "guild-banker" or "clients" in realm:
        parser.error("realm must be guild-banker with clients defined separately")
    client_ids = [client["clientId"] for client in clients["clients"]]
    if sorted(client_ids) != ["guild-banker-api", "guild-banker-ui"]:
        parser.error("clients must contain guild-banker-api and guild-banker-ui exactly once")

    realm["clients"] = clients["clients"]
    if args.check:
        print("Keycloak realm and client JSON definitions are valid")
        return
    destination = directory / "import" / "guild-banker-realm.json"
    destination.parent.mkdir(parents=True, exist_ok=True)
    destination.write_text(json.dumps(realm, indent=2) + "\n")
    print("Prepared infra/keycloak/import/guild-banker-realm.json (password remains an env placeholder)")


if __name__ == "__main__":
    main()
