package beacon

import (
	"bytes"
	"github.com/joernweissenborn/eventual2go"
	"log"
	"net"
	"sync"
	"time"
)

type Beacon struct {
	m          *sync.Mutex
	conf       *Config
	payload    []byte
	stop       chan struct{}
	listenConn *net.UDPConn
	outConns   []*net.UDPConn
	in         *eventual2go.StreamController
	silence    *eventual2go.Completer
	silent     bool
	logger     *log.Logger
}

func New(payload []byte, conf *Config) (b *Beacon) {
	b = new(Beacon)
	b.m = new(sync.Mutex)
	b.payload = payload
	b.conf = conf
	conf.init()
	b.init()
	b.setup()
	return
}

func (b *Beacon) init() {
	b.in = eventual2go.NewStreamController()
	b.stop = make(chan struct{})
	b.outConns = []*net.UDPConn{}
	b.silence = eventual2go.NewCompleter()
	b.silent = true
	b.logger = log.New(b.conf.Logger, "beacon ", log.Lshortfile)
}

func (b *Beacon) setup() {
	b.logger.Println("Setting up")
	b.setupListener()
	b.setupSender()
}

func (b *Beacon) Stop() {
	b.m.Lock()
	defer b.m.Unlock()
	b.logger.Println("Stopping")
	if !b.silent {
		b.silence.Complete(nil)
	}
	close(b.stop)
	b.in.Close()
	b.logger.Println("Stopped")
}

func (b *Beacon) setupSender() {
	for _, addr := range b.conf.PingAddresses {
		b.setupOutgoing(addr)
	}
}

func (b *Beacon) setupOutgoing(addr string) {
	BROADCAST_IPv4 := net.IPv4bcast

	socket, err := net.DialUDP("udp4", &net.UDPAddr{
		IP:   net.ParseIP(addr),
		Port: 0},
		&net.UDPAddr{
			IP:   BROADCAST_IPv4,
			Port: b.conf.Port,
		})
	if err == nil {
		b.outConns = append(b.outConns, socket)
	}

}

func (b *Beacon) setupListener() (err error) {
	var ip net.IP
	ip = net.IPv4(224, 0, 0, 251)

	b.logger.Printf("Listen Address is %s", ip)

	b.listenConn, err = net.ListenMulticastUDP("udp4", nil, &net.UDPAddr{
		IP:   ip,
		Port: b.conf.Port,
	})
	return
}

func (b *Beacon) Run() {
	go b.listen()
}

func (b *Beacon) listen() {

	c := make(chan struct{})
	go b.getSignal(c)
	for {
		select {
		case <-b.stop:
			return

		case <-c:
			go b.getSignal(c)
		}
	}

}

func (b *Beacon) getSignal(c chan struct{}) {
	b.m.Lock()
	defer b.m.Unlock()
	data := make([]byte, 1024)
	read, remoteAddr, _ := b.listenConn.ReadFromUDP(data)
	if !b.in.Closed().Completed() {
		b.in.Add(Signal{remoteAddr.IP[len(remoteAddr.IP)-4:], data[:read]})
		c <- struct{}{}
	}
}

func (b *Beacon) Signals() *eventual2go.Stream {
	return b.in.Where(b.noEcho)
}

func (b *Beacon) Ping() {
	if b.silent {
		b.silent = false
		for _, conn := range b.outConns {
			go b.ping(conn)
		}
	}
}

func (b *Beacon) Silence() {
	b.m.Lock()
	defer b.m.Unlock()
	if !b.silent {
		b.silence.Complete(nil)
		b.silence = eventual2go.NewCompleter()
		b.silent = true
	}
}

func (b *Beacon) Silent() bool {
	return b.silent
}
func (b *Beacon) ping(c *net.UDPConn) {

	t := time.NewTimer(b.conf.PingInterval)
	silence := b.silence.Future().AsChan()
	for {
		select {
		case <-silence:
			b.logger.Println("Silencing")
			return

		case <-t.C:
			c.Write(b.payload)
			t.Reset(b.conf.PingInterval)
		}
	}

}

func (b *Beacon) noEcho(d eventual2go.Data) bool {
	return !bytes.Equal(d.(Signal).Data, b.payload)
}
