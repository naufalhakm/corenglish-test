package enum

type TaskStatus string

const (
	StatusToDo       TaskStatus = "TO_DO"
	StatusInProgress TaskStatus = "IN_PROGRESS"
	StatusDone       TaskStatus = "DONE"
)

func (s TaskStatus) IsValid() bool {
	return s == StatusToDo || s == StatusInProgress || s == StatusDone
}
