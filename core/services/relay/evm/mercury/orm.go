package mercury

import (
	"crypto/sha256"

	"github.com/ethereum/go-ethereum/common"
	"github.com/smartcontractkit/sqlx"

	ocrtypes "github.com/smartcontractkit/libocr/offchainreporting2plus/types"

	"github.com/smartcontractkit/chainlink/v2/core/logger"
	"github.com/smartcontractkit/chainlink/v2/core/services/pg"
	"github.com/smartcontractkit/chainlink/v2/core/services/relay/evm/mercury/wsrpc/pb"
)

// Mapper for the `mercury_transmit_requests` table.
type transmitRequest struct {
	Payload      []byte
	ConfigDigest common.Hash
	Epoch        uint32
	Round        uint8
	ExtraHash    common.Hash
}

type ORM struct {
	q pg.Q
}

func NewORM(db *sqlx.DB, lggr logger.Logger, cfg pg.QConfig) *ORM {
	namedLogger := lggr.Named("MercuryORM")
	q := pg.NewQ(db, namedLogger, cfg)
	return &ORM{
		q: q,
	}
}

// InsertTransmitRequest inserts one transmit request if the payload does not exist already.
func (o *ORM) InsertTransmitRequest(req *pb.TransmitRequest, reportCtx ocrtypes.ReportContext, qopts ...pg.QOpt) error {
	q := o.q.WithOpts(qopts...)
	err := q.ExecQ(`
		INSERT INTO mercury_transmit_requests (payload, payload_hash, config_digest, epoch, round, extra_hash)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (payload_hash) DO NOTHING
	`, req.Payload, hashPayload(req.Payload), reportCtx.ConfigDigest[:], reportCtx.Epoch, reportCtx.Round, reportCtx.ExtraHash[:])
	return err
}

// DeleteTransmitRequest deletes one transmit request if it exists.
func (o *ORM) DeleteTransmitRequest(req *pb.TransmitRequest, qopts ...pg.QOpt) error {
	q := o.q.WithOpts(qopts...)
	err := q.ExecQ(`
		DELETE FROM mercury_transmit_requests
		WHERE payload_hash = $1
	`, hashPayload(req.Payload))
	return err
}

// GetTransmitRequests returns all transmit requests in chronologically descending order.
func (o *ORM) GetTransmitRequests(qopts ...pg.QOpt) ([]*Transmission, error) {
	q := o.q.WithOpts(qopts...)
	rows := make([]transmitRequest, 0)
	// The priority queue uses epoch and round to sort transmissions so order by
	// the same fields here for optimal insertion into the pq.
	err := q.Select(&rows, `
		SELECT payload, config_digest, epoch, round, extra_hash
		FROM mercury_transmit_requests
		ORDER BY epoch DESC, round DESC
	`)
	if err != nil {
		return nil, err
	}

	transmissions := make([]*Transmission, len(rows))
	for i, row := range rows {
		transmissions[i] = &Transmission{
			Req: &pb.TransmitRequest{Payload: row.Payload},
			ReportCtx: ocrtypes.ReportContext{
				ReportTimestamp: ocrtypes.ReportTimestamp{
					ConfigDigest: ocrtypes.ConfigDigest(row.ConfigDigest),
					Epoch:        row.Epoch,
					Round:        row.Round,
				},
				ExtraHash: row.ExtraHash,
			},
		}
	}
	return transmissions, nil
}

func hashPayload(payload []byte) []byte {
	checksum := sha256.Sum256(payload)
	return checksum[:]
}
