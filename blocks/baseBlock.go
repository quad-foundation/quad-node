package blocks

import (
	"fmt"
	"github.com/chainpqc/chainpqc-node/common"
	memDatabase "github.com/chainpqc/chainpqc-node/database"
)

type BaseHeader struct {
	PreviousHash     common.Hash      `json:"previous_hash"`
	Difficulty       int32            `json:"difficulty"`
	Height           int64            `json:"height"`
	DelegatedAccount common.Address   `json:"delegated_account"`
	OperatorAccount  common.Address   `json:"operator_account"`
	Signature        common.Signature `json:"signature"`
	SignatureMessage []byte           `json:"signature_message"`
}

type BaseBlock struct {
	BaseHeader       BaseHeader  `json:"header"`
	BlockHeaderHash  common.Hash `json:"block_header_hash"`
	BlockTimeStamp   int64       `json:"block_time_stamp"`
	RewardPercentage int16       `json:"reward_percentage"`
}

type AnyBlock interface {
	GetBaseBlock() BaseBlock
	GetBlockHeaderHash() common.Hash
	GetBlockTimeStamp() int64
	GetRewardPercentage() int16
	GetChain() uint8
	GetTransactionsHash() common.Hash
	GetBlockHash() common.Hash
	CalcBlockHash() (common.Hash, error)
	CheckProofOfSynergy() bool
	GetBytes() []byte
	GetFromBytes([]byte) (AnyBlock, error)
}

func (b BaseHeader) GetBytesWithoutSignature() []byte {
	rb := b.PreviousHash.GetBytes()
	rb = append(rb, common.GetByteInt32(b.Difficulty)...)
	rb = append(rb, common.GetByteInt64(b.Height)...)
	rb = append(rb, b.DelegatedAccount.GetBytes()...)
	rb = append(rb, b.OperatorAccount.GetBytes()...)
	return rb
}

func (b BaseHeader) GetBytes() []byte {
	rb := b.PreviousHash.GetBytes()
	rb = append(rb, common.GetByteInt32(b.Difficulty)...)
	rb = append(rb, common.GetByteInt64(b.Height)...)
	rb = append(rb, b.DelegatedAccount.GetBytes()...)
	rb = append(rb, b.OperatorAccount.GetBytes()...)
	rb = append(rb, common.BytesToLenAndBytes(b.SignatureMessage)...)
	rb = append(rb, b.Signature.GetBytes()...)
	return rb
}

func (bh *BaseHeader) GetFromBytes(b []byte) ([]byte, error) {
	if len(b) < 84+common.SignatureLength {
		return nil, fmt.Errorf("not enough bytes to decode BaseHeader")
	}
	hash, err := common.GetHashFromBytes(b[:32])
	if err != nil {
		return nil, err
	}
	bh.PreviousHash = hash
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
	msgb, b, err := common.BytesWithLenToBytes(b[84:])
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

func (bb BaseBlock) GetBytes() []byte {
	b := bb.BaseHeader.GetBytes()
	b = append(b, bb.BlockHeaderHash.GetBytes()...)
	b = append(b, common.GetByteInt64(bb.BlockTimeStamp)...)
	b = append(b, common.GetByteInt16(bb.RewardPercentage)...)
	return b
}

func (bb *BaseBlock) GetFromBytes(b []byte) ([]byte, error) {
	if len(b) < 84+common.SignatureLength+44 {
		return nil, fmt.Errorf("not enough bytes to decode BaseBlock")
	}
	b, err := bb.BaseHeader.GetFromBytes(b)
	if err != nil {
		return nil, err
	}
	hash, err := common.GetHashFromBytes(b[:32])
	if err != nil {
		return nil, err
	}
	bb.BlockHeaderHash = hash
	bb.BlockTimeStamp = common.GetInt64FromByte(b[32:40])
	bb.RewardPercentage = common.GetInt16FromByte(b[40:44])
	return b[44:], nil
}

func (b BaseHeader) CalcHash() (common.Hash, error) {
	toByte, err := common.CalcHashToByte(b.GetBytes())
	if err != nil {
		return common.Hash{}, err
	}
	hash := common.Hash{}
	hash, err = hash.Init(toByte)
	if err != nil {
		return common.Hash{}, err
	}
	return hash, nil
}

func StoreBlock(ab AnyBlock) error {
	err := memDatabase.Store(append(common.BlocksDBPrefix[:], ab.GetBlockHash().GetBytes()...), ab.GetBytes())
	if err != nil {
		return err
	}
	bh := common.GetByteInt64(ab.GetBaseBlock().BaseHeader.Height)
	err = memDatabase.Store(append(common.RootHashByHeightDBPrefix[:], bh...), ab.GetBlockHash().GetBytes())
	if err != nil {
		return err
	}
	return nil
}
func LoadHashOfBlock(height int64) (common.Hash, error) {
	bh := common.GetByteInt64(height)
	hashb, err := memDatabase.Load(append(common.RootHashByHeightDBPrefix[:], bh...))
	if err != nil {
		return common.Hash{}, err
	}
	hash, err := common.GetHashFromBytes(hashb)
	if err != nil {
		return common.Hash{}, err
	}
	return hash, nil
}

func LoadBlock(height int64) (AnyBlock, error) {
	bh := common.GetByteInt64(height)
	hash, err := memDatabase.Load(append(common.RootHashByHeightDBPrefix[:], bh...))
	if err != nil {
		return nil, err
	}
	abl, err := memDatabase.Load(append(common.BlocksDBPrefix[:], hash...))
	if err != nil {
		return nil, err
	}
	var bl AnyBlock

	chain := common.GetChainForHeight(height)
	switch chain {
	case 0:
		b, err := TransactionsBlock{}.GetFromBytes(abl)
		if err != nil {
			return nil, err
		}
		bl = b
	case 1:
		b, err := PubKeysBlock{}.GetFromBytes(abl)
		if err != nil {
			return nil, err
		}
		bl = b
	case 2:
		b, err := StakesBlock{}.GetFromBytes(abl)
		if err != nil {
			return nil, err
		}
		bl = b
	case 3:
		b, err := DexBlock{}.GetFromBytes(abl)
		if err != nil {
			return nil, err
		}
		bl = b
	case 4:
		b, err := ContractsBlock{}.GetFromBytes(abl)
		if err != nil {
			return nil, err
		}
		bl = b
	default:
		return nil, fmt.Errorf("no valid chain number")
	}

	return bl, nil
}
