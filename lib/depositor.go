package lib

import (
	"github.com/Sirupsen/logrus"
	gm "github.com/jpoehls/gophermail"
	md "github.com/luksen/maildir"
)

type depositor interface {
	Deposit(*gm.Message) error
}

type maildirDepositor struct {
	dir md.Dir
}

func newMaildirDepositor(dirname string) *maildirDepositor {
	dir := md.Dir(dirname)
	err := dir.Create()
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"maildir": dirname,
			"err":     err.Error(),
		}).Fatal("Problem initializing maildir")
	}
	return &maildirDepositor{dir}
}

func (m *maildirDepositor) Deposit(msg *gm.Message) error {

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
