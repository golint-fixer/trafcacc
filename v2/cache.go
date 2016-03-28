package trafcacc

import "sync"

type connCache struct {
	sync.RWMutex
	seqence map[uint32]*packet
	lastack uint32
}

type writeCache struct {
	sync.RWMutex
	conns map[uint64]*connCache
}

func newWriteCache() *writeCache {
	return &writeCache{conns: make(map[uint64]*connCache)}
}

func (c *writeCache) add(p *packet) {
	key := packetKey(p.Senderid, p.Connid)

	c.RLock()
	cn, exist := c.conns[key]
	c.RUnlock()
	if !exist {
		cn = &connCache{
			seqence: make(map[uint32]*packet),
		}
		c.Lock()
		c.conns[key] = cn
		c.Unlock()
	}

	cn.Lock()
	if p.Seqid > cn.lastack {
		if _, ok := cn.seqence[p.Seqid]; !ok {
			cn.seqence[p.Seqid] = p
		}
	}
	cn.Unlock()

	return
}

func (c *writeCache) get(senderid, connid, seqid uint32) *packet {
	key := packetKey(senderid, connid)

	c.RLock()
	cn, exist := c.conns[key]
	c.RUnlock()
	if !exist {
		return nil
	}

	cn.Lock()
	defer cn.Unlock()
	return cn.seqence[seqid]
}

func (c *writeCache) ack(senderid, connid, seqid uint32) {
	key := packetKey(senderid, connid)

	c.RLock()
	cn, exist := c.conns[key]
	c.RUnlock()
	if !exist {
		return
	}

	cn.Lock()
	for k := range cn.seqence {
		if k <= seqid {
			delete(cn.seqence, k)
		} else {
			// maybe performance effictive
			break
		}
	}
	cn.Unlock()
}

func (c *writeCache) close(senderid, connid, seqid uint32) {
	key := packetKey(senderid, connid)

	c.RLock()
	delete(c.conns, key)
	c.RUnlock()
}
