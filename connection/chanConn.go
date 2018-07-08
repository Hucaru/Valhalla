package connection

type ClientChanConn struct {
	ClientConnection
	userID        int32
	isLogedIn     bool
	isAdmin       bool
	worldID       int32
	chanID        int32
	closeCallback []func()
}

// NewChanConnection -
func NewChanConnection(conn ClientConnection) *ClientChanConn {
	loginConn := &ClientChanConn{ClientConnection: conn, isAdmin: false}
	return loginConn
}

func (c *ClientChanConn) Close() {
	if c.isLogedIn {
		for i := range c.closeCallback {
			c.closeCallback[i]()
		}
	}

	c.Conn.Close()
}

func (c *ClientChanConn) String() string {
	return c.ClientConnection.String()
}

func (c *ClientChanConn) SetUserID(val int32) {
	c.userID = val
}

func (c *ClientChanConn) GetUserID() int32 {
	return c.userID
}

func (c *ClientChanConn) SetAdmin(val bool) {
	c.isAdmin = val
}

func (c *ClientChanConn) IsAdmin() bool {
	return c.isAdmin
}

func (c *ClientChanConn) SetIsLoggedIn(val bool) {
	c.isLogedIn = val
}

func (c *ClientChanConn) GetIsLoggedIn() bool {
	return c.isLogedIn
}

func (c *ClientChanConn) SetWorldID(val int32) {
	c.worldID = val
}

func (c *ClientChanConn) GetWorldID() int32 {
	return c.worldID
}

func (c *ClientChanConn) SetChanID(val int32) {
	c.chanID = val
}

func (c *ClientChanConn) GetChanID() int32 {
	return c.chanID
}

func (c *ClientChanConn) AddCloseCallback(callbacK func()) {
	c.closeCallback = append(c.closeCallback, callbacK)
}
