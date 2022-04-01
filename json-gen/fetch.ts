import { ApiPromise, WsProvider } from '@polkadot/api';
import * as fs from "fs";

// Construct
const main = async () => {
  const wsProvider = new WsProvider('ws://127.0.0.1:9944');
  const api = await ApiPromise.create({ provider: wsProvider });

  const meta = await api.rpc.state.getMetadata()
  fs.writeFileSync("meta.json", JSON.stringify(meta.asLatest.toHuman()))
}

main().catch(console.error).finally(() => process.exit())
//
