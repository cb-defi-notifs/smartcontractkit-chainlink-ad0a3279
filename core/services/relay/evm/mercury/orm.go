package mercury

import (
	"context"
	"crypto/sha256"

	"github.com/ethereum/go-ethereum/common"
	"github.com/smartcontractkit/sqlx"

	ocrtypes "github.com/smartcontractkit/libocr/offchainreporting2plus/types"

	"github.com/smartcontractkit/chainlink/v2/core/logger"
	"github.com/smartcontractkit/chainlink/v2/core/services/pg"
	"github.com/smartcontractkit/chainlink/v2/core/services/relay/evm/mercury/wsrpc/pb"
)

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
func (o *ORM) InsertTransmitRequest(ctx context.Context, req *pb.TransmitRequest, reportCtx ocrtypes.ReportContext, qopts ...pg.QOpt) error {
	q := o.q.WithOpts(append([]pg.QOpt{pg.WithParentCtx(ctx)}, qopts...)...)
	err := q.ExecQ(`
		INSERT INTO mercury_transmit_requests (payload, payload_hash, config_digest, epoch, round, extra_hash)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (payload_hash) DO NOTHING
	`, req.Payload, hashPayload(req.Payload), reportCtx.ConfigDigest[:], reportCtx.Epoch, reportCtx.Round, reportCtx.ExtraHash[:])
	return err
}

// DeleteTransmitRequest deletes one transmit request if it exists.
func (o *ORM) DeleteTransmitRequest(ctx context.Context, req *pb.TransmitRequest, qopts ...pg.QOpt) error {
	q := o.q.WithOpts(append([]pg.QOpt{pg.WithParentCtx(ctx)}, qopts...)...)
	err := q.ExecQ(`
		DELETE FROM mercury_transmit_requests
		WHERE payload_hash = $1
	`, hashPayload(req.Payload))
	return err
}

// GetTransmitRequests returns all transmit requests in chronologically descending order.
func (o *ORM) GetTransmitRequests(ctx context.Context, qopts ...pg.QOpt) ([]*Transmission, error) {
	q := o.q.WithOpts(append([]pg.QOpt{pg.WithParentCtx(ctx)}, qopts...)...)
	// The priority queue uses epoch and round to sort transmissions so order by
	// the same fields here for optimal insertion into the pq.
	rows, err := q.QueryContext(ctx, `
		SELECT payload, config_digest, epoch, round, extra_hash
		FROM mercury_transmit_requests
		ORDER BY epoch DESC, round DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transmissions []*Transmission
	for rows.Next() {
		transmission := &Transmission{Req: &pb.TransmitRequest{}}
		var digest, extraHash common.Hash

		err := rows.Scan(
			&transmission.Req.Payload,
			&digest,
			&transmission.ReportCtx.Epoch,
			&transmission.ReportCtx.Round,
			&extraHash,
		)
		if err != nil {
			return nil, err
		}
		transmission.ReportCtx.ConfigDigest = ocrtypes.ConfigDigest(digest)
		transmission.ReportCtx.ExtraHash = extraHash

		transmissions = append(transmissions, transmission)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return transmissions, nil
}

// PruneTransmitRequests keeps at most maxSize rows in the table, deleting the
// oldest transactions.
func (o *ORM) PruneTransmitRequests(ctx context.Context, maxSize int, qopts ...pg.QOpt) error {
	q := o.q.WithOpts(append([]pg.QOpt{pg.WithParentCtx(ctx)}, qopts...)...)
	// Prune the oldest requests by epoch and round.
	return q.ExecQ(`
		DELETE FROM mercury_transmit_requests
		WHERE payload_hash IN (
		    SELECT payload_hash
			FROM mercury_transmit_requests
			ORDER BY epoch, round
			LIMIT ( 
			    SELECT GREATEST(COUNT(*) - $1, 0)
			    FROM mercury_transmit_requests
			)
		)
	`, maxSize)
}

func hashPayload(payload []byte) []byte {
	checksum := sha256.Sum256(payload)
	return checksum[:]
}
