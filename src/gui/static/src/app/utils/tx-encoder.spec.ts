import { readJSON } from 'karma-read-json';
import { TxEncoder } from './tx-encoder';
import BigNumber from 'bignumber.js';

describe('TxEncoder', () => {

  describe('check encoding', () => {
    const txs = readJSON('test-fixtures/encoded-txs.json').txs;

    for (let i = 0; i < txs.length; i++) {
      it('encode tx ' + i, () => {
        (txs[i].outputs as any[]).forEach(output => {
          output.coins = new BigNumber(output.coins).dividedBy(1000000).toString();
          output.hours = new BigNumber(output.hours).toString();
        });

        expect(TxEncoder.encode(txs[i].inputs, txs[i].outputs, txs[i].signatures, txs[i].innerHash)).toBe(txs[i].raw);
      });
    }
  });
});
