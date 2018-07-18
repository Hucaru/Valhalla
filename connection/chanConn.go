package connection

type Channel struct {
	Client
	userID        int32
	isLogedIn     bool
	isAdmin       bool
	worldID       int32
	chanID        int32
	closeCallback []func()
}

// NewChanConnection -
func NewChannel(conn Client) *Channel {
	loginConn := &Channel{Client: conn, isAdmin: false}
	return loginConn
}

func (c *Channel) Close() {
	if c.isLogedIn {
		for i := range c.closeCallback {
			c.closeCallback[i]()
		}
	}

	c.Conn.Close()
}

func (c *Channel) String() string {
	return c.Client.String()
}

func (c *Channel) SetUserID(val int32) {
	c.userID = val
}

func (c *Channel) GetUserID() int32 {
	return c.userID
}

func (c *Channel) SetAdmin(val bool) {
	c.isAdmin = val
}

func (c *Channel) IsAdmin() bool {
	return c.isAdmin
}

func (c *Channel) SetIsLoggedIn(val bool) {
	c.isLogedIn = val
}

func (c *Channel) GetIsLoggedIn() bool {
	return c.isLogedIn
}

func (c *Channel) SetWorldID(val int32) {
	c.worldID = val
}

func (c *Channel) GetWorldID() int32 {
	return c.worldID
}

func (c *Channel) SetChanID(val int32) {
	c.chanID = val
}

func (c *Channel) GetChanID() int32 {
	return c.chanID
}

func (c *Channel) AddCloseCallback(callbacK func()) {
	c.closeCallback = append(c.closeCallback, callbacK)
}
