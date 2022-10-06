const HDHWalletProvider = require('@truffle/hdwallet-provider');
const {interface, bytecode } = require('compile');
const Web3 = require('web3');

const provider = new HDHWalletProvider(
    '',
    ''
);

const web3 = new Web3(provider);


const deploy = async () => {
    const accounts = await web3.eth.getAccounts();
    console.log('Fetched accounts:', accounts);

    const result = await new web3.eth.Contract(JSON.parse(interface))
        .deploy({data: bytecode, arguments: ['Heyho!']})
        .send({from: accounts[0], gas:1000000});

    console.log('Contract deployed to', result.options.address);
};

deploy();
