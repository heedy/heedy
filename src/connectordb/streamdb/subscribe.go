package streamdb

/*
import "github.com/apcera/nats"

//SubscribeUser subscribes to the given user
func (o *Database) SubscribeUser(username string, chn chan Message) (*nats.Subscription, error) {
	usr, err := o.ReadUser(username)
	if err!=nil {
		return nil,err
	}
	return o.SubscribeUserByID(usr.UserId,chn)
}
func (o *Database) SubscribeUserByID(userID int64, chn chan Message) (*nats.Subscription, error) {
	return o.msg.Subscribe(getTimebatchUserName(userID)+"/>")
}
func (o *Database) Subscribe(path string) (*nats.Subscription, error) {

}
*/
