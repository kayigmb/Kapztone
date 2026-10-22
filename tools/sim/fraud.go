package main

import (
	"context"
	"fmt"
)

// tamper does a lot's database fraud the way an insider with database access
// would, after the lot was sealed, and returns the passport to verify.
func (s *sim) tamper(ctx context.Context, l lot) (string, error) {
	switch l.plan.Fraud {
	case F1: // the quantity on the record no longer matches what was signed
		_, err := s.db.Exec(ctx, `UPDATE passports SET quantity = quantity + quantity / 10 + 1 WHERE id = $1`, l.final)
		return l.final, err
	case F2: // custody moved with no handover: the record names a holder nobody signed for
		// Back to the producer that owns the origin facility. Not the origin
		// passport's holder: a lot never transformed is its own origin, and
		// that "move" would change nothing.
		tag, err := s.db.Exec(ctx,
			`UPDATE passports SET organisation_id = owner.id
			 FROM (SELECT f.organisation_id AS id FROM passports o JOIN facilities f ON f.id = o.facility_id WHERE o.id = $2) owner
			 WHERE passports.id = $1 AND passports.organisation_id <> owner.id`,
			l.final, l.passports[0])
		if err == nil && tag.RowsAffected() != 1 {
			err = fmt.Errorf("lot %s: no other holder to move it to", l.code)
		}
		return l.final, err
	case F5: // a copy of a sealed passport under a new id and code, hash and seal included
		var id string
		err := s.db.QueryRow(ctx,
			`INSERT INTO passports
			 SELECT (jsonb_populate_record(NULL::passports,
			         to_jsonb(p) || jsonb_build_object('id', gen_random_uuid(), 'code', p.code || '-C'))).*
			 FROM passports p WHERE p.id = $1
			 RETURNING id::text`, l.final).Scan(&id)
		return id, err
	case F6: // a certificate's fingerprint swapped, as if a different document had been issued
		if len(l.certs) == 0 {
			return "", fmt.Errorf("lot %s has no certificate to alter", l.code)
		}
		_, err := s.db.Exec(ctx, `UPDATE attestations SET doc_hash = '0x' || repeat('ab', 32) WHERE id = $1`, l.certs[len(l.certs)-1])
		return l.final, err
	}
	return "", fmt.Errorf("unknown database fraud %q", l.plan.Fraud)
}
