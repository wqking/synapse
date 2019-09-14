package explorer

import (
	"encoding/binary"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/phoreproject/synapse/chainhash"
	"github.com/phoreproject/synapse/primitives"
	"github.com/prysmaticlabs/go-ssz"

	"github.com/jinzhu/gorm"
	"github.com/phoreproject/synapse/beacon/app"
	logger "github.com/sirupsen/logrus"
)

// Explorer is a blockchain explorer.
// The explorer streams blocks from the beacon chain as they are received
// and then keeps track of its own blockchain so that it can access more
// info like forking.
type Explorer struct {
	app *app.BeaconApp

	config app.Config

	database *Database
}

// NewExplorer creates a new block explorer
func NewExplorer(c app.Config, gormDB *gorm.DB) (*Explorer, error) {
	return &Explorer{
		app:      app.NewBeaconApp(c),
		database: NewDatabase(gormDB),
		config:   c,
	}, nil
}

// WaitForConnections waits until beacon app is connected
func (ex *Explorer) WaitForConnections(numConnections int) {
	for {
		if ex.app.GetHostNode().PeersConnected() >= numConnections {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func combineHashes(in [][32]byte) []byte {
	out := make([]byte, 32*len(in))

	for i, h := range in {
		copy(out[32*i:32*(i+1)], h[:])
	}

	return out
}

func splitHashes(in []byte) [][32]byte {
	out := make([][32]byte, len(in)/32)

	for i := range out {
		copy(out[i][:], in[32*i:32*(i+1)])
	}

	return out
}

func (ex *Explorer) postProcessHook(block *primitives.Block, state *primitives.State, receipts []primitives.Receipt) {
	validators := make(map[int]Validator)

	// Update Validators
	for id, v := range state.ValidatorRegistry {
		var idBytes [4]byte
		binary.BigEndian.PutUint32(idBytes[:], uint32(id))
		pubAndID := append(v.Pubkey[:], idBytes[:]...)
		validatorHash := chainhash.HashH(pubAndID)

		var newV Validator

		ex.database.database.Where(Validator{ValidatorHash: validatorHash[:]}).FirstOrCreate(&newV)

		newV.Pubkey = v.Pubkey[:]
		newV.WithdrawalCredentials = v.WithdrawalCredentials[:]
		newV.Status = v.Status
		newV.LatestStatusChangeSlot = v.LatestStatusChangeSlot
		newV.ExitCount = v.ExitCount
		newV.ValidatorID = uint64(id)

		ex.database.database.Save(&newV)

		validators[id] = newV
	}

	for _, r := range receipts {
		var idBytes [4]byte
		binary.BigEndian.PutUint32(idBytes[:], r.Index)
		pubAndID := append(state.ValidatorRegistry[r.Index].Pubkey[:], idBytes[:]...)
		validatorHash := chainhash.HashH(pubAndID)

		receipt := &Transaction{
			Amount:        r.Amount,
			RecipientHash: validatorHash[:],
			Type:          r.Type,
			Slot:          r.Slot,
		}

		if receipt.Amount > 0 {
			ex.database.database.Create(receipt)
		}
	}

	var epochCount int

	epochStart := state.Slot - (state.Slot % ex.config.NetworkConfig.EpochLength)

	ex.database.database.Model(&Epoch{}).Where(&Epoch{StartSlot: epochStart}).Count(&epochCount)

	if epochCount == 0 {
		var assignments []Assignment

		for i := epochStart; i < epochStart+ex.config.NetworkConfig.EpochLength; i++ {
			assignmentForSlot, err := state.GetShardCommitteesAtSlot(i, ex.config.NetworkConfig)
			if err != nil {
				panic(err)
			}

			for _, as := range assignmentForSlot {
				committeeHashes := make([][32]byte, len(as.Committee))
				for i, member := range as.Committee {
					var idBytes [4]byte
					binary.BigEndian.PutUint32(idBytes[:], member)
					pubAndID := append(state.ValidatorRegistry[member].Pubkey[:], idBytes[:]...)
					committeeHashes[i] = chainhash.HashH(pubAndID)
				}

				assignment := &Assignment{
					Shard:           as.Shard,
					Slot:            i,
					CommitteeHashes: combineHashes(committeeHashes),
				}

				ex.database.database.Create(assignment)

				assignments = append(assignments, *assignment)
			}
		}

		ex.database.database.Create(&Epoch{
			StartSlot:  epochStart,
			Committees: assignments,
		})
	}

	blockHash, err := ssz.HashTreeRoot(block)
	if err != nil {
		panic(err)
	}

	proposerIdx, err := state.GetBeaconProposerIndex(block.BlockHeader.SlotNumber, ex.app.GetBlockchain().GetConfig())
	if err != nil {
		panic(err)
	}

	var idBytes [4]byte
	binary.BigEndian.PutUint32(idBytes[:], proposerIdx)
	pubAndID := append(state.ValidatorRegistry[proposerIdx].Pubkey[:], idBytes[:]...)
	proposerHash := chainhash.HashH(pubAndID)

	blockDB := &Block{
		ParentBlockHash: block.BlockHeader.ParentRoot[:],
		StateRoot:       block.BlockHeader.StateRoot[:],
		RandaoReveal:    block.BlockHeader.RandaoReveal[:],
		Signature:       block.BlockHeader.Signature[:],
		Hash:            blockHash[:],
		Slot:            block.BlockHeader.SlotNumber,
		Proposer:        proposerHash[:],
	}

	ex.database.database.Create(blockDB)

	// Update attestations
	for _, att := range block.BlockBody.Attestations {
		participants, err := state.GetAttestationParticipants(att.Data, att.ParticipationBitfield, ex.config.NetworkConfig)
		if err != nil {
			panic(err)
		}

		participantHashes := make([][32]byte, len(participants))

		for i, p := range participants {
			var idBytes [4]byte
			binary.BigEndian.PutUint32(idBytes[:], p)
			pubAndID := append(state.ValidatorRegistry[p].Pubkey[:], idBytes[:]...)
			validatorHash := chainhash.HashH(pubAndID)

			participantHashes[i] = validatorHash
		}

		// TODO: fixme

		attestation := &Attestation{
			ParticipantHashes:   combineHashes(participantHashes),
			Signature:           att.AggregateSig[:],
			Slot:                att.Data.Slot,
			Shard:               att.Data.Shard,
			BeaconBlockHash:     att.Data.BeaconBlockHash[:],
			ShardBlockHash:      att.Data.ShardBlockHash[:],
			LatestCrosslinkHash: att.Data.LatestCrosslinkHash[:],
			BlockID:             blockDB.ID,
		}

		ex.database.database.Create(attestation)
	}
}

func (ex *Explorer) exit() {
	ex.app.Exit()

	os.Exit(0)
}

// StartExplorer starts the block explorer
func (ex *Explorer) StartExplorer() error {
	signalHandler := make(chan os.Signal, 1)
	signal.Notify(signalHandler, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-signalHandler

		ex.exit()
	}()

	ex.app.RunWithoutBlock()

	ex.app.GetSyncManager().RegisterPostProcessHook(ex.postProcessHook)

	ex.WaitForConnections(1)

	logger.Info("Ready to run.")

	ex.app.WaitForAppExit()

	return nil
}
