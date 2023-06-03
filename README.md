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

Ports needed to be opened:

    TCP/9009 - wallet - node communication
    TCP/9091 - transaction main chain
    TCP/9081 - nonce messages
    TCP/9071 - self nonce messages
    TCP/9061 - syncing
    TCP/9051 - staking/DEX transaction
    TCP/9092 - side chain transactions

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

Running galaxy blockchain on more than 1 server:

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
