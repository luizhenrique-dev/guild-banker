package importer

import "errors"

var (
	ErrImportBatchNotFound  = errors.New("import batch not found")
	ErrImportItemNotFound   = errors.New("import item not found")
	ErrRequesterIsNotMember = errors.New("requester is not a member of guild")
	ErrInvalidImportStatus  = errors.New("invalid import status")
)
