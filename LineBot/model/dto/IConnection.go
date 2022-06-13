package dto

type IConnection interface {
	Connect(url string) (IConnection, error)
	Disconnect()
	Get(userID string) any
	Insert(model MessageModel) (interface{}, error)
}
