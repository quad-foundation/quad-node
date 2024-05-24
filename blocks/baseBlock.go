package blocks

import (
	"bytes"
	"fmt"
	"github.com/quad-foundation/quad-node/common"
	memDatabase "github.com/quad-foundation/quad-node/database"
	"github.com/quad-foundation/quad-node/wallet"
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
	Supply           int64       `json:"supply"`
}

// GetString returns a string representation of BaseHeader.
func (b *BaseHeader) GetString() string {
	return fmt.Sprintf("PreviousHash: %s\nDifficulty: %d\nHeight: %d\nDelegatedAccount: %s\nOperatorAccount: %s\nRootMerkleTree: %s\nSignature: %s\nSignatureMessage: %x",
		b.PreviousHash.GetHex(), b.Difficulty, b.Height, b.DelegatedAccount.GetHex(), b.OperatorAccount.GetHex(), b.RootMerkleTree.GetHex(), b.Signature.GetHex(), b.SignatureMessage)
}

// GetString returns a string representation of BaseBlock.
func (b *BaseBlock) GetString() string {
	return fmt.Sprintf("Header: {%s}\nBlockHeaderHash: %s\nBlockTimeStamp: %d\nRewardPercentage: %d\nSupply: %d",
		b.BaseHeader.GetString(), b.BlockHeaderHash.GetHex(), b.BlockTimeStamp, b.RewardPercentage, b.Supply)
}

func (b *BaseHeader) GetBytesWithoutSignature() []byte {
	rb := b.PreviousHash.GetBytes()
	rb = append(rb, common.GetByteInt32(b.Difficulty)...)
	rb = append(rb, common.GetByteInt64(b.Height)...)
	rb = append(rb, b.DelegatedAccount.GetBytes()...)
	rb = append(rb, b.OperatorAccount.GetBytes()...)
	rb = append(rb, b.RootMerkleTree.GetBytes()...)
	//rb = append(rb, b.SignatureMessage...)
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
	rb = append(rb, common.BytesToLenAndBytes(b.Signature.GetBytes())...)
	//log.Println("block ", b.Height, " len bytes ", len(rb))
	return rb
}

func (bh *BaseHeader) Verify() bool {
	signatureBlockHeaderMessage := bh.GetBytesWithoutSignature()
	if bytes.Compare(signatureBlockHeaderMessage, bh.SignatureMessage) != 0 {
		return false
	}
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
	w := wallet.GetActiveWallet()
	sign, err := w.Sign(calcHash)
	if err != nil {
		return common.Signature{}, nil, err
	}
	return *sign, signatureBlockHeaderMessage, nil
}

func (bh *BaseHeader) GetFromBytes(b []byte) ([]byte, error) {
	if len(b) < 116+common.SignatureLength {
		return nil, fmt.Errorf("not enough bytes to decode BaseHeader")
	}
	//log.Println("block decompile len bytes ", len(b))

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
	sigBytes, b, err := common.BytesWithLenToBytes(b[:])
	if err != nil {
		return nil, err
	}
	sig, err := common.GetSignatureFromBytes(sigBytes, opAddress)
	if err != nil {
		return nil, err
	}
	bh.Signature = sig
	return b, nil
}

func (bb *BaseBlock) GetBytes() []byte {
	b := bb.BaseHeader.GetBytes()
	b = append(b, bb.BlockHeaderHash.GetBytes()...)
	b = append(b, common.GetByteInt64(bb.BlockTimeStamp)...)
	b = append(b, common.GetByteInt16(bb.RewardPercentage)...)
	b = append(b, common.GetByteInt64(bb.Supply)...)
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
	bb.Supply = common.GetInt64FromByte(b[42:50])
	return b[50:], nil
}

func (b *BaseHeader) CalcHash() (common.Hash, error) {
	toByte, err := common.CalcHashToByte(b.GetBytes())
	if err != nil {
		return common.Hash{}, err
	}
	hash := common.GetHashFromBytes(toByte)
	return hash, nil
}
