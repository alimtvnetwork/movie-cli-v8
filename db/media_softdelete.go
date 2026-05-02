// db/media_softdelete.go — soft-delete helpers introduced by V4.
//
// Spec: spec/08-app/10-remove-move-rescan/remove-command/01-spec.md
//
// SoftDeleteMedia flips IsDeleted=1 and MediaStatusId=Removed for the given
// MediaId. RestoreMedia is the inverse, used by `movie undo`.
// QueryMediaIDsByWhere runs an arbitrary parameterised WHERE clause built by
// cmd.BuildConditionSQL and returns the matching MediaIds (capped).
package db

import "github.com/alimtvnetwork/movie-cli-v8/apperror"

// MaxRemoveMatches caps how many rows a single rm/move can target.
const MaxRemoveMatches = 10000

// SoftDeleteMedia marks the row as Removed without dropping it.
func (d *DB) SoftDeleteMedia(mediaID int64) error {
	_, err := d.Exec(`
		UPDATE Media
		SET IsDeleted = 1,
		    MediaStatusId = ?,
		    UpdatedAt = datetime('now')
		WHERE MediaId = ?`, MediaStatusRemoved, mediaID)
	return err
}

// RestoreMedia reverts a soft delete (used by movie undo).
func (d *DB) RestoreMedia(mediaID int64) error {
	_, err := d.Exec(`
		UPDATE Media
		SET IsDeleted = 0,
		    MediaStatusId = ?,
		    UpdatedAt = datetime('now')
		WHERE MediaId = ?`, MediaStatusActive, mediaID)
	return err
}

// MarkMediaMissing flags a row as on-disk-missing without soft-deleting it.
func (d *DB) MarkMediaMissing(mediaID int64) error {
	_, err := d.Exec(`
		UPDATE Media
		SET MediaStatusId = ?,
		    UpdatedAt = datetime('now')
		WHERE MediaId = ?`, MediaStatusMissing, mediaID)
	return err
}

// QueryMediaIDsByWhere runs SELECT MediaId FROM Media WHERE <where> LIMIT cap.
func (d *DB) QueryMediaIDsByWhere(where string, args []any) ([]int64, error) {
	q := "SELECT MediaId FROM Media WHERE " + where + " LIMIT ?"
	args = append(args, MaxRemoveMatches)
	rows, err := d.Query(q, args...)
	if err != nil {
		return nil, apperror.Wrap("query media ids", err)
	}
	defer rows.Close()
	var ids []int64
	for rows.Next() {
		var id int64
		if scanErr := rows.Scan(&id); scanErr != nil {
			return nil, apperror.Wrap("scan media id", scanErr)
		}
		ids = append(ids, id)
	}
	return ids, nil
}
