# Own implementation of ADNL client on TCP 

Following description in this [ADNL TCP - Liteserver](https://docs.ton.org/develop/network/adnl-tcp). This implementation is just for learning purpose a full implementation can be found in [tonutils-go](https://github.com/xssnick/tonutils-go/tree/master).

The only operation supported is the initial handshake.
1. [DONE] Handshake connection

As this is for educational purpose, and there are full libraries already implementing it, there's no sense on extending this more. Is a good idea as a educational project tho.

## What I learned from this?

1. The way ADNL protocol pack the data been sent on the network.
2. Got a better understanding of how works the handshake on this protocol.
3. Writing TCP code again, it's been a while. 

