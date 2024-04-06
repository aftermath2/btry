# BTRY

BTRY is an accountless lottery that uses asymmetric cryptography to register bets/prizes, and the Lightning Network to send and receive payments. The lotteries are:

- Completely private
- Globally available
- Permissionless
- Transparent and auditable
- Bitcoin only

Onion site: http://22id55jzspf3mo5duk5z4honwhqdl7ebtmkemccknmpkniytxwqzzzyd.onion

> [!Note]
> BTRY requires a browser proxied through [Tor](https://www.torproject.org) to be accessed. It can't be accessed through clearnet.

## Lottery

Users participate for the opportunity of winning the funds that were bet in the same lottery. Each one lasts 144 Bitcoin blocks (~24 hours).

Winning tickets are generated using the bytes of the Bitcoin block hash that was mined at the lottery height target. Any user can generate the winning tickets themselves and verify that the prizes were correctly assigned.

BTRY decodes the hash and iterates the bytes in reverse, it uses two numbers to calculate each winning ticket. The formula used is $(a ^ b)\mod prizePool$.

For example:

```go
blockHash = "000000000000000000003bc0544004a6e74beb66b21b1e564eb81dbd478d67c6"
decodedBlockHashBytes = [0 0 0 0 0 0 0 0 0 0 59 192 84 64 4 166 231 75 235 102 178 27 30 86 78 184 29 189 71 141 103 198]
prizePool = 10,000

firstWinner = (198 ^ 103) % prizePool = 9,392
secondWinner = (141 ^ 71) % prizePool = 5,941
thirdWinner = (189 ^ 29) % prizePool = 4,909
...
eighthWinner = (75 ^ 231) % prizePool = 1,875
```

### Bets

One payment is one bet and the number of sats is the number of tickets the user gets (1 sat = 1 ticket). Bets can be as little as 1 sat and as big as the capacity available.

In this lottery, ticket numbers are not chosen by the user but rather assigned sequentially. 

> For example, if the first player bets 500,000 sats, it will have tickets from 1 to 500,000 (including the last one). A second user betting 100,000 sats will have tickets from 500,001 to 600,000.

A single ticket can win multiple prizes. All users participate for the **99.609375%** of the prize pool.

### Prizes

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

Prizes expire after **720 blocks**, so make sure to withdraw them within this window. This is to avoid having liquidity locked for long periods of time, which would disable the ability of receiving more bets.

If you would like the prizes to be sent to you automatically, consider linking a lightning address to your private key and BTRY will attempt to pay the winners after they are known. Please note that this may degrade your privacy.

> Users can also opt to receive notifications through telegram in case of winning.

### Authentication

No account required, just an [ed25519](https://en.wikipedia.org/wiki/EdDSA#Ed25519) key pair. It can be generated randomly by the client or provided by the user, please make sure to back it up since it's the only way you can withdraw your prizes.

For higher privacy it's suggested not to re-use the same secret on different days.

To generate your own private key you could use this simple command

```console
openssl genpkey -algorithm ed25519 -out private_key.pem
openssl asn1parse -in private_key.pem -offset 14 | tail -c 65
```

> [!Warning]
> Do not share your private key. It is never sent to the server, only its derived public key and signature are used to store your bets and claim the prizes in case you win.

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
