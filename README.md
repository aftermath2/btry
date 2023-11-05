# BTRY

BTRY is an accountless lottery that uses asymmetric cryptography to register bets/prizes, and the Lightning Network to send and receive payments.

- Completely private
- Globally available
- Permissionless
- Transparent and auditable
- Bitcoin only

Onion site: <url>.onion

> BTRY requires a browser proxied through Tor to be accessed.

## Lottery

Users participate for the opportunity of winning the funds that were bet in the same UTC day. Winning tickets are chosen using cryptographically secure random numbers and announced at 00:00:00 UTC.

### Bets

One payment is one bet and the number of sats is the number of tickets the user gets (1 sat = 1 ticket).

In this lottery, ticket numbers are not chosen by the user but rather assigned sequentially. 

> For example, if the first player bets 500,000 sats, it will have tickets from 1 to 500,000 (including the last one). A second user betting 100,000 sats will have tickets from 500,001 to 600,000.

A single ticket can win multiple prizes. All users participate for the **99.609375%** of the prize pool.

### Prizes

Prizes expire after **5 days**, so make sure to withdraw them within this window. This is to avoid having liquidity locked for long periods of time, which would disable the ability of receiving more bets.

> Users can opt to receive notifications through telegram in case of winning. Messaging services identifiers are stored for one week and permanently deleted after that.

Prizes distribution as a percentage of the prize pool:

| Place | Prize |
| --- | --- |
| 1st | 50 |
| 2nd | 25 |
| 3rd | 12.5 |
| 4th | 6.25 |
| 5th | 3.125 |
| 6th | 1.5625 |
| 7th | 0.78125 |
| 8th | 0.390625 |
|  |  |
| BTRY fee | 0.390625 |

### Authentication

No account required, just an [ed25519](https://en.wikipedia.org/wiki/EdDSA#Ed25519) key pair. It can be generated randomly by the client or provided by the user, please make sure to back it up since it's the only way you can withdraw your prizes.

For higher privacy it's suggested not to re-use the same secret on different days.

To generate your own private key you could use this simple command

```
openssl genpkey -algorithm ed25519 -out private_key.pem
```

> [!Warning]
> Do not share your private key. It is never sent to the server, only its derived public key and signature are used to store your bets and claim the prizes in case you win.

## Battles (coming soon)

Battles are a **non-custodial** PvP game based on HODL contracts where two players compete for a fixed amount of money.

The funds are never in the hands of BTRY, it merely acts as a coordinator and decides which payment to settle (loser -> winner) and which one to cancel (winner -> loser).

Initially, the idea is simple, both users choose one number between 0 and 1000, BTRY generates one randomly and the one that got it closer wins. However, we may move towards other forms of deciding a winner, like real events outcomes. There has been demand for a sports betting platform based on Lightning on some forums and these technologies may be a great fit for that. 

The random number generation is based on provably fair random numbers, so both parties can validate that the results are fair and not manipulated.

## Building BTRY

> [!Note]
> Requires Go 1.21 and Node.js 20+ installed

```console
git clone https://github.com/aftermath2/btry.git
cd btry
./build.sh
```

### Macaroons

BTRY uses several RPC methods to perform operations with a Lightning Node, we suggest creating a fine-grained macaroon for it:

```
lncli bakemacaroon uri:/lnrpc.Lightning/AddInvoice uri:/lnrpc.Lightning/DecodePayReq uri:/routerrpc.Router/SendPaymentV2 uri:/lnrpc.Lightning/ListChannels uri:/lnrpc.Lightning/SubscribeChannelEvents uri:/lnrpc.Lightning/SubscribeInvoices uri:/routerrpc.Router/TrackPayments
```