package server

import "net"

type ConnectionPool struct {
	list map[int]net.Conn
}

func NewConnectionPool() ConnectionPool {
	pool := ConnectionPool {
		list: make(map[int]net.Conn),
	}

	return pool
}

// add collection to pool
func (pool ConnectionPool) Add(connection net.Conn) int {
	nextConnectionId := len(pool.list)
	pool.list[nextConnectionId] = connection
	return nextConnectionId
}

// remove connection from pool
func (pool ConnectionPool) Remove(connectionId int) {
	delete(pool.list, connectionId)
}

// get size of connections pool
func (pool ConnectionPool) Size() int {
	return len(pool.list)
}

// iterator
func (pool ConnectionPool) Range(callback func(net.Conn, int)) {
	for connectionId, connection := range pool.list {
		callback(connection, connectionId)
	}
}
