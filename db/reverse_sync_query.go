// reverse_sync_query.go — Focused query helpers for reverse-sync.
//
// Reverse sync needs UpdatedAt + IsDeleted, which are not on the
// standard Media struct. Returning a tiny ReverseSyncRow keeps the
// hot loop allocation-free and avoids bloating Media.
//
// Spec: spec/08-app/10-remove-move-rescan/rescan-reconciliation/02-reverse-sync-spec.md
package db

import "github.com/alimtvnetwork/movie-cli-v7/apperror"

// ReverseSyncRow is the minimal projection used by the reverse-sync loop.
type ReverseSyncRow struct {
	UpdatedAt       string
	CurrentFilePath string
	Type            string
	ID              int64
	IsDeleted       bool
}

// ListReverseSyncRows returns every Media row whose OriginalFilePath
// starts with scanDir, including soft-deleted ones (the reverse pass
// must purge sidecars for IsDeleted rows).
func (d *DB) ListReverseSyncRows(scanDir string) ([]ReverseSyncRow, error) {
	prefix := scanDir
	if prefix != "" && prefix[len(prefix)-1] != '/' && prefix[len(prefix)-1] != '\\' {
		prefix += "/"
	}
	rows, err := d.Query(`
		SELECT MediaId,
		       COALESCE(CurrentFilePath, ''),
		       Type,
		       UpdatedAt,
		       COALESCE(IsDeleted, 0)
		FROM   Media
		WHERE  OriginalFilePath LIKE ?`, prefix+"%")
	if err != nil {
		return nil, apperror.Wrap("query reverse-sync rows", err)
	}
	defer rows.Close()

	var out []ReverseSyncRow
	for rows.Next() {
		var r ReverseSyncRow
		var deletedInt int
		if scanErr := rows.Scan(&r.ID, &r.CurrentFilePath, &r.Type, &r.UpdatedAt, &deletedInt); scanErr != nil {
			return nil, apperror.Wrap("scan reverse-sync row", scanErr)
		}
		r.IsDeleted = deletedInt != 0
		out = append(out, r)
	}
	return out, nil
}
