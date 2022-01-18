package error

import (
	"fmt"
)

type Message struct {
	Message string
}

func (msg Message) Error() string {
	return fmt.Sprintf("error: %s", msg.Message)
}

var (
	UnexpectedError error = Message{
		Message: "something went wrong",
	}

	DBCreateTxError error = Message{
		Message: "can not create transaction",
	}

	DBRollbackError error = Message{
		Message: "can not rollback",
	}

	DBCommitError error = Message{
		Message: "can not commit",
	}

	DBScanError error = Message{
		Message: "can not scan",
	}

	DBInsertError error = Message{
		Message: "can not insert",
	}

	DBSelectError error = Message{
		Message: "can not select",
	}

	DBUpdateError error = Message{
		Message: "can not update",
	}

	DBNothingUpdated error = Message{
		Message: "nothing updated",
	}

	UInternalError error = Message{
		Message: "internal error",
	}

	UAlreadyExist error = Message{
		Message: "already exist",
	}

	UNotExist error = Message{
		Message: "not exist",
	}

	UNotFound error = Message{
		Message: "not found",
	}

	UUnableToUpdate error = Message{
		Message: "unable to update",
	}

	UUnableToCreate error = Message{
		Message: "unable to create",
	}

	UUnableToGet error = Message{
		Message: "unable to get",
	}

	InternalError error = Message{
		Message: "Internal error",
	}

	NotExist error = Message{
		Message: "Not exist",
	}

	InsertError error = Message{
		Message: "Insert error",
	}

	BadUpdate error = Message{
		Message: "Bad update",
	}

	ConflictError error = Message{
		Message: "Conflict error",
	}
)
