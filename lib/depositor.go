package lib

import gm "github.com/jpoehls/gophermail"
import md "github.com/luksen/maildir"

type depositor interface {
	Deposit(*gm.Message) error
}

type maildir struct {
	dir string
}

func (m *maildir) Deposit(msg *gm.Message) error {
	dir := md.Dir(m.dir)
	dir.Create() // TODO err?

	delivery, err := dir.NewDelivery()
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
