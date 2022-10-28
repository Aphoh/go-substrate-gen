# Getting Human Readable JSON
There is a tiny script in `./fetch.ts` which contains the code to grab a human-readable version of the metadata. It will place it in `./meta.json`.

## Use
This project runs via node.js, and so you will need to have npm installed.

To install the dependencies and run, navigate into `json-gen`, run 
```
npm install
```
and to run the script, run
```
npx ts-node fetch.ts
```