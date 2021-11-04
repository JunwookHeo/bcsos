package blockchain

type Transaction struct {
	Id   []byte
	Data []byte
}

func (t *Transaction) Hash() []byte {
	return nil
}

func SerializeTransaction(t Transaction) []byte {
	return Serialize(t)
}

func DeserializeTransaction(d []byte) *Transaction {
	var t Transaction
	Deserialize(d, &t)
	return &t
}

func CreateTransaction() *Transaction {
	return nil
}
