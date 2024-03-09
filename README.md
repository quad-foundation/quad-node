# Node go QUAD - currently in building process

Works for Ubuntu 20.04+ and go1.19+

Install OQS library:

    git clone https://github.com/open-quantum-safe/liboqs.git
    cd liboqs/
    
Compile OQS with `-DBUILD_SHARED_LIBS=ON` and install
    
    mkdir build && cd build
    cmake -GNinja -DBUILD_SHARED_LIBS=ON ..    
    ninja
    sudo ninja install

Install prerequisites

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

    TransactionTopic[0]: 9091,
    TransactionTopic[1]: 9092,
    TransactionTopic[2]: 9093,
    TransactionTopic[3]: 9094,
    TransactionTopic[4]: 9095,
    NonceTopic[0]:       8091,
    NonceTopic[1]:       8092,
    NonceTopic[2]:       8093,
    NonceTopic[3]:       8094,
    NonceTopic[4]:       8095,
    SelfNonceTopic[0]:   7091,
    SelfNonceTopic[1]:   7092,
    SelfNonceTopic[2]:   7093,
    SelfNonceTopic[3]:   7094,
    SelfNonceTopic[4]:   7095,
    SyncTopic[0]:        6091,
    SyncTopic[1]:        6092,
    SyncTopic[2]:        6093,
    SyncTopic[3]:        6094,
    SyncTopic[4]:        6095,

    9009 - wallet - node communication


To create account and manage wallet:

    go run cmd/generateNewWallet/main.go

Run Node:

    go run cmd/mining/main.go
 
Run GUI:

    go run cmd/gui/main.go
