package bc

func stateChange(tx transaction) error {

	/*
	if _, exists := State[tx.Payload.From]; !exists {
		return errors.New("Sender does not exist in the State.")
	}

	if _, exists := State[tx.Payload.To]; !exists {
		return errors.New("Receiver does not exist in the State.")
	}

	if uint64(tx.Payload.Amount) > State[tx.Payload.From].Balance {
		return errors.New("Sender does not have enough funds for the transaction.")
	}

	accSender := State[tx.Payload.From]
	accSender.TxCnt += 1
	accSender.Balance -= uint64(tx.Payload.Amount)
	State[tx.Payload.From] = accSender

	accReceiver := State[tx.Payload.To]
	accReceiver.Balance += uint64(tx.Payload.Amount)
	State[tx.Payload.To] = accReceiver

	//all good*/
	return nil
}

func stateRollBack(index int, txData []transaction) {

	//in case the first entry failed we don't need to rollback
	/*if index == -1 {
		return
	}

	for cnt := index; cnt >= 0; cnt-- {
		tx := txData[cnt]
		accSender := State[tx.Payload.From]
		accSender.TxCnt -= 1
		accSender.Balance += uint64(tx.Payload.Amount)
		State[tx.Payload.From] = accSender

		accReceiver := State[tx.Payload.To]
		accReceiver.Balance -= uint64(tx.Payload.Amount)
		State[tx.Payload.To] = accReceiver
	}*/
}
