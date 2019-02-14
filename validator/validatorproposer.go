package validator

import (
	"context"
	"encoding/binary"
	"fmt"

	"github.com/phoreproject/synapse/beacon/config"
	"github.com/sirupsen/logrus"

	"github.com/golang/protobuf/ptypes/empty"

	"github.com/phoreproject/prysm/shared/ssz"
	"github.com/phoreproject/synapse/bls"
	"github.com/phoreproject/synapse/chainhash"
	"github.com/phoreproject/synapse/pb"
	"github.com/phoreproject/synapse/primitives"
)

const maxAttemptsAttestation = 10

func hammingWeight(b uint8) int {
	b = b - ((b >> 1) & 0x55)
	b = (b & 0x33) + ((b >> 2) & 0x33)
	return int(((b + (b >> 4)) & 0x0F) * 0x01)
}

func (v *Validator) proposeBlock(information proposerAssignment) error {
	attestations, err := v.mempool.attestationMempool.getAttestationsToInclude(information.slot, v.config)
	if err != nil {
		return err
	}

	v.logger.WithFields(logrus.Fields{
		"attestationSize": len(attestations),
		"mempoolSize":     v.mempool.attestationMempool.size(),
	}).Debug("creating block")

	parentRootBytes, err := v.blockchainRPC.GetBlockHash(context.Background(), &pb.GetBlockHashRequest{SlotNumber: information.slot - 1})
	if err != nil {
		return err
	}

	parentRoot, err := chainhash.NewHash(parentRootBytes.Hash)
	if err != nil {
		return err
	}

	stateRootBytes, err := v.blockchainRPC.GetStateRoot(context.Background(), &empty.Empty{})
	if err != nil {

	}

	stateRoot, err := chainhash.NewHash(stateRootBytes.StateRoot)
	if err != nil {
		return err
	}

	var proposerSlotsBytes [8]byte
	binary.BigEndian.PutUint64(proposerSlotsBytes[:], v.keystore.GetProposerSlots(v.id))

	key := v.keystore.GetKeyForValidator(v.id)

	randaoSig, err := bls.Sign(key, proposerSlotsBytes[:], bls.DomainRandao)
	if err != nil {
		return err
	}

	newBlock := primitives.Block{
		BlockHeader: primitives.BlockHeader{
			SlotNumber:   information.slot,
			ParentRoot:   *parentRoot,
			StateRoot:    *stateRoot,
			RandaoReveal: randaoSig.Serialize(),
			Signature:    bls.EmptySignature.Serialize(),
		},
		BlockBody: primitives.BlockBody{
			Attestations:      attestations,
			ProposerSlashings: []primitives.ProposerSlashing{},
			CasperSlashings:   []primitives.CasperSlashing{},
			Deposits:          []primitives.Deposit{},
			Exits:             []primitives.Exit{},
		},
	}

	blockHash, err := ssz.TreeHash(newBlock)
	if err != nil {
		return err
	}

	v.logger.WithField("blockHash", fmt.Sprintf("%x", blockHash)).Info("signing block")

	psd := primitives.ProposalSignedData{
		Slot:      information.slot,
		Shard:     config.MainNetConfig.BeaconShardNumber,
		BlockHash: blockHash,
	}

	psdHash, err := ssz.TreeHash(psd)
	if err != nil {
		return err
	}

	sig, err := bls.Sign(v.keystore.GetKeyForValidator(v.id), psdHash[:], bls.DomainProposal)
	if err != nil {
		return err
	}
	newBlock.BlockHeader.Signature = sig.Serialize()

	v.logger.Debug("submitting block")

	submitBlockRequest := &pb.SubmitBlockRequest{
		Block: newBlock.ToProto(),
	}

	_, err = v.blockchainRPC.SubmitBlock(context.Background(), submitBlockRequest)

	v.logger.Debug("submitted block")

	for _, a := range attestations {
		v.mempool.attestationMempool.removeAttestationsFromBitfield(a.Data.Slot, a.Data.Shard, a.ParticipationBitfield)
	}

	v.keystore.IncrementProposerSlots(v.id)
	return err
}