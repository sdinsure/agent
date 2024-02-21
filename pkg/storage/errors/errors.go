package storageerrors

import (
	"errors"

	"github.com/jackc/pgconn"
	pgerrcode "github.com/jackc/pgerrcode"
	sdierrors "github.com/sdinsure/agent/pkg/errors"
	"gorm.io/gorm"
)

var (
	stdNotFound         = errors.New("storage: not found")
	stdStatusConflicted = errors.New("storage: status conflicted")
	stdDuplicatedKey    = errors.New("storage: violate unique constraints")
)

var NotFoundError = sdierrors.New(sdierrors.CodeNotFound, stdNotFound)
var StatusConflictedError = sdierrors.New(sdierrors.CodeStatusConflicted, stdStatusConflicted)
var DuplicatedKeyError = sdierrors.New(sdierrors.CodeBadParameters, stdDuplicatedKey)

func WrapStorageError(err error) *sdierrors.Error {
	if err == nil {
		return nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return NotFoundError
	}
	// internal postgres error
	if e, ok := err.(*pgconn.PgError); ok {
		if e.Code == pgerrcode.UniqueViolation {
			return DuplicatedKeyError
		}
	}
	return sdierrors.New(sdierrors.CodeUnknown, err)
}
