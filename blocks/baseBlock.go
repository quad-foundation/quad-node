package blocks

import (
	"fmt"
	"github.com/chainpqc/chainpqc-node/common"
	memDatabase "github.com/chainpqc/chainpqc-node/database"
	"github.com/chainpqc/chainpqc-node/wallet"
	"log"
)

type BaseHeader struct {
	PreviousHash     common.Hash      `json:"previous_hash"`
	Difficulty       int32            `json:"difficulty"`
	Height           int64            `json:"height"`
	DelegatedAccount common.Address   `json:"delegated_account"`
	OperatorAccount  common.Address   `json:"operator_account"`
	RootMerkleTree   common.Hash      `json:"root_merkle_tree"`
	Signature        common.Signature `json:"signature"`
	SignatureMessage []byte           `json:"signature_message"`
}

type BaseBlock struct {
	BaseHeader       BaseHeader  `json:"header"`
	BlockHeaderHash  common.Hash `json:"block_header_hash"`
	BlockTimeStamp   int64       `json:"block_time_stamp"`
	RewardPercentage int16       `json:"reward_percentage"`
}

//
//type AnyBlock interface {
//	GetBaseBlock() BaseBlock
//	GetBlockHeaderHash() common.Hash
//	GetBlockTimeStamp() int64
//	GetRewardPercentage() int16
//	GetChain() uint8
//	GetTransactionsHash() common.Hash
//	GetBlockHash() common.Hash
//	CalcBlockHash() (common.Hash, error)
//	CheckProofOfSynergy() bool
//	GetBytes() []byte
//	GetFromBytes([]byte) (AnyBlock, error)
//	GetTransactionsHashes(*transactionType.PatriciaMerkleTree) ([]common.Hash, error)
//}

func (b *BaseHeader) GetBytesWithoutSignature() []byte {
	rb := b.PreviousHash.GetBytes()
	rb = append(rb, common.GetByteInt32(b.Difficulty)...)
	rb = append(rb, common.GetByteInt64(b.Height)...)
	rb = append(rb, b.DelegatedAccount.GetBytes()...)
	rb = append(rb, b.OperatorAccount.GetBytes()...)
	rb = append(rb, b.RootMerkleTree.GetBytes()...)
	rb = append(rb, b.SignatureMessage...)
	return rb
}

func (b *BaseHeader) GetBytes() []byte {
	rb := b.PreviousHash.GetBytes()
	rb = append(rb, common.GetByteInt32(b.Difficulty)...)
	rb = append(rb, common.GetByteInt64(b.Height)...)
	rb = append(rb, b.DelegatedAccount.GetBytes()...)
	rb = append(rb, b.OperatorAccount.GetBytes()...)
	rb = append(rb, b.RootMerkleTree.GetBytes()...)
	rb = append(rb, common.BytesToLenAndBytes(b.SignatureMessage)...)
	rb = append(rb, b.Signature.GetBytes()...)
	log.Println("block ", b.Height, " len bytes ", len(rb))
	return rb
}

func (bh *BaseHeader) VerifyTransaction() bool {
	signatureBlockHeaderMessage := bh.GetBytesWithoutSignature()
	calcHash, err := common.CalcHashToByte(signatureBlockHeaderMessage)
	if err != nil {
		return false
	}
	a := bh.OperatorAccount.GetBytes()
	pk, err := memDatabase.MainDB.Get(append(common.PubKeyDBPrefix[:], a...))
	if err != nil {
		return false
	}
	return wallet.Verify(calcHash, bh.Signature.GetBytes(), pk)
}

func (bh *BaseHeader) Sign() (common.Signature, []byte, error) {
	signatureBlockHeaderMessage := bh.GetBytesWithoutSignature()
	calcHash, err := common.CalcHashToByte(signatureBlockHeaderMessage)
	if err != nil {
		return common.Signature{}, nil, err
	}
	w := wallet.EmptyWallet()
	w = w.GetWallet()
	sign, err := w.Sign(calcHash)
	if err != nil {
		return common.Signature{}, nil, err
	}
	return sign, signatureBlockHeaderMessage, nil
}

func (bh *BaseHeader) GetFromBytes(b []byte) ([]byte, error) {
	if len(b) < 116+common.SignatureLength {
		return nil, fmt.Errorf("not enough bytes to decode BaseHeader")
	}
	log.Println("block decompile len bytes ", len(b))

	bh.PreviousHash = common.GetHashFromBytes(b[:32])
	bh.Difficulty = common.GetInt32FromByte(b[32:36])
	bh.Height = common.GetInt64FromByte(b[36:44])
	address, err := common.BytesToAddress(b[44:64])
	if err != nil {
		return nil, err
	}
	bh.DelegatedAccount = address
	opAddress, err := common.BytesToAddress(b[64:84])
	if err != nil {
		return nil, err
	}
	bh.OperatorAccount = opAddress
	bh.RootMerkleTree = common.GetHashFromBytes(b[84:116])
	msgb, b, err := common.BytesWithLenToBytes(b[116:])
	if err != nil {
		return nil, err
	}
	bh.SignatureMessage = msgb
	sig, err := common.GetSignatureFromBytes(b[:common.SignatureLength], opAddress)
	if err != nil {
		return nil, err
	}
	bh.Signature = sig
	return b[common.SignatureLength:], nil
}

func (bb *BaseBlock) GetBytes() []byte {
	b := bb.BaseHeader.GetBytes()
	b = append(b, bb.BlockHeaderHash.GetBytes()...)
	b = append(b, common.GetByteInt64(bb.BlockTimeStamp)...)
	b = append(b, common.GetByteInt16(bb.RewardPercentage)...)
	return b
}

func (bb *BaseBlock) GetFromBytes(b []byte) ([]byte, error) {
	if len(b) < 116+common.SignatureLength+44 {
		return nil, fmt.Errorf("not enough bytes to decode BaseBlock")
	}
	b, err := bb.BaseHeader.GetFromBytes(b)
	if err != nil {
		return nil, err
	}
	bb.BlockHeaderHash = common.GetHashFromBytes(b[:32])
	bb.BlockTimeStamp = common.GetInt64FromByte(b[32:40])
	bb.RewardPercentage = common.GetInt16FromByte(b[40:42])
	return b[42:], nil
}

func (b *BaseHeader) CalcHash() (common.Hash, error) {
	toByte, err := common.CalcHashToByte(b.GetBytes())
	if err != nil {
		return common.Hash{}, err
	}
	hash := common.GetHashFromBytes(toByte)
	return hash, nil
}
