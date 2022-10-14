const assert = require('assert');
const ganache = require('ganache-cli');
const Web3 = require('web3');
const web3 = new Web3(ganache.provider());

const {abi, evm} = require('../compile');


let lottery;
let accounts;

beforeEach(async () => {
    accounts = await web3.eth.getAccounts();

    lottery = await new web3.eth.Contract(abi)
        .deploy({
            data: evm.bytecode.object
        })
        .send({from: accounts[0], gas: '1000000'});
});

describe('Lottery Contract', () => {

    it('should deploy the lottery contract', () => {
        assert.ok(lottery.options.address);
    });

    it('should allow someone to enter the lottery', async () => {

        //given
        await lottery.methods.enter().send({
            from: accounts[0],
            value: web3.utils.toWei('0.02', 'ether')
        });

        //when
        const players = await lottery.methods.getPlayers().call({
            from: accounts[0]
        });

        //then
        assert.equal(accounts[0], players[0]);
        assert.equal(1, players.length);
    });

    it('should allow multiple people to enter the lottery', async () => {

        //given
        await lottery.methods.enter().send({
            from: accounts[0],
            value: web3.utils.toWei('0.02', 'ether')
        });
        await lottery.methods.enter().send({
            from: accounts[1],
            value: web3.utils.toWei('0.02', 'ether')
        });
        await lottery.methods.enter().send({
            from: accounts[2],
            value: web3.utils.toWei('0.02', 'ether')
        });

        //when
        const players = await lottery.methods.getPlayers().call({
            from: accounts[0]
        });

        //then
        assert.equal(accounts[0], players[0]);
        assert.equal(accounts[1], players[1]);
        assert.equal(accounts[2], players[2]);
        assert.equal(3, players.length);
    });

    it('should require a minimum amount of ether to enter', async () => {
        //given
        try {
            await lottery.methods.enter().send({
                from: accounts[0],
                value: 0 // No Wei!
            });
            assert(false);
        } catch (e) {
            assert(e);
        }

    });

    it('should only allow a manager to pick a winner ', async () => {
        try {
            await lottery.methods.pickWinner().send({
                from: accounts[1]
            });
            assert(false);
        } catch (e) {
            assert(e);
        }
    });

    it("should send money to the winner and resets the players array", async () => {
        await lottery.methods.enter().send({
            from: accounts[0],
            value: web3.utils.toWei("2", "ether"),
        });

        const initialBalance = await web3.eth.getBalance(accounts[0]);
        await lottery.methods.pickWinner().send({from: accounts[0]});
        const finalBalance = await web3.eth.getBalance(accounts[0]);
        const difference = finalBalance - initialBalance;

        assert(difference > web3.utils.toWei("1.8", "ether"));
    });

});

