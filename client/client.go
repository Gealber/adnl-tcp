// Implementation of ADNL on TCP
// Following specifications in https://docs.ton.org/develop/network/adnl-tcp
package client

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"hash"
	"io/ioutil"
	"net"
	"time"

	"github.com/xssnick/tonutils-go/adnl"
	"golang.org/x/crypto/ed25519"

	"github.com/Gealber/adnl-tcp/config"
	"github.com/Gealber/adnl-tcp/utils"
)

type PublicKeyED25519 struct {
	Key ed25519.PublicKey `tl:"int256"`
}

type Client struct {
	privateKey ed25519.PrivateKey
	publicKey  ed25519.PublicKey

	// used by server to encrypt the messages it sends
	// and used by the client to decrypt the received messages
	strmR cipher.Stream
	// used by client to encrypt the messages it sends
	// and used by the server to decrypt the received messages
	strmW cipher.Stream

	// tcp connection
	conn net.Conn
}

type GlobalConfig struct {
	LiteServers []LiteServer `json:"liteservers"`
}

type LiteServer struct {
	// IP integer representation of an IPV4
	IP   int64        `json:"ip"`
	Port int          `json:"port"`
	ID   LiteServerID `json:"id"`
}

func (lts *LiteServer) ConnStr() string {
	return fmt.Sprintf("%s:%d", utils.IntToIPV4(lts.IP), lts.Port)
}

type LiteServerID struct {
	Type string `json:"@type"`
	// Key encoded as base64
	Key string `json:"key"`
}

// pkt represents a regular ADNL TCP packet.
type pkt struct {
	// 4 bytes of packet size in little endian
	size []byte
	// 32 bytes nonce, randome bytes to protect against checksum attack
	nonce []byte
	// payload bytes
	payload []byte
	// 32 bytes SHA256 checksum from nonce and payload
	checksum []byte
}

// handshakePkt
type handshakePkt struct {
	// sever key id
	serverKeyID []byte
	// our public key
	publicKey ed25519.PublicKey
	// sha256 hash from the random 160 bytes
	hash hash.Hash
	// random 160 bytes encrypted
	encrBasis []byte
}

func (pkt *handshakePkt) Len() int {
	return len(pkt.serverKeyID) + len(pkt.publicKey) + pkt.hash.Size() + len(pkt.encrBasis)
}

func (pkt *handshakePkt) assemblyPkt() []byte {
	result := make([]byte, pkt.Len())
	utils.AssemblyBytesSlices(result, pkt.serverKeyID, pkt.publicKey, pkt.hash.Sum(nil), pkt.encrBasis)

	return result
}

// newHandShakePkt ...
func newHandShakePkt(
	serverKeyID []byte,
	privateKey ed25519.PrivateKey,
	publicKey ed25519.PublicKey,
	rndBasis []byte,
) (*handshakePkt, error) {
	h := sha256.New()
	h.Write(rndBasis)
	checksum := h.Sum(nil)

	// 32 bytes
	key := make([]byte, 32)
	// we should use a shared key instead of our public key
	sharedKey, err := utils.SharedKey(privateKey, serverKeyID)
	if err != nil {
		return nil, err
	}

	utils.AssemblyBytesSlices(key, sharedKey[:16], checksum[16:32])
	// 16 bytes
	iv := make([]byte, 16)
	utils.AssemblyBytesSlices(iv, checksum[:4], sharedKey[20:32])

	blk, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	stream := cipher.NewCTR(blk, iv)
	// encrBasis := make([]byte, len(rndBasis))
	stream.XORKeyStream(rndBasis, rndBasis)

	// serverKeyID
	serverKeyID, err = utils.ToKeyID(adnl.PublicKeyED25519{Key: serverKeyID})
	if err != nil {
		return nil, err
	}

	return &handshakePkt{
		serverKeyID: serverKeyID,
		publicKey:   publicKey,
		hash:        h,
		encrBasis:   rndBasis,
	}, nil
}

func New(cfg *config.AppConfig) (*Client, error) {
	// seed := cfg.Client.PrivateKey

	// if seed == "" {
	// 	return nil, errors.New("empty seed checkout environment variable CLIENT_PRIVATE_KEY")
	// }
	pubKey, privateKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		return nil, err
	}

	return &Client{privateKey: privateKey, publicKey: pubKey}, nil
}

func (c *Client) LoadGlobalConfig(filename string) (*GlobalConfig, error) {
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var cfg GlobalConfig
	err = json.Unmarshal(b, &cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (c *Client) Connect(lts LiteServer) error {
	errChn := make(chan error)
	// dial to the desired server
	conn, err := net.Dial("tcp", lts.ConnStr())
	if err != nil {
		return err
	}
	c.conn = conn

	// start listening
	go func() {
        errChn <- c.listen()
	}()

	err = c.handshake(lts)
	if err != nil {
		return fmt.Errorf("handshake error: %v", err)
	}

	err = <-errChn

	return err
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) listen() error {
	// we need to listening and parse responses to our connection
    for {
        // set SetReadDeadline
	    err := c.conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	    if err != nil {
	    	return fmt.Errorf("SetReadDeadline failed: %w", err)
	    }

	    // we should read the size of the whole packet first
	    // the size is specified in the first 4 bytes
	    size, err := c.readPktSize()
	    if err != nil {
	    	return fmt.Errorf("readPktSize error: %v", err)
	    }

	    fmt.Println("PACKET SIZE:", size)

        return nil
	}

	// return nil
}

func (c *Client) readPktSize() (uint32, error) {
	size := make([]byte, 4)

	_, err := c.conn.Read(size)
	if err != nil {
		return 0, err
	}

	// descrypt the size
	c.ctrDecrypt(size)

	// decode as a little endian
	return binary.LittleEndian.Uint32(size), nil
}

func (c *Client) createCTRStreams(rndBasis []byte) error {
	var err error
	// strmR and strmW are AES-CTR ciphers
	c.strmR, err = utils.NewCTRCipher(rndBasis[:32], rndBasis[64:80])
	if err != nil {
		return err
	}

	c.strmW, err = utils.NewCTRCipher(rndBasis[32:64], rndBasis[80:96])
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) handshake(lts LiteServer) error {
	// generate 160 random bytes
	// basis for AES encryption
	rndBasis, err := utils.GenerateRandomBytes(160)
	if err != nil {
		return err
	}

	// create ctr ciphers
	c.createCTRStreams(rndBasis)

	// perform handshake
	srvID, err := base64.StdEncoding.DecodeString(lts.ID.Key)
	if err != nil {
		return err
	}

	// craft handshake packet
	handshakePkt, err := newHandShakePkt(srvID, c.privateKey, c.publicKey, rndBasis)
	if err != nil {
		return err
	}

	// send handshake packet
	// handshakePkt as argument
	data := handshakePkt.assemblyPkt()

	_, err = c.conn.Write(data)

	return err
}

func (c *Client) ctrDecrypt(data []byte) {
	// we need to decrypt this pkt with the streamA
	c.strmR.XORKeyStream(data, data)
}
