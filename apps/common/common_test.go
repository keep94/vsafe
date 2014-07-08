package common_test

import (
  "github.com/gorilla/sessions"
  "github.com/keep94/vsafe"
  "github.com/keep94/vsafe/apps/common"
  "testing"
)

func TestKey(t *testing.T) {
  userSession := common.CreateUserSession(
      &sessions.Session{Values: make(map[interface{}]interface{})})
  if out := userSession.Key(); out != nil {
    t.Error("Expected nil")
  }
  userSession.SetKey(&vsafe.Key{Id: 17})
  if out := userSession.Key().Id; out != 17 {
    t.Errorf("Expected 17, got %d", out)
  }
  userSession.SetKey(nil)
  if out := userSession.Key(); out != nil {
    t.Error("Expected nil again")
  }
}
  
