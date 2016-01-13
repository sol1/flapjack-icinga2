/*
Copied from https://github.com/flapjack/flapjack/blob/master/src/flapjack/event.go
(to ease local debugging/testing)
*/

package flapjack_icinga2

import (
  "errors"
)

// FlapjackEvent is a basic representation of a Flapjack event.
// Find more at http://flapjack.io/docs/1.0/development/DATA_STRUCTURES
type FlapjackEvent struct {
  Entity  string `json:"entity"`
  Check   string `json:"check"`
  Type    string `json:"type"`
  State   string `json:"state"`
  Summary string `json:"summary"`
  Time    int64  `json:"time"`
}

// IsValid performs basic validations on the event data.
func (e FlapjackEvent) IsValid() error {
  // FIXME: provide validation errors for each failure
  if len(e.Entity) == 0 {
    return errors.New("no entity")
  }
  if len(e.Check) == 0 {
    return errors.New("no check")
  }
  if len(e.State) == 0 {
    return errors.New("no state")
  }
  if len(e.Summary) == 0 {
    return errors.New("no summary")
  }
  return nil
}
