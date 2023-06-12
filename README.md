# Node go chainpqc

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

    git clone https://github.com/chainpqc/chainpqc-node
    cd chainpqc-node

install go modules

    go get ./...

    sudo cp .config/liboqs.pc.linux /usr/share/pkgconfig/liboqs.pc

    mkdir ~/.chainpqc

    cp -r config ~/.chainpqc/

You need to choose proper genesis file (you need to change name of chosen file to `genesis.json` in catalog `.chainpqc/config`)

    mv ~/.chaipqc/config/genesis_ami_falcon20.json ~/.chainpqc/config/genesis.json

Copy env file and change accordingly.

    cp .env_sample ~/.chainpqc/.env

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


Copy genesis wallet if you are DELEGATED_ACCOUNT=1. Different DELEGATED_ACCOUNT must have different wallets:

    cp -rf db/Falcon-512 ~/.chainpqc/db/wallet/

To create account and manage wallet (optional, please do not run unless you know what you are doing :)

    cd cmd/walletManagement
    go run main.go

Run Node

    cd cmd/mining
    go run main.go [IP_NODE]

IP_NODE (optional) = IP of other node which is just runned and we would like to connect to. In the case of first node just leave it blanck.
    
Run GUI (default password "a" - please do not change):

    cd cmd/gui
    go run main.go [IP_MINER]

IP_MINER - in most cases it is 127.0.0.1

Running ChainPQC blockchain on more than 1 server:

First server is configured as above.

Edit ~/.chainpqc/.env and set DELEGATED_ACCOUNT=1

and run by command:

    go run cmd/mining/main.go

then run GUI on the same computer:

    go run cmd/gui/main.go

In the case of running next nodes one has to create new wallets. IMPORTANT one node per one computer. One cannot run 2 nodes on one computer:

    go run cmd/walletManagement/main.go
    
In wallet management just set password 'a' and push button `Generate new wallet`. After succesfull wallet creation, close window walletManagement.

In the case of all servers/nodes you need to go through the flow in this README up to line 61. Only on first node you should copy wallet from repo, so exec line 65 in this README. The rest nodes must have other public/private keys.

Edit ~/.chainpqc/.env and set DELEGATED_ACCOUNT to unique number between 2 and 255. IMPORTANT each node has to have unique number.

Now run node by:

    go run cmd/mining/main.go IP_OF_FIRST_NODE
    
    where IP_OF_FIRST_NODE is the IP address of first node.
    
Then on the computer with running miner:

    go run cmd/gui/main.go
    
Only node which generate genesis block, so first node has to be configured and run seperately. All rest nodes are configured and runned similar.
