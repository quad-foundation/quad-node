package blocks

import (
	"github.com/quad-foundation/quad-node/common"
)

// ProcessBlockPubKey : store pubkey on each transaction
func ProcessBlockEncryption(block Block) error {
	if len(block.BaseBlock.BaseHeader.Encryption1[:]) != 0 {
		enc1, err := FromBytesToEncryptionConfig(block.BaseBlock.BaseHeader.Encryption1[:], 1)
		if err != nil {
			return err
		}

		common.PubKeyLength = enc1.PubKeyLength
		common.PrivateKeyLength = enc1.PrivateKeyLength
		common.SignatureLength = enc1.SignatureLength
		common.SigName = enc1.SigName
		common.IsValid = enc1.IsValid
		common.IsPaused = enc1.IsPaused
	}

	if len(block.BaseBlock.BaseHeader.Encryption2[:]) != 0 {
		enc2, err := FromBytesToEncryptionConfig(block.BaseBlock.BaseHeader.Encryption2[:], 2)
		if err != nil {
			return err
		}

		common.PubKeyLength2 = enc2.PubKeyLength
		common.PrivateKeyLength2 = enc2.PrivateKeyLength
		common.SignatureLength2 = enc2.SignatureLength
		common.SigName2 = enc2.SigName
		common.IsValid2 = enc2.IsValid
		common.IsPaused2 = enc2.IsPaused
	}
	return nil
}
