# Node go QUAD

Works for Ubuntu 20.04+ and go1.22.5+

Install OQS library:

    git clone https://github.com/open-quantum-safe/liboqs.git
    cd liboqs/
    git checkout ea44f39
    
Compile OQS with `-DBUILD_SHARED_LIBS=ON` and install
    
    mkdir build && cd build
    cmake -GNinja -DBUILD_SHARED_LIBS=ON ..    
    ninja
    sudo ninja install

Install prerequisites
    
    sudo apt update
    sudo apt install librocksdb-dev
    sudo apt install libzmq3-dev
    sudo apt install pkg-config
    sudo apt install build-essential
    sudo apt install qtbase5-dev qtchooser qt5-qmake qtbase5-dev-tools

Reload dynamic libraries

    sudo ldconfig -v

Clone project source code

    git clone https://github.com/quad-foundation/quad-node.git
    cd quad-node

install go modules

    go get ./...

    mkdir ~/.quad

Copy env file and change accordingly.

    cp .quad/.env ~/.quad/.env

In the case you are the first who run blockchain and generate genesis block you need to set in .env: DELEGATED_ACCOUNT=1. In other case if you join to other node which is running you can choose unique DELEGATED_ACCOUNT > 1 and < 255.

Ports TCP needed to be opened:

    TransactionTopic: 9091,
    NonceTopic:       8091,
    SelfNonceTopic:   7091,
    SyncTopic:        6091,

    9009 - wallet - node communication


To create account and manage wallet:

    go run cmd/generateNewWallet/main.go

Run Node:

    go run cmd/mining/main.go
 
Run GUI:

    go run cmd/gui/main.go
