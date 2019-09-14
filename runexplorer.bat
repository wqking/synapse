start cmd.exe /c "synapsebeacon -level trace -listen /ip4/0.0.0.0/tcp/11781 -rpclisten 127.0.0.1:11782 -chainconfig testnet.json -datadir=E:\temp\synapse\ex_beacon > E:\temp\synapse\ex_beacon_stdout.log 2>&1"
timeout 5 >nul
start cmd.exe /c "synapseexplorer -level trace -listen /ip4/0.0.0.0/tcp/21781 -chainconfig testnet.json -datadir=E:\temp\synapse\ex_explorer -connect /ip4/0.0.0.0/tcp/11781/ipfs/12D3KooWGQsPt4MMnbftNsh8iNtPY3QwMvrVddf9nDbZsVagPSTs > E:\temp\synapse\ex_explorer_stdout.log 2>&1"
start cmd.exe /c "synapsevalidator -beaconhost :11782 -networkid testnet -validators 0-255 > E:\temp\synapse\ex_validator_stdout.log 2>&1"
