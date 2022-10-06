// https://infura.io/dashboard/ethereum/b8636711061a4f439d6d616cbd84a403/settings
// https://goerli.etherscan.io/address/0xc0ed2724680188d81abef54869ae9c6e1d04753e

const HDWalletProvider = require('@truffle/hdwallet-provider');
const Web3 = require('web3');

const {abi, evm} = require('../compile');

const provider = new HDWalletProvider(
    '',
    ''
);

const web3 = new Web3(provider);

const deploy = async () => {
    const accounts = await web3.eth.getAccounts();

    console.log('Attempting to deploy from account', accounts[0]);

    const result = await new web3.eth.Contract(abi)
        .deploy({data: evm.bytecode.object, arguments: ['Hi there!']})
        .send({gas: '1000000', from: accounts[0]});

    console.log('Contract deployed to', result.options.address);
    provider.engine.stop();
};

deploy();
