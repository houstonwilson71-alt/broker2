---
name: BSC Sniper build decisions
description: Key dependency and type fixes for the BSC meme-coin sniper project at bsc-sniper/
---

## pgx/v5 version constraint
Use `github.com/jackc/pgx/v5 v5.6.0` (requires Go 1.20+). v5.5.4 ships a pgservicefile pseudo-version with an invalid commit hash (de7065d787b7) that is rejected by both the Replit package firewall and the Go module proxy. v5.10.0+ requires Go 1.25.
**Why:** The invalid pgservicefile version breaks go mod tidy on Replit.
**How to apply:** Pin pgx/v5 to v5.6.0 in go.mod; remove the explicit pgservicefile indirect entry and let tidy resolve it.

## Go module proxy bypass
Run `go mod tidy` with `GOPROXY=https://proxy.golang.org,direct GONOSUMDB='*' GOFLAGS='-mod=mod'`.
Using `GOPROXY=direct` alone hits GitHub auth failures for deep dependencies (e.g. tyler-smith/go-bip39).
**Why:** Replit package firewall doesn't have all go-ethereum transitive deps cached.

## ethereum.CallMsg vs bind.CallMsg
In go-ethereum, use `ethereum.CallMsg` (from `github.com/ethereum/go-ethereum`) for `rpc.CallContract` calls, not `bind.CallMsg`. The `accounts/abi/bind` package does not export `CallMsg`.
**Why:** Compilation error `undefined: bind.CallMsg`.

## Next.js version
Use `next@14.2.35` (latest 14.x). `14.2.3` is blocked by the Replit security policy.

## docker-compose.yml
Remove the `version:` key — it is obsolete in Compose v2 and triggers a warning.
