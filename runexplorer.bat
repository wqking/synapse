start cmd.exe /c "synapsebeacon -level trace -listen /ip4/0.0.0.0/tcp/11781 -rpclisten /ip4/127.0.0.1/tcp/11782 -chaincfg testnet.json -datadir=E:\temp\synapse\ex_beacon > E:\temp\synapse\ex_beacon_stdout.log 2>&1"
timeout 5 >nul
start cmd.exe /c "synapseshard.exe -level trace -beacon /ip4/127.0.0.1/tcp/11782 -listen /ip4/127.0.0.1/tcp/11783 > E:\temp\synapse\ex_shard_stdout.log 2>&1"
start cmd.exe /c "synapsevalidator -beacon /ip4/127.0.0.1/tcp/11782 -shard /ip4/127.0.0.1/tcp/11783 -rootkey testnet -networkid testnet -validators 0-255 > E:\temp\synapse\ex_validator_stdout.log 2>&1"
start cmd.exe /c "synapseexplorer -dbdriver mysql -dbhost localhost -dbdatabase synapse -dbuser root -dbpassword "" -level trace -listen /ip4/0.0.0.0/tcp/21781 -chainconfig testnet.json -datadir=E:\temp\synapse\ex_explorer -connect /ip4/0.0.0.0/tcp/11781/ipfs/12D3KooWGQsPt4MMnbftNsh8iNtPY3QwMvrVddf9nDbZsVagPSTs > E:\temp\synapse\ex_explorer_stdout.log 2>&1"
