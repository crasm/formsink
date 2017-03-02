package lib

import (
	"github.com/Sirupsen/logrus"
	"github.com/jpoehls/gophermail"
	"github.com/luksen/maildir"
)

type depositor interface {
	Deposit(*gophermail.Message) error
}

type maildirDepositor struct {
	dir maildir.Dir
}

func newMaildirDepositor(dirname string) *maildirDepositor {
	dir := maildir.Dir(dirname)
	err := dir.Create()
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"maildir": dirname,
			"err":     err.Error(),
		}).Fatal("Problem initializing maildir")
	}
	return &maildirDepositor{dir}
}

func (m *maildirDepositor) Deposit(msg *gophermail.Message) error {

	delivery, err := m.dir.NewDelivery()
	if err != nil {
		return err
	}
	defer delivery.Close()

	msgBytes, err := msg.Bytes()
	if err != nil {
		return err
	}

	_, err = delivery.Write(msgBytes)
	return err
}
