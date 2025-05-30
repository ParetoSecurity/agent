package check

type CheckState string

const (
	CheckStatePassed   CheckState = "pass"
	CheckStateFailed   CheckState = "fail"
	CheckStateDisabled CheckState = "off"
	CheckStateError    CheckState = "error"
)

type Check interface {
	Name() string
	PassedMessage() string
	FailedMessage() string
	Run() error
	Passed() bool
	IsRunnable() bool
	UUID() string
	Status() string
	RequiresRoot() bool
}
