package ctimeout

import "fmt"

// Confirm timeout value for the Request method,in seconds. Nested under Params.
type ConfirmTimeout struct {
	Timeout int `json:"confirm-timeout,omitempty"`
}

func (c *ConfirmTimeout) GetTimeout() int {
	return c.Timeout
}

func (c *ConfirmTimeout) SetTimeout(t int) error {
	if t <= 0 {
		return fmt.Errorf("confirm-timeout should be positive integer")
	}
	c.Timeout = t
	return nil
}
