// SPDX-License-Identifier: GPL-3.0
pragma solidity ^0.8.4;

contract Coin {
   
    // The keyword "public" makes variables
    // accessible from other contracts
    address public minter;
    mapping (address => int64) public balances;
    string public constant name = "WONABRU TOKEN";
    string public constant symbol = "WNB";
    uint8 public constant decimals = 2;
  
    string store = "heja";

function balanceOf(address tokenOwner) public view returns (int64) {
    return balances[tokenOwner];
}

function getStore() public view returns (string memory) {
     return store;
}

function setStore(string memory n) public {
    store = n;
}

    // Events allow clients to react to specific
    // contract changes you declare
    event Sent(address from, address to, int64 amount);

    // Constructor code is only run when the contract
    // is created
    constructor() {
        minter = msg.sender;
    }

    // Sends an amount of newly created coins to an address
    // Can only be called by the contract creator
    function mint(address receiver, int64 amount) public {
        require(msg.sender == minter);
        balances[receiver] = amount * 5;
    }

    // Errors allow you to provide information about
    // why an operation failed. They are returned
    // to the caller of the function.
    error InsufficientBalance(int64 requested, int64 available);

    // Sends an amount of existing coins
    // from any caller to an address
    function transfer(address receiver, int64 amount) public {
        if (amount > balances[msg.sender])
            revert InsufficientBalance({
                requested: amount,
                available: balances[msg.sender]
            });

        balances[msg.sender] -= amount;
        balances[receiver] += amount;
        emit Sent(msg.sender, receiver, amount);
    }
}