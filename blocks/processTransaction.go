package blocks

import (
	"bytes"
	"fmt"
	"github.com/quad-foundation/quad-node/account"
	"github.com/quad-foundation/quad-node/common"
	"github.com/quad-foundation/quad-node/transactionsDefinition"
	"log"
)

func CheckStakingTransaction(tx transactionsDefinition.Transaction, sumAmount int64, sumFee int64) bool {
	fee := tx.GasPrice * tx.GasUsage
	amount := tx.TxData.Amount
	address := tx.GetSenderAddress()
	acc := account.GetAccountByAddressBytes(address.GetBytes())
	if bytes.Compare(acc.Address[:], address.GetBytes()) != 0 {
		log.Println("no account found in check staking transaction")
		return false
	}
	if acc.Balance < fee {
		log.Println("not enough funds on account to cover fee")
		return false
	}
	if acc.Balance < sumFee {
		log.Println("not enough funds on account to cover sumFee")
		return false
	}
	addressRecipient := tx.TxData.Recipient
	n, err := account.IntDelegatedAccountFromAddress(addressRecipient)
	if err != nil {
		log.Println(err)
		return false
	}
	if n > 0 && n < 256 {
		accStaking := account.GetStakingAccountByAddressBytes(address.GetBytes(), n)
		if bytes.Compare(accStaking.DelegatedAccount[:], addressRecipient.GetBytes()) != 0 {
			if amount <= 0 {
				log.Println("no staking account found in check staking transaction")
				return false
			}

		}
		if amount < common.MinStakingUser && amount > 0 {
			log.Println("staking amount has to be larger than ", common.MinStakingUser)
			return false
		}

		if accStaking.StakedBalance+amount < common.MinStakingUser && accStaking.StakedBalance+amount != 0 {
			log.Println("not enough staked balance. Staking has to be larger than ", common.MinStakingUser)
			return false
		}
		// check for all transactions together

		if sumAmount < common.MinStakingUser && sumAmount > 0 {
			log.Println("staking amount has to be larger than ", common.MinStakingUser)
			return false
		}
		if accStaking.StakedBalance+sumAmount < common.MinStakingUser && accStaking.StakedBalance+sumAmount != 0 {
			log.Println("not enough staked balance. Staking has to be larger than ", common.MinStakingUser)
			return false
		}
	}
	if n >= 256 && n < 512 {

		accStaking := account.GetStakingAccountByAddressBytes(address.GetBytes(), n%256)
		if bytes.Compare(accStaking.Address[:], address.GetBytes()) != 0 {
			log.Println("no staking account found in check staking transaction (rewards)")
			return false
		}
		if accStaking.StakingRewards+amount < 0 {
			log.Println("not enough rewards balance. Rewards has to be larger than ", 0)
			return false
		}
	}
	return true
}

func ProcessTransaction(tx transactionsDefinition.Transaction, height int64) error {
	fee := tx.GasPrice * tx.GasUsage
	amount := tx.TxData.Amount
	address := tx.GetSenderAddress()
	addressRecipient := tx.TxData.Recipient
	n, err := account.IntDelegatedAccountFromAddress(addressRecipient)
	if err == nil { // this is delegated account
		if n > 0 && n < 256 {
			if amount >= common.MinStakingUser {
				err := account.Stake(address.GetBytes(), amount, height, n)
				if err != nil {
					return err
				}
			} else if amount < 0 {
				err := account.Unstake(address.GetBytes(), amount, height, n)
				if err != nil {
					return err
				}

			} else {
				return fmt.Errorf("wrong amount in staking/unstaking")
			}
			err = AddBalance(address.ByteValue, -fee-amount)
			if err != nil {
				return err
			}
		}
		if n >= 256 && n < 512 {

			accStaking := account.GetStakingAccountByAddressBytes(address.GetBytes(), n%256)
			if bytes.Compare(accStaking.Address[:], address.GetBytes()) != 0 {
				return fmt.Errorf("no staking account found in check staking transaction (rewards)")
			}
			if amount > 0 {
				log.Println("not implemented")
				//err := account.Reward(accStaking.Address[:], amount, height, n%256)
				//if err != nil {
				//	return err
				//}
			} else if amount < 0 {
				err := account.WithdrawReward(accStaking.Address[:], amount, height, n%256)
				if err != nil {
					return err
				}
				err = AddBalance(address.ByteValue, -fee-amount)
				if err != nil {
					return err
				}
			} else {
				return fmt.Errorf("wrong amount in rewarding")
			}
		}
	} else { // this is not delegated account so standard transaction
		err := AddBalance(address.ByteValue, -fee-amount)
		if err != nil {
			return err
		}

		err = AddBalance(addressRecipient.ByteValue, amount)
		if err != nil {
			return err
		}
	}
	return nil
}
